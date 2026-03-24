// 管理台登录 / 登出
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

func AdminLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.AdminLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
			return
		}
		l := logic.NewAdminAuth(r.Context(), svcCtx)
		resp, err := l.AdminLogin(&req)
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func AdminLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		l := logic.NewAdminAuth(r.Context(), svcCtx)
		resp, err := l.AdminLogout(r.Header.Get("X-Admin-Token"))
		if err != nil {
			openapi.WriteFromErr(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
