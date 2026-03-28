package logic

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/gateway/internal/requestx"
)

// UpstreamPayoutNotify handles mock_psp-style JSON async payout notify. Response body is SUCCESS/FAIL plain text.
func (c *Checkout) UpstreamPayoutNotify(w http.ResponseWriter, r *http.Request) {
	reqID := requestx.FromContext(c.ctx)
	channelID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("channel_id")), 10, 64)
	orderNo := strings.TrimSpace(r.URL.Query().Get("order_no"))
	if channelID <= 0 || orderNo == "" {
		c.Errorf("request_id=%s action=upstream_payout bad_query channel_id=%d order_no=%q", reqID, channelID, orderNo)
		http.Error(w, "bad query", http.StatusBadRequest)
		return
	}

	list, err := c.svcCtx.ChannelRpc.ListChannels(c.ctx, &channelpb.ListChannelsReq{})
	if err != nil {
		c.Errorf("request_id=%s action=upstream_payout list_channels err=%v", reqID, err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	var chRow *channelpb.ChannelRow
	for _, row := range list.GetChannels() {
		if row.GetId() == channelID {
			chRow = row
			break
		}
	}
	if chRow == nil {
		c.Errorf("request_id=%s action=upstream_payout channel_not_found id=%d", reqID, channelID)
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}

	cfg := channelRowToConfig(chRow)
	drv, err := c.svcCtx.ChannelDrivers.Payout(cfg.DriverKey)
	if err != nil {
		c.Errorf("request_id=%s action=upstream_payout no_driver key=%s err=%v", reqID, cfg.DriverKey, err)
		http.Error(w, "no driver", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	r2 := r.Clone(c.ctx)
	r2.Body = io.NopCloser(bytes.NewReader(body))

	parsed, err := drv.VerifyPayoutNotify(c.ctx, cfg, r2)
	if err != nil {
		c.Infof("request_id=%s action=upstream_payout verify_fail err=%v", reqID, err)
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	}
	if strings.TrimSpace(parsed.MerchantOrderNo) != orderNo {
		c.Errorf("request_id=%s action=upstream_payout order_mismatch query=%s body=%s", reqID, orderNo, parsed.MerchantOrderNo)
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	}

	switch parsed.Status {
	case channeldriver.PayoutStatusProcessing:
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(true))
		return
	case channeldriver.PayoutStatusSuccess:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutSuccess(c.ctx, orderNo, parsed.UpstreamOrderNo)
		if merr != nil {
			c.Errorf("request_id=%s action=upstream_payout mark_success err=%v", reqID, merr)
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
			return
		}
		if changed {
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 1) {
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	case channeldriver.PayoutStatusFailed:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutFailed(c.ctx, orderNo)
		if merr != nil {
			c.Errorf("request_id=%s action=upstream_payout mark_failed err=%v", reqID, merr)
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
			return
		}
		if changed {
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 2) {
			channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	default:
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayoutNotifyResponse(false))
	}
}

func (c *Checkout) payoutOrderTerminalStatus(orderNo string, want int32) bool {
	r, err := c.svcCtx.OrderRpc.GetPayoutOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: orderNo})
	if err != nil || r.GetOrder() == nil {
		return false
	}
	return r.GetOrder().GetStatus() == want
}
