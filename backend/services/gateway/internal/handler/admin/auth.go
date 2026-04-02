// 管理台登录 / 登出
package handler

import (
	"net/http"

	"github.com/gloopai/platform/gateway/internal/apiresp"
	"github.com/gloopai/platform/gateway/internal/logic"
	"github.com/gloopai/platform/common/requestx"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		var req types.AdminLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminAuth(r.Context(), svcCtx)
		resp, err := l.AdminLogin(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = requestx.Ensure(r, w)
		l := logic.NewAdminAuth(r.Context(), svcCtx)
		resp, err := l.AdminLogout(r.Header.Get("X-Admin-Token"))
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
