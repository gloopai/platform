// 开放接口：下单、查单、上游回调、收银台终端
package handler

import (
	"net/http"

	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/openapi"
	"github.com/gloopai/pay/gateway/internal/requestx"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.CreateOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.CreateOrder(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func QueryOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.QueryOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryOrder(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func CreatePayoutOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.CreatePayinOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.CreatePayoutOrder(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func QueryPayoutOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.QueryOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryPayoutOrder(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func QueryMerchantBalanceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.MerchantBalanceQueryReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}
		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryMerchantBalance(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UpstreamNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.UpstreamNotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.UpstreamNotify(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func TerminalOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.TerminalOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.TerminalOrder(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func TerminalPayHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.TerminalPayReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}
		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.TerminalPay(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
