package adminshell

import (
	"github.com/gloopai/platform/common/consulx"
	"github.com/gloopai/platform/gateway/internal/config"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/service-hub/hubclient"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"google.golang.org/grpc"

	"github.com/zeromicro/go-zero/zrpc"
)

// ShellConfig mirrors [config.Config] fields required by ServiceHub shell handlers (login, RBAC, jobs, …).
// Exposed so product gateways (e.g. pay) can build options without importing platform internal packages.
type ShellConfig struct {
	ServiceName string `json:",optional"`
	AdminServer rest.RestConf
	Timezone    string `json:",optional"`
	OpenAPI     struct {
		MaxBodyBytes      int64 `json:",optional"`
		TrustForwardedFor bool  `json:",optional"`
	}
	RateLimit struct {
		RedisAddr           string `json:",optional"`
		RedisPassword       string `json:",optional"`
		RedisDB             int    `json:",optional"`
		KeyPrefix           string `json:",optional"`
		LoginLimitPerWindow int64  `json:",optional"`
		LoginWindowSeconds  int64  `json:",optional"`
	}
	AdminToken string `json:",optional"`
	JwtSecret  string `json:",optional"`
	AdminOpLog struct {
		Exclude []string `json:",optional"`
	} `json:",optional"`
	Consul struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
	ServiceHubRpc zrpc.RpcClientConf
	OpsMonitor    struct {
		Services []string `json:",optional"`
	}
}

// ShellOptions wires a shared hubclient and config into the platform shell stack.
type ShellOptions struct {
	Config         ShellConfig
	ServiceHub     hubclient.ServiceHub
	RateRedis      *redis.Client
	RuntimeConfig  *consulx.ConfigStore
	ServiceHubConn *grpc.ClientConn
}

func shellConfigToInternal(c ShellConfig) config.Config {
	var out config.Config
	out.ServiceName = c.ServiceName
	out.AdminServer = c.AdminServer
	out.Timezone = c.Timezone
	out.OpenAPI = c.OpenAPI
	out.RateLimit = c.RateLimit
	out.AdminToken = c.AdminToken
	out.JwtSecret = c.JwtSecret
	out.AdminOpLog = c.AdminOpLog
	out.Consul = c.Consul
	out.ServiceHubRpc = c.ServiceHubRpc
	out.OpsMonitor = c.OpsMonitor
	return out
}

// RegisterShell registers the ServiceHub admin shell using shared hub + mapped config.
// Use this from gateways outside the platform module (e.g. pay-gateway) that cannot import internal/svc.
func RegisterShell(server *rest.Server, mw Middlewares, o ShellOptions) {
	ic := shellConfigToInternal(o.Config)
	shellCtx := svc.NewShellServiceContext(svc.ShellOptions{
		Config:         ic,
		ServiceHub:     o.ServiceHub,
		RateRedis:      o.RateRedis,
		RuntimeConfig:  o.RuntimeConfig,
		ServiceHubConn: o.ServiceHubConn,
	})
	RegisterServiceHubShell(server, mw, shellCtx)
}
