package svc

import (
	"strings"
	"time"

	"github.com/gloopai/platform/common/configkv"
	"github.com/gloopai/platform/common/consulx"
	"github.com/gloopai/platform/common/gatewaymw"
	"github.com/gloopai/platform/common/requestx"
	"github.com/gloopai/platform/service-hub/hubclient"
	"github.com/gloopai/platform/gateway/internal/apiresp"
	"github.com/gloopai/platform/gateway/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config config.Config

	OpenAPIParamsParse rest.Middleware
	LoginRateLimit     rest.Middleware
	AdminAuth          rest.Middleware
	AdminRBAC          rest.Middleware
	AdminOpLog         rest.Middleware

	ServiceHub hubclient.ServiceHub

	RuntimeConfig *consulx.ConfigStore

	RateRedis      *redis.Client
	ServiceHubConn *grpc.ClientConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	consulx.RegisterResolver()

	serviceHubCli := zrpc.MustNewClient(c.ServiceHubRpc)

	trustForwarded := c.OpenAPI.TrustForwardedFor
	openAPIBodyMax := c.OpenAPI.MaxBodyBytes
	if openAPIBodyMax <= 0 {
		openAPIBodyMax = 262144
	}
	openAPIParamsParse := gatewaymw.NewOpenAPIParamsParse(gatewaymw.OpenAPIParamsParseOptions{
		MaxBodyBytes:        openAPIBodyMax,
		Fail:                apiresp.Fail,
		CodePayloadTooLarge: apiresp.CodePayloadTooLarge,
		CodeInvalidParams:   apiresp.CodeInvalidParams,
	}).Handle

	rateRedisAddr := strings.TrimSpace(c.RateLimit.RedisAddr)
	if rateRedisAddr == "" {
		rateRedisAddr = "127.0.0.1:6379"
	}
	rateRedis := redis.NewClient(&redis.Options{
		Addr:     rateRedisAddr,
		Password: c.RateLimit.RedisPassword,
		DB:       c.RateLimit.RedisDB,
	})
	rateLimiter := gatewaymw.NewRedisRateLimiter(rateRedis)
	ratePrefix := strings.TrimSpace(c.RateLimit.KeyPrefix)
	if ratePrefix == "" {
		ratePrefix = "gateway:admin:ratelimit"
	}
	loginWindow := time.Duration(c.RateLimit.LoginWindowSeconds) * time.Second
	if loginWindow <= 0 {
		loginWindow = 60 * time.Second
	}
	loginLimit := c.RateLimit.LoginLimitPerWindow
	if loginLimit <= 0 {
		loginLimit = 60
	}

	serviceName := strings.TrimSpace(c.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(c.AdminServer.Name)
	}
	if serviceName == "" {
		serviceName = "gateway"
	}
	var runtimeCfg *consulx.ConfigStore
	if cfg, err := consulx.NewConfigStore("", configkv.GlobalConfigPrefix(), configkv.ServiceConfigPrefix(serviceName)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}

	serviceHub := hubclient.New(serviceHubCli)

	return &ServiceContext{
		Config: c,

		OpenAPIParamsParse: openAPIParamsParse,
		LoginRateLimit: gatewaymw.NewLoginRateLimit(gatewaymw.LoginRateLimitOptions{
			Limiter:             rateLimiter,
			KeyPrefix:           ratePrefix,
			Limit:               loginLimit,
			Window:              loginWindow,
			TrustForwarded:      trustForwarded,
			Fail:                apiresp.Fail,
			CodeInvalidParams:   apiresp.CodeInvalidParams,
			CodeUnavailable:     apiresp.CodeUnavailable,
			CodeTooManyRequests: apiresp.CodeTooManyRequests,
		}).Handle,
		AdminAuth: gatewaymw.NewAdminAuth(gatewaymw.AdminAuthOptions{
			MasterToken:      c.AdminToken,
			JWTSecret:        c.JwtSecret,
			Fail:             apiresp.Fail,
			CodeUnauthorized: apiresp.CodeUnauthorized,
		}).Handle,
		AdminRBAC: gatewaymw.NewAdminRBAC(gatewaymw.AdminRBACOptions[*hubclient.AdminApiRule]{
			Hub:              serviceHub,
			TTL:              10 * time.Second,
			Fail:             apiresp.Fail,
			AdminIDFromCtx:   gatewaymw.AdminIDFromContext,
			CodeUnauthorized: apiresp.CodeUnauthorized,
			CodeForbidden:    apiresp.CodeForbidden,
		}).Handle,
		AdminOpLog: gatewaymw.NewAdminOpLog(gatewaymw.AdminOpLogOptions[*hubclient.AdminApiRule]{
			Hub:              oplogServiceHub{sh: serviceHub},
			TrustForwarded:   trustForwarded,
			Excludes:         c.AdminOpLog.Exclude,
			RequestIDFromCtx: requestx.FromContext,
			HeaderRequestID:  requestx.HeaderRequestID,
			AdminIDFromCtx:   gatewaymw.AdminIDFromContext,
		}).Handle,

		ServiceHub: serviceHub,

		RuntimeConfig:  runtimeCfg,
		RateRedis:      rateRedis,
		ServiceHubConn: serviceHubCli.Conn(),
	}
}

