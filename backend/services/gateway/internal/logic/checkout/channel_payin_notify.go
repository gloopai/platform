package logic

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/orderclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/gateway/internal/requestx"
	"github.com/gloopai/pay/gateway/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// contracts.PayinOrderStatus: success = 2
const payinNotifyStatusSuccess = 2

func headerStringMap(r *http.Request) map[string]string {
	if r == nil {
		return nil
	}
	m := make(map[string]string, len(r.Header))
	for k, vv := range r.Header {
		if len(vv) > 0 {
			m[k] = vv[0]
		}
	}
	return m
}

// ChannelPayinNotify handles channel (PSP) JSON async notify. Response body is SUCCESS/FAIL plain text.
func (c *Checkout) ChannelPayinNotify(w http.ResponseWriter, r *http.Request) {
	reqID := requestx.FromContext(c.ctx)
	channelID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("channel_id")), 10, 64)
	orderNo := strings.TrimSpace(r.URL.Query().Get("order_no"))
	if channelID <= 0 || orderNo == "" {
		c.Errorf("request_id=%s action=channel_payin_notify bad_query channel_id=%d order_no=%q", reqID, channelID, orderNo)
		http.Error(w, "bad query", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	vr, err := c.svcCtx.ChannelRpc.ChannelVerifyPayinNotify(c.ctx, &channelpb.ChannelVerifyPayinNotifyReq{
		ChannelId:       channelID,
		ExpectedOrderNo: orderNo,
		Method:          r.Method,
		Header:          headerStringMap(r),
		Body:            body,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
		c.Errorf("request_id=%s action=channel_payin_notify verify_rpc err=%v", reqID, err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !vr.GetVerifyOk() {
		w.Header().Set("Content-Type", vr.GetResponseContentType())
		_, _ = w.Write(vr.GetResponseBody())
		return
	}
	if vr.GetPayinStatus() != payinNotifyStatusSuccess {
		br, berr := c.svcCtx.ChannelRpc.ChannelBuildPayinNotifyResponse(c.ctx, &channelpb.ChannelBuildPayinNotifyResponseReq{
			ChannelId: channelID,
			Success:   false,
		})
		if berr == nil && br != nil {
			w.Header().Set("Content-Type", br.GetContentType())
			_, _ = w.Write(br.GetBody())
		}
		return
	}

	reqN := &types.ChannelNotifyReq{
		OrderNo:        orderNo,
		PaidAmount:     vr.GetPaidAmountMinor(),
		ChannelTradeNo: vr.GetChannelOrderNo(),
		ChannelId:      channelID,
		Sign:           "",
	}
	resp, _ := c.channelNotifyCore(reqID, reqN)
	ok := resp != nil && resp.Ok
	br, berr := c.svcCtx.ChannelRpc.ChannelBuildPayinNotifyResponse(c.ctx, &channelpb.ChannelBuildPayinNotifyResponseReq{
		ChannelId: channelID,
		Success:   ok,
	})
	if berr != nil || br == nil {
		c.Errorf("request_id=%s action=channel_payin_notify build_response err=%v", reqID, berr)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", br.GetContentType())
	_, _ = w.Write(br.GetBody())
}

// channelNotifyCore is shared mark paid + credit + nsq (no sign check).
func (c *Checkout) channelNotifyCore(reqID string, req *types.ChannelNotifyReq) (*types.ChannelNotifyResp, error) {
	if strings.TrimSpace(req.OrderNo) == "" || strings.TrimSpace(req.ChannelTradeNo) == "" || req.ChannelId <= 0 || req.PaidAmount <= 0 {
		return notifyFail(NotifyCodeInvalidNotifyParams, "invalid notify params"), nil
	}

	getResp, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return notifyFail(NotifyCodeOrderNotFound, "order not found"), nil
	}
	o := getResp.GetOrder()
	c.Infof("request_id=%s action=channel_notify_core order_no=%s merchant_id=%s paid_amount=%d channel_id=%d", reqID, req.OrderNo, o.GetMerchantId(), req.PaidAmount, req.ChannelId)

	if o.GetChannelId() != req.ChannelId {
		c.Errorf("request_id=%s action=channel_notify_core channel_mismatch order_no=%s order_ch=%d notify_ch=%d", reqID, req.OrderNo, o.GetChannelId(), req.ChannelId)
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
		OrderNo:        req.OrderNo,
		PaidAmount:     req.PaidAmount,
		ChannelTradeNo: req.ChannelTradeNo,
		ChannelId:      req.ChannelId,
	})
	if err != nil {
		c.Errorf("request_id=%s action=channel_notify_core mark_paid_failed order_no=%s err=%v", reqID, req.OrderNo, err)
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
