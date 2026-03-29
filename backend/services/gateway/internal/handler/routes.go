// scaffold/platform-admin：仅注册公共探活与管理端平台接口（RBAC、后台用户、展示配置等）。

package handler

import (
	"net/http"

	adminhandler "github.com/gloopai/pay/gateway/internal/handler/admin"
	"github.com/gloopai/pay/gateway/internal/svc"

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
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.OpenAPIParamsParseMiddleware, serverCtx.LoginRateLimitMiddleware},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/login",
					Handler: adminhandler.AdminLoginHandler(serverCtx),
				},
			}...,
		),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.AdminAuthMiddleware, serverCtx.AdminRBACMiddleware},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/logout",
					Handler: adminhandler.AdminLogoutHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/me",
					Handler: adminhandler.AdminMeHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/my_menu",
					Handler: adminhandler.AdminMyMenuHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/roles",
					Handler: adminhandler.AdminListRbacRolesHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/rbac/roles",
					Handler: adminhandler.AdminCreateRbacRoleHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/roles/:id",
					Handler: adminhandler.AdminUpdateRbacRoleHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/v1/admin/rbac/roles/:id",
					Handler: adminhandler.AdminDeleteRbacRoleHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/menus",
					Handler: adminhandler.AdminListRbacMenusHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/rbac/menus",
					Handler: adminhandler.AdminCreateRbacMenuHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/menus/:id",
					Handler: adminhandler.AdminUpdateRbacMenuHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/v1/admin/rbac/menus/:id",
					Handler: adminhandler.AdminDeleteRbacMenuHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/roles/:id/menus",
					Handler: adminhandler.AdminGetRbacRoleMenusHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/roles/:id/menus",
					Handler: adminhandler.AdminSetRbacRoleMenusHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/admin_users/:id/roles",
					Handler: adminhandler.AdminGetRbacUserRolesHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/admin_users/:id/roles",
					Handler: adminhandler.AdminSetRbacUserRolesHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/permissions",
					Handler: adminhandler.AdminListRbacPermissionsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/rbac/permissions",
					Handler: adminhandler.AdminCreateRbacPermissionHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/permissions/:id",
					Handler: adminhandler.AdminUpdateRbacPermissionHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/v1/admin/rbac/permissions/:id",
					Handler: adminhandler.AdminDeleteRbacPermissionHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/roles/:id/perm_keys",
					Handler: adminhandler.AdminGetRbacRolePermKeysHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/rbac/roles/:id/perm_keys",
					Handler: adminhandler.AdminSetRbacRolePermKeysHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/rbac/api_rules",
					Handler: adminhandler.AdminListRbacApiRulesHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/rbac/api_rules",
					Handler: adminhandler.AdminUpsertRbacApiRuleHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/v1/admin/rbac/api_rules/:id",
					Handler: adminhandler.AdminDeleteRbacApiRuleHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/admin_users",
					Handler: adminhandler.AdminListUsersHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/admin_users",
					Handler: adminhandler.AdminCreateUserHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/admin_users/:id",
					Handler: adminhandler.AdminUpdateUserHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/v1/admin/admin_users/:id",
					Handler: adminhandler.AdminDeleteUserHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/admin_users/:id/reset_password",
					Handler: adminhandler.AdminResetUserPasswordHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/admin_users/:id/mfa/setup",
					Handler: adminhandler.AdminMfaSetupHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/admin_users/:id/mfa/confirm",
					Handler: adminhandler.AdminMfaConfirmHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/v1/admin/admin_users/:id/mfa/disable",
					Handler: adminhandler.AdminMfaDisableHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/display_settings",
					Handler: adminhandler.AdminDisplaySettingsHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/v1/admin/display_settings",
					Handler: adminhandler.AdminUpdateDisplaySettingsHandler(serverCtx),
				},
			}...,
		),
	)
}
