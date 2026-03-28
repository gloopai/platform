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
	"github.com/gloopai/pay/gateway/internal/types"
)

func channelRowToConfig(row *channelpb.ChannelRow) *channeldriver.ChannelConfig {
	if row == nil {
		return nil
	}
	gw := strings.TrimSpace(row.GetGatewayUrl())
	mer := strings.TrimSpace(row.GetUpstreamMerchantNo())
	sig := row.GetSignSecret()
	rsa := row.GetRsaPrivateKey()
	if uc := strings.TrimSpace(row.GetUpstreamConfig()); uc != "" {
		jg, jm, js, jr := channeldriver.ConfigFieldsFromUpstreamJSON(uc)
		if jg != "" {
			gw = jg
		}
		if jm != "" {
			mer = jm
		}
		if js != "" {
			sig = js
		}
		if jr != "" {
			rsa = jr
		}
	}
	return channeldriver.ConfigFromDriverKey(
		row.GetId(),
		strings.TrimSpace(row.GetPayinType()),
		gw,
		mer,
		sig,
		rsa,
		row.GetSupportsPayin(),
		row.GetSupportsPayout(),
	)
}

// UpstreamPayinNotify handles PSP-style JSON async notify (e.g. mock_psp). Response body is SUCCESS/FAIL plain text.
func (c *Checkout) UpstreamPayinNotify(w http.ResponseWriter, r *http.Request) {
	reqID := requestx.FromContext(c.ctx)
	channelID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("channel_id")), 10, 64)
	orderNo := strings.TrimSpace(r.URL.Query().Get("order_no"))
	if channelID <= 0 || orderNo == "" {
		c.Errorf("request_id=%s action=upstream_payin bad_query channel_id=%d order_no=%q", reqID, channelID, orderNo)
		http.Error(w, "bad query", http.StatusBadRequest)
		return
	}

	gch, err := c.svcCtx.ChannelRpc.GetChannel(c.ctx, &channelpb.GetChannelReq{ChannelId: channelID})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
		c.Errorf("request_id=%s action=upstream_payin get_channel err=%v", reqID, err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	chRow := gch.GetChannel()
	if chRow == nil {
		c.Errorf("request_id=%s action=upstream_payin channel_not_found id=%d", reqID, channelID)
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}

	cfg := channelRowToConfig(chRow)
	drv, err := c.svcCtx.ChannelDrivers.Payin(cfg.DriverKey)
	if err != nil {
		c.Errorf("request_id=%s action=upstream_payin no_driver key=%s err=%v", reqID, cfg.DriverKey, err)
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

	parsed, err := drv.VerifyPayinNotify(c.ctx, cfg, r2)
	if err != nil {
		c.Infof("request_id=%s action=upstream_payin verify_fail err=%v", reqID, err)
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayinNotifyResponse(false))
		return
	}
	if strings.TrimSpace(parsed.MerchantOrderNo) != orderNo {
		c.Errorf("request_id=%s action=upstream_payin order_mismatch query=%s body=%s", reqID, orderNo, parsed.MerchantOrderNo)
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayinNotifyResponse(false))
		return
	}
	if parsed.Status != channeldriver.PayinStatusSuccess {
		channeldriver.WriteUpstreamNotify(w, drv, drv.PayinNotifyResponse(false))
		return
	}

	req := &types.UpstreamNotifyReq{
		OrderNo:         orderNo,
		PaidAmount:      parsed.PaidAmountMinor,
		UpstreamTradeNo: parsed.UpstreamOrderNo,
		ChannelId:       channelID,
		Sign:            "",
	}
	resp, _ := c.upstreamNotifyCore(reqID, req)
	ok := resp != nil && resp.Ok
	channeldriver.WriteUpstreamNotify(w, drv, drv.PayinNotifyResponse(ok))
}

// upstreamNotifyCore is shared mark paid + credit + nsq (no sign check).
func (c *Checkout) upstreamNotifyCore(reqID string, req *types.UpstreamNotifyReq) (*types.UpstreamNotifyResp, error) {
	if strings.TrimSpace(req.OrderNo) == "" || strings.TrimSpace(req.UpstreamTradeNo) == "" || req.ChannelId <= 0 || req.PaidAmount <= 0 {
		return notifyFail(NotifyCodeInvalidNotifyParams, "invalid notify params"), nil
	}

	getResp, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return notifyFail(NotifyCodeOrderNotFound, "order not found"), nil
	}
	o := getResp.GetOrder()
	c.Infof("request_id=%s action=upstream_notify_core order_no=%s merchant_id=%s paid_amount=%d channel_id=%d", reqID, req.OrderNo, o.GetMerchantId(), req.PaidAmount, req.ChannelId)

	if o.GetChannelId() != req.ChannelId {
		c.Errorf("request_id=%s action=upstream_notify_core channel_mismatch order_no=%s order_ch=%d notify_ch=%d", reqID, req.OrderNo, o.GetChannelId(), req.ChannelId)
		return notifyFail(NotifyCodeChannelMismatch, "notify channel does not match order channel"), nil
	}

	if o.GetStatus() == 1 {
		if samePaidSnapshot(o, req) {
			return c.settlePaidOrderAndNotify(reqID, o, req, NotifyCodeIdempotentReplayAccepted, "idempotent replay accepted")
		}
		return notifyFail(NotifyCodeReplayPayloadMismatch, "replay payload mismatch"), nil
	}
	if o.GetStatus() != 0 {
		return notifyFail(NotifyCodeOrderNotPending, "order not pending"), nil
	}

	markResp, err := c.svcCtx.OrderRpc.MarkPaid(c.ctx, &orderclient.MarkPaidReq{
		OrderNo:         req.OrderNo,
		PaidAmount:      req.PaidAmount,
		UpstreamTradeNo: req.UpstreamTradeNo,
		ChannelId:       req.ChannelId,
	})
	if err != nil {
		c.Errorf("request_id=%s action=upstream_notify_core mark_paid_failed order_no=%s err=%v", reqID, req.OrderNo, err)
		return notifyFail(NotifyCodeMarkPaidFailed, "mark paid failed"), nil
	}

	if !markResp.GetChanged() {
		latest, ge := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: req.OrderNo})
		if ge != nil {
			return notifyFail(NotifyCodeMarkPaidRace, "mark paid race"), nil
		}
		if samePaidSnapshot(latest.GetOrder(), req) {
			return c.settlePaidOrderAndNotify(reqID, latest.GetOrder(), req, NotifyCodeIdempotentRaceAccepted, "idempotent race accepted")
		}
		return notifyFail(NotifyCodeMarkPaidRaceMismatch, "mark paid race mismatch"), nil
	}

	return c.settlePaidOrderAndNotify(reqID, o, req, "", "")
}
