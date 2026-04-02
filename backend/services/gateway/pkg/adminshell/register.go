package adminshell

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// Middlewares are the go-zero middlewares applied to admin shell groups (same as platform gateway scaffold).
type Middlewares struct {
	OpenAPIParamsParse rest.Middleware
	LoginRateLimit     rest.Middleware
	AdminAuth          rest.Middleware
	AdminRBAC          rest.Middleware
	AdminOpLog         rest.Middleware
}

// Handlers holds HTTP handlers for each shell route; all must be non-nil (go-zero [rest.Route] uses [http.HandlerFunc]).
type Handlers struct {
	Login                 http.HandlerFunc
	Logout                http.HandlerFunc
	Me                    http.HandlerFunc
	MyMenu                http.HandlerFunc
	ListRbacRoles         http.HandlerFunc
	CreateRbacRole        http.HandlerFunc
	UpdateRbacRole        http.HandlerFunc
	DeleteRbacRole        http.HandlerFunc
	ListRbacMenus         http.HandlerFunc
	CreateRbacMenu        http.HandlerFunc
	UpdateRbacMenu        http.HandlerFunc
	DeleteRbacMenu        http.HandlerFunc
	GetRbacRoleMenus      http.HandlerFunc
	SetRbacRoleMenus      http.HandlerFunc
	GetRbacUserRoles      http.HandlerFunc
	SetRbacUserRoles      http.HandlerFunc
	ListRbacPermissions   http.HandlerFunc
	CreateRbacPermission  http.HandlerFunc
	UpdateRbacPermission  http.HandlerFunc
	DeleteRbacPermission  http.HandlerFunc
	GetRbacRolePermKeys   http.HandlerFunc
	SetRbacRolePermKeys   http.HandlerFunc
	ListRbacApiRules      http.HandlerFunc
	UpsertRbacApiRule     http.HandlerFunc
	DeleteRbacApiRule     http.HandlerFunc
	ListUsers             http.HandlerFunc
	CreateUser            http.HandlerFunc
	UpdateUser            http.HandlerFunc
	DeleteUser            http.HandlerFunc
	ResetUserPassword     http.HandlerFunc
	MfaSetup              http.HandlerFunc
	MfaConfirm            http.HandlerFunc
	MfaDisable            http.HandlerFunc
	DisplaySettings       http.HandlerFunc
	OperationLogs         http.HandlerFunc
	UpdateDisplaySettings http.HandlerFunc
	ListScheduledJobKeys  http.HandlerFunc
	ListScheduledJobs     http.HandlerFunc
	CreateScheduledJob    http.HandlerFunc
	UpdateScheduledJob    http.HandlerFunc
	ToggleScheduledJob    http.HandlerFunc
	RunScheduledJob       http.HandlerFunc
	ListJobWorkerNodes    http.HandlerFunc
	ListScheduledJobRuns  http.HandlerFunc
	GetScheduledJobRun    http.HandlerFunc
	RetryScheduledJobRun  http.HandlerFunc
}

// Register adds the admin shell routes to server (same paths/methods as platform gateway RegisterAdminHandlers).
func Register(server *rest.Server, mw Middlewares, h Handlers) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{mw.OpenAPIParamsParse, mw.LoginRateLimit},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/v1/admin/login", Handler: h.Login},
			}...,
		),
	)
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{mw.AdminAuth, mw.AdminRBAC, mw.AdminOpLog},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/v1/admin/logout", Handler: h.Logout},
				{Method: http.MethodGet, Path: "/v1/admin/me", Handler: h.Me},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/my_menu", Handler: h.MyMenu},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/roles", Handler: h.ListRbacRoles},
				{Method: http.MethodPost, Path: "/v1/admin/rbac/roles", Handler: h.CreateRbacRole},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/roles/:id", Handler: h.UpdateRbacRole},
				{Method: http.MethodDelete, Path: "/v1/admin/rbac/roles/:id", Handler: h.DeleteRbacRole},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/menus", Handler: h.ListRbacMenus},
				{Method: http.MethodPost, Path: "/v1/admin/rbac/menus", Handler: h.CreateRbacMenu},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/menus/:id", Handler: h.UpdateRbacMenu},
				{Method: http.MethodDelete, Path: "/v1/admin/rbac/menus/:id", Handler: h.DeleteRbacMenu},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/roles/:id/menus", Handler: h.GetRbacRoleMenus},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/roles/:id/menus", Handler: h.SetRbacRoleMenus},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/admin_users/:id/roles", Handler: h.GetRbacUserRoles},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/admin_users/:id/roles", Handler: h.SetRbacUserRoles},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/permissions", Handler: h.ListRbacPermissions},
				{Method: http.MethodPost, Path: "/v1/admin/rbac/permissions", Handler: h.CreateRbacPermission},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/permissions/:id", Handler: h.UpdateRbacPermission},
				{Method: http.MethodDelete, Path: "/v1/admin/rbac/permissions/:id", Handler: h.DeleteRbacPermission},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/roles/:id/perm_keys", Handler: h.GetRbacRolePermKeys},
				{Method: http.MethodPut, Path: "/v1/admin/rbac/roles/:id/perm_keys", Handler: h.SetRbacRolePermKeys},
				{Method: http.MethodGet, Path: "/v1/admin/rbac/api_rules", Handler: h.ListRbacApiRules},
				{Method: http.MethodPost, Path: "/v1/admin/rbac/api_rules", Handler: h.UpsertRbacApiRule},
				{Method: http.MethodDelete, Path: "/v1/admin/rbac/api_rules/:id", Handler: h.DeleteRbacApiRule},
				{Method: http.MethodGet, Path: "/v1/admin/admin_users", Handler: h.ListUsers},
				{Method: http.MethodPost, Path: "/v1/admin/admin_users", Handler: h.CreateUser},
				{Method: http.MethodPut, Path: "/v1/admin/admin_users/:id", Handler: h.UpdateUser},
				{Method: http.MethodDelete, Path: "/v1/admin/admin_users/:id", Handler: h.DeleteUser},
				{Method: http.MethodPost, Path: "/v1/admin/admin_users/:id/reset_password", Handler: h.ResetUserPassword},
				{Method: http.MethodPost, Path: "/v1/admin/admin_users/:id/mfa/setup", Handler: h.MfaSetup},
				{Method: http.MethodPost, Path: "/v1/admin/admin_users/:id/mfa/confirm", Handler: h.MfaConfirm},
				{Method: http.MethodPost, Path: "/v1/admin/admin_users/:id/mfa/disable", Handler: h.MfaDisable},
				{Method: http.MethodGet, Path: "/v1/admin/display_settings", Handler: h.DisplaySettings},
				{Method: http.MethodGet, Path: "/v1/admin/op_logs", Handler: h.OperationLogs},
				{Method: http.MethodPut, Path: "/v1/admin/display_settings", Handler: h.UpdateDisplaySettings},
				{Method: http.MethodGet, Path: "/v1/admin/jobs/keys", Handler: h.ListScheduledJobKeys},
				{Method: http.MethodGet, Path: "/v1/admin/jobs", Handler: h.ListScheduledJobs},
				{Method: http.MethodPost, Path: "/v1/admin/jobs", Handler: h.CreateScheduledJob},
				{Method: http.MethodPut, Path: "/v1/admin/jobs/:id", Handler: h.UpdateScheduledJob},
				{Method: http.MethodPost, Path: "/v1/admin/jobs/:id/toggle", Handler: h.ToggleScheduledJob},
				{Method: http.MethodPost, Path: "/v1/admin/jobs/:id/run", Handler: h.RunScheduledJob},
				{Method: http.MethodGet, Path: "/v1/admin/job_workers", Handler: h.ListJobWorkerNodes},
				{Method: http.MethodGet, Path: "/v1/admin/job_runs", Handler: h.ListScheduledJobRuns},
				{Method: http.MethodGet, Path: "/v1/admin/job_runs/:id", Handler: h.GetScheduledJobRun},
				{Method: http.MethodPost, Path: "/v1/admin/job_runs/:id/retry", Handler: h.RetryScheduledJobRun},
			}...,
		),
	)
}
