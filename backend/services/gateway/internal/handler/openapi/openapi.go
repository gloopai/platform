// 商户签名 OpenAPI：代收/代付下单与查询、余额查询（挂 OpenAPIServer）
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

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.CreateOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.CreateOrder(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func QueryOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.QueryOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryOrder(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func CreatePayoutOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.CreatePayinOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.CreatePayoutOrder(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func QueryPayoutOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.QueryOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}

		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryPayoutOrder(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func QueryMerchantBalanceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.MerchantBalanceQueryReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewCheckout(r.Context(), svcCtx)
		resp, err := l.QueryMerchantBalance(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
