package adminshell

import (
	adminhandler "github.com/gloopai/platform/gateway/internal/handler/admin"
	"github.com/gloopai/platform/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterServiceHubShell registers the standard admin shell (login, RBAC, users, display settings, scheduled jobs, …)
// backed by ServiceHub. The gateway binary must supply a [svc.ServiceContext] with ServiceHub and Config populated
// (see [svc.NewShellServiceContext] when reusing an existing hub client).
func RegisterServiceHubShell(server *rest.Server, mw Middlewares, shellCtx *svc.ServiceContext) {
	Register(server, mw, ServiceHubHandlers(shellCtx))
}

// ServiceHubHandlers binds platform shell HTTP handlers to shellCtx.
func ServiceHubHandlers(shellCtx *svc.ServiceContext) Handlers {
	return Handlers{
		Login:                 adminhandler.AdminLoginHandler(shellCtx),
		Logout:                adminhandler.AdminLogoutHandler(shellCtx),
		Me:                    adminhandler.AdminMeHandler(shellCtx),
		MyMenu:                adminhandler.AdminMyMenuHandler(shellCtx),
		ListRbacRoles:         adminhandler.AdminListRbacRolesHandler(shellCtx),
		CreateRbacRole:        adminhandler.AdminCreateRbacRoleHandler(shellCtx),
		UpdateRbacRole:        adminhandler.AdminUpdateRbacRoleHandler(shellCtx),
		DeleteRbacRole:        adminhandler.AdminDeleteRbacRoleHandler(shellCtx),
		ListRbacMenus:         adminhandler.AdminListRbacMenusHandler(shellCtx),
		CreateRbacMenu:        adminhandler.AdminCreateRbacMenuHandler(shellCtx),
		UpdateRbacMenu:        adminhandler.AdminUpdateRbacMenuHandler(shellCtx),
		DeleteRbacMenu:        adminhandler.AdminDeleteRbacMenuHandler(shellCtx),
		GetRbacRoleMenus:      adminhandler.AdminGetRbacRoleMenusHandler(shellCtx),
		SetRbacRoleMenus:      adminhandler.AdminSetRbacRoleMenusHandler(shellCtx),
		GetRbacUserRoles:      adminhandler.AdminGetRbacUserRolesHandler(shellCtx),
		SetRbacUserRoles:      adminhandler.AdminSetRbacUserRolesHandler(shellCtx),
		ListRbacPermissions:   adminhandler.AdminListRbacPermissionsHandler(shellCtx),
		CreateRbacPermission:  adminhandler.AdminCreateRbacPermissionHandler(shellCtx),
		UpdateRbacPermission:  adminhandler.AdminUpdateRbacPermissionHandler(shellCtx),
		DeleteRbacPermission:  adminhandler.AdminDeleteRbacPermissionHandler(shellCtx),
		GetRbacRolePermKeys:   adminhandler.AdminGetRbacRolePermKeysHandler(shellCtx),
		SetRbacRolePermKeys:   adminhandler.AdminSetRbacRolePermKeysHandler(shellCtx),
		ListRbacApiRules:      adminhandler.AdminListRbacApiRulesHandler(shellCtx),
		UpsertRbacApiRule:     adminhandler.AdminUpsertRbacApiRuleHandler(shellCtx),
		DeleteRbacApiRule:     adminhandler.AdminDeleteRbacApiRuleHandler(shellCtx),
		ListUsers:             adminhandler.AdminListUsersHandler(shellCtx),
		CreateUser:            adminhandler.AdminCreateUserHandler(shellCtx),
		UpdateUser:            adminhandler.AdminUpdateUserHandler(shellCtx),
		DeleteUser:            adminhandler.AdminDeleteUserHandler(shellCtx),
		ResetUserPassword:     adminhandler.AdminResetUserPasswordHandler(shellCtx),
		MfaSetup:              adminhandler.AdminMfaSetupHandler(shellCtx),
		MfaConfirm:            adminhandler.AdminMfaConfirmHandler(shellCtx),
		MfaDisable:            adminhandler.AdminMfaDisableHandler(shellCtx),
		DisplaySettings:       adminhandler.AdminDisplaySettingsHandler(shellCtx),
		OperationLogs:         adminhandler.AdminOperationLogsHandler(shellCtx),
		UpdateDisplaySettings: adminhandler.AdminUpdateDisplaySettingsHandler(shellCtx),
		ListScheduledJobKeys:  adminhandler.AdminListScheduledJobKeysHandler(shellCtx),
		ListScheduledJobs:     adminhandler.AdminListScheduledJobsHandler(shellCtx),
		CreateScheduledJob:    adminhandler.AdminCreateScheduledJobHandler(shellCtx),
		UpdateScheduledJob:    adminhandler.AdminUpdateScheduledJobHandler(shellCtx),
		ToggleScheduledJob:    adminhandler.AdminToggleScheduledJobHandler(shellCtx),
		RunScheduledJob:       adminhandler.AdminRunScheduledJobHandler(shellCtx),
		ListJobWorkerNodes:    adminhandler.AdminListJobWorkerNodesHandler(shellCtx),
		ListScheduledJobRuns:  adminhandler.AdminListScheduledJobRunsHandler(shellCtx),
		GetScheduledJobRun:    adminhandler.AdminGetScheduledJobRunHandler(shellCtx),
		RetryScheduledJobRun:  adminhandler.AdminRetryScheduledJobRunHandler(shellCtx),
	}
}
