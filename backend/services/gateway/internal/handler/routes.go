// scaffold/platform-admin：仅注册公共探活与管理端平台接口（RBAC、后台用户、展示配置等）。

package handler

import (
	"net/http"

	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/gateway/pkg/adminshell"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterCommonHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: HealthHandler(),
			},
		},
	)
	_ = serverCtx
}

func RegisterAdminHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	adminshell.RegisterServiceHubShell(server, adminshell.Middlewares{
		OpenAPIParamsParse: serverCtx.OpenAPIParamsParse,
		LoginRateLimit:     serverCtx.LoginRateLimit,
		AdminAuth:          serverCtx.AdminAuth,
		AdminRBAC:          serverCtx.AdminRBAC,
		AdminOpLog:         serverCtx.AdminOpLog,
	}, serverCtx)
}
