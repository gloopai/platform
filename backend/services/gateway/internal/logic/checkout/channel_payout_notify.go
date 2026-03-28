package logic

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/orderclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/gateway/internal/requestx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// contracts.PayoutOrderStatus
const (
	payoutNotifyStatusProcessing = 1
	payoutNotifyStatusSuccess    = 2
	payoutNotifyStatusFailed     = 3
)

// ChannelPayoutNotify handles channel (PSP) JSON async payout notify. Response body is SUCCESS/FAIL plain text.
func (c *Checkout) ChannelPayoutNotify(w http.ResponseWriter, r *http.Request) {
	reqID := requestx.FromContext(c.ctx)
	channelID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("channel_id")), 10, 64)
	orderNo := strings.TrimSpace(r.URL.Query().Get("order_no"))
	if channelID <= 0 || orderNo == "" {
		c.Errorf("request_id=%s action=channel_payout_notify bad_query channel_id=%d order_no=%q", reqID, channelID, orderNo)
		http.Error(w, "bad query", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	vr, err := c.svcCtx.ChannelRpc.ChannelVerifyPayoutNotify(c.ctx, &channelpb.ChannelVerifyPayoutNotifyReq{
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
		c.Errorf("request_id=%s action=channel_payout_notify verify_rpc err=%v", reqID, err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	writeFail := func() {
		br, berr := c.svcCtx.ChannelRpc.ChannelBuildPayoutNotifyResponse(c.ctx, &channelpb.ChannelBuildPayoutNotifyResponseReq{
			ChannelId: channelID,
			Success:   false,
		})
		if berr == nil && br != nil {
			w.Header().Set("Content-Type", br.GetContentType())
			_, _ = w.Write(br.GetBody())
		}
	}
	writeOk := func() {
		br, berr := c.svcCtx.ChannelRpc.ChannelBuildPayoutNotifyResponse(c.ctx, &channelpb.ChannelBuildPayoutNotifyResponseReq{
			ChannelId: channelID,
			Success:   true,
		})
		if berr == nil && br != nil {
			w.Header().Set("Content-Type", br.GetContentType())
			_, _ = w.Write(br.GetBody())
		}
	}

	if !vr.GetVerifyOk() {
		w.Header().Set("Content-Type", vr.GetResponseContentType())
		_, _ = w.Write(vr.GetResponseBody())
		return
	}

	switch vr.GetPayoutStatus() {
	case payoutNotifyStatusProcessing:
		writeOk()
		return
	case payoutNotifyStatusSuccess:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutSuccess(c.ctx, orderNo, vr.GetChannelOrderNo())
		if merr != nil {
			c.Errorf("request_id=%s action=channel_payout_notify mark_success err=%v", reqID, merr)
			writeFail()
			return
		}
		if changed {
			writeOk()
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 1) {
			writeOk()
			return
		}
		writeFail()
		return
	case payoutNotifyStatusFailed:
		changed, merr := c.svcCtx.ServiceHub.MarkPayoutFailed(c.ctx, orderNo)
		if merr != nil {
			c.Errorf("request_id=%s action=channel_payout_notify mark_failed err=%v", reqID, merr)
			writeFail()
			return
		}
		if changed {
			writeOk()
			return
		}
		if c.payoutOrderTerminalStatus(orderNo, 2) {
			writeOk()
			return
		}
		writeFail()
		return
	default:
		writeFail()
	}
}

func (c *Checkout) payoutOrderTerminalStatus(orderNo string, want int32) bool {
	r, err := c.svcCtx.OrderRpc.GetPayoutOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: orderNo})
	if err != nil || r.GetOrder() == nil {
		return false
	}
	return r.GetOrder().GetStatus() == want
}
