package svc

import (
	"github.com/gloopai/platform/gateway/internal/config"
	"github.com/gloopai/platform/service-hub/hubclient"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/gloopai/platform/common/consulx"
)

// ShellOptions carries the dependencies needed by ServiceHub-backed admin shell handlers (login, RBAC, users, jobs, …).
// Middlewares are registered separately via [github.com/gloopai/platform/gateway/pkg/adminshell.Register].
type ShellOptions struct {
	Config         config.Config
	ServiceHub     hubclient.ServiceHub
	RateRedis      *redis.Client
	RuntimeConfig  *consulx.ConfigStore
	ServiceHubConn *grpc.ClientConn
}

// NewShellServiceContext builds a [ServiceContext] with only shell-related fields set.
// Used when the gateway binary already owns zrpc clients and [hubclient.ServiceHub] (e.g. pay-gateway sharing CoreRpc).
func NewShellServiceContext(o ShellOptions) *ServiceContext {
	return &ServiceContext{
		Config:         o.Config,
		ServiceHub:     o.ServiceHub,
		RateRedis:      o.RateRedis,
		RuntimeConfig:  o.RuntimeConfig,
		ServiceHubConn: o.ServiceHubConn,
	}
}
