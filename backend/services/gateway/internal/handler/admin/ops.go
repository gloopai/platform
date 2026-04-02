// 管理台：运维监控
package handler

import (
	"net/http"

	"github.com/gloopai/platform/gateway/internal/apiresp"
	"github.com/gloopai/platform/gateway/internal/logic"
	"github.com/gloopai/platform/gateway/internal/svc"
)

func AdminOpsServicesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminOps(r.Context(), svcCtx)
		resp, err := l.ServicesStatus()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
