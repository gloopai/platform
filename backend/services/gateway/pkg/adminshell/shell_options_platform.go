package adminshell

import (
	"github.com/gloopai/platform/gateway/internal/svc"
)

// ShellOptionsFromPlatformGateway builds [ShellOptions] from the platform scaffold gateway [svc.ServiceContext]
// (e.g. platform-admin gateway main). Use pay-gateway’s [svc.ServiceContext.AdminShellOptions] for pay.
func ShellOptionsFromPlatformGateway(ctx *svc.ServiceContext) ShellOptions {
	c := ctx.Config
	var sc ShellConfig
	sc.ServiceName = c.ServiceName
	sc.AdminServer = c.AdminServer
	sc.Timezone = c.Timezone
	sc.OpenAPI = c.OpenAPI
	sc.RateLimit = c.RateLimit
	sc.AdminToken = c.AdminToken
	sc.JwtSecret = c.JwtSecret
	sc.AdminOpLog = c.AdminOpLog
	sc.Consul = c.Consul
	sc.ServiceHubRpc = c.ServiceHubRpc
	sc.OpsMonitor = c.OpsMonitor
	return ShellOptions{
		Config:         sc,
		ServiceHub:     ctx.ServiceHub,
		RateRedis:      ctx.RateRedis,
		RuntimeConfig:  ctx.RuntimeConfig,
		ServiceHubConn: ctx.ServiceHubConn,
	}
}
