package logic

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gloopai/pay/channeldriver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	gch, err := c.svcCtx.ChannelRpc.GetChannel(c.ctx, &channelpb.GetChannelReq{ChannelId: channelID})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
		c.Errorf("request_id=%s action=upstream_payout get_channel err=%v", reqID, err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	chRow := gch.GetChannel()
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
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	}
	if strings.TrimSpace(parsed.MerchantOrderNo) != orderNo {
		c.Errorf("request_id=%s action=upstream_payout order_mismatch query=%s body=%s", reqID, orderNo, parsed.MerchantOrderNo)
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	}

	switch parsed.Status {
	case channeldriver.PayoutStatusProcessing:
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(true))
		return
	case channeldriver.PayoutStatusSuccess:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutSuccess(c.ctx, orderNo, parsed.ChannelOrderNo)
		if merr != nil {
			c.Errorf("request_id=%s action=upstream_payout mark_success err=%v", reqID, merr)
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
			return
		}
		if changed {
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 1) {
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	case channeldriver.PayoutStatusFailed:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutFailed(c.ctx, orderNo)
		if merr != nil {
			c.Errorf("request_id=%s action=upstream_payout mark_failed err=%v", reqID, merr)
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
			return
		}
		if changed {
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 2) {
			channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(true))
			return
		}
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
		return
	default:
		channeldriver.WriteChannelNotify(w, drv, drv.PayoutNotifyResponse(false))
	}
}

func (c *Checkout) payoutOrderTerminalStatus(orderNo string, want int32) bool {
	r, err := c.svcCtx.OrderRpc.GetPayoutOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: orderNo})
	if err != nil || r.GetOrder() == nil {
		return false
	}
	return r.GetOrder().GetStatus() == want
}
