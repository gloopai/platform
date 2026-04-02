package svc

import (
	"strings"
	"time"

	"github.com/gloopai/platform/common/configkv"
	"github.com/gloopai/platform/common/consulx"
	"github.com/gloopai/platform/common/grpcclient/servicehubclient"
	"github.com/gloopai/platform/gateway/internal/config"
	"github.com/gloopai/platform/gateway/internal/middleware"
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

	ServiceHub servicehubclient.ServiceHub

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
	openAPIParamsParse := middleware.NewOpenAPIParamsParse(openAPIBodyMax).Handle

	rateRedisAddr := strings.TrimSpace(c.RateLimit.RedisAddr)
	if rateRedisAddr == "" {
		rateRedisAddr = "127.0.0.1:6379"
	}
	rateRedis := redis.NewClient(&redis.Options{
		Addr:     rateRedisAddr,
		Password: c.RateLimit.RedisPassword,
		DB:       c.RateLimit.RedisDB,
	})
	rateLimiter := middleware.NewRedisRateLimiter(rateRedis)
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

	return &ServiceContext{
		Config: c,

		OpenAPIParamsParse: openAPIParamsParse,
		LoginRateLimit: middleware.NewLoginRateLimit(
			rateLimiter, ratePrefix, loginLimit, loginWindow, trustForwarded,
		).Handle,
		AdminAuth:  middleware.NewAdminAuth(c.AdminToken, c.JwtSecret).Handle,
		AdminRBAC:  middleware.NewAdminRBAC(servicehubclient.New(serviceHubCli), 10*time.Second).Handle,
		AdminOpLog: middleware.NewAdminOpLog(servicehubclient.New(serviceHubCli), trustForwarded, c.AdminOpLog.Exclude).Handle,

		ServiceHub: servicehubclient.New(serviceHubCli),

		RuntimeConfig:  runtimeCfg,
		RateRedis:      rateRedis,
		ServiceHubConn: serviceHubCli.Conn(),
	}
}
