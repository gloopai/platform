package handler

import (
	"net/http"

	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminLogoutLogic(r.Context(), svcCtx)
		resp, err := l.AdminLogout(r.Header.Get("X-Admin-Token"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
