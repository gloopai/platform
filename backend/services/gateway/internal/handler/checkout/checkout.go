// 收银台终端与上游回调（挂 AdminServer / CheckoutServer）
package handler

import (
	"net/http"

	"github.com/gloopai/pay/gateway/internal/apiresp"
	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/requestx"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ChannelNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.ChannelNotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.ChannelNotify(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func UpstreamPayinNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		l := logic.NewCheckout(r.Context(), svcCtx)
		l.UpstreamPayinNotify(w, r)
	}
}

func UpstreamPayoutNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		l := logic.NewCheckout(r.Context(), svcCtx)
		l.UpstreamPayoutNotify(w, r)
	}
}

func TerminalOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.TerminalOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.TerminalOrder(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func TerminalPayHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.TerminalPayReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.TerminalPay(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
