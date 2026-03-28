// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"strings"
	"time"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/mockpsp"
	"github.com/gloopai/pay/channeldriver/mockpsp2"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/common/grpcclient/servicehubclient"
	"github.com/gloopai/pay/common/grpcclient/settleclient"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config config.Config

	OpenAPIParamsParseMiddleware  rest.Middleware
	MerchantSignMiddleware        rest.Middleware
	OpenAPIRateLimitMiddleware    rest.Middleware
	LoginRateLimitMiddleware      rest.Middleware
	AdminAuthMiddleware           rest.Middleware
	AdminRBACMiddleware           rest.Middleware
	MerchantConsoleAuthMiddleware rest.Middleware

	ServiceHub servicehubclient.ServiceHub

	OrderRpc    orderclient.Order
	SettleRpc   settleclient.Settle
	ChannelRpc  channelclient.Channel
	MerchantRpc merchantclient.Merchant

	NsqProducer *nsq.Producer

	RuntimeConfig *consulx.ConfigStore

	// readiness deps
	ReplayRedis    *redis.Client
	RateRedis      *redis.Client
	TradeConn      *grpc.ClientConn
	CoreConn       *grpc.ClientConn
	ServiceHubConn *grpc.ClientConn

	ChannelDrivers *channeldriver.Registry
}

func NewServiceContext(c config.Config) *ServiceContext {
	consulx.RegisterResolver()

	tradeCli := zrpc.MustNewClient(c.TradeRpc)
	coreCli := zrpc.MustNewClient(c.CoreRpc)
	serviceHubCli := zrpc.MustNewClient(c.ServiceHubRpc)

	producer, err := nsq.NewProducer(c.Nsq.NsqdTCPAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	if err := producer.Ping(); err != nil {
		panic(err)
	}

	replayAddr := strings.TrimSpace(c.ReplayGuard.RedisAddr)
	if replayAddr == "" {
		replayAddr = "127.0.0.1:6379"
	}
	replayRedis := redis.NewClient(&redis.Options{
		Addr:     replayAddr,
		Password: c.ReplayGuard.RedisPassword,
		DB:       c.ReplayGuard.RedisDB,
	})
	replayPrefix := strings.TrimSpace(c.ReplayGuard.KeyPrefix)
	if replayPrefix == "" {
		replayPrefix = "pay:openapi:replay"
	}
	replayTTL := time.Duration(c.ReplayGuard.TTLSeconds) * time.Second
	if replayTTL <= 0 {
		replayTTL = 10 * time.Minute
	}
	replayGuard := middleware.NewRedisReplayGuard(replayRedis, replayPrefix, replayTTL)

	trustForwarded := c.OpenAPI.TrustForwardedFor
	openAPIBodyMax := c.OpenAPI.MaxBodyBytes
	openAPIParamsParse := middleware.NewOpenAPIParamsParseMiddleware(openAPIBodyMax).Handle
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
		ratePrefix = "pay:openapi:ratelimit"
	}
	openAPIWindow := time.Duration(c.RateLimit.OpenAPIWindowSeconds) * time.Second
	if openAPIWindow <= 0 {
		openAPIWindow = 60 * time.Second
	}
	loginWindow := time.Duration(c.RateLimit.LoginWindowSeconds) * time.Second
	if loginWindow <= 0 {
		loginWindow = 60 * time.Second
	}
	// OpenAPILimitPerWindow: >0 为每窗口请求上限；0 表示未配置，默认 600；-1 表示关闭 OpenAPI 限流（middleware 中 limit<=0 直接放行）
	openAPILimit := c.RateLimit.OpenAPILimitPerWindow
	switch {
	case openAPILimit < 0:
		openAPILimit = 0
	case openAPILimit == 0:
		openAPILimit = 600
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
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(serviceName)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}

	chReg := channeldriver.NewRegistry()
	_ = mockpsp.RegisterAll(chReg, mockpsp.New(mockpsp.DefaultDriverKey))
	_ = mockpsp2.RegisterAll(chReg, mockpsp2.New(mockpsp2.DefaultDriverKey))

	return &ServiceContext{
		Config: c,

		OpenAPIParamsParseMiddleware:  openAPIParamsParse,
		MerchantSignMiddleware:        middleware.NewMerchantSignMiddleware(merchantclient.NewMerchant(coreCli), replayGuard, c.ReplayGuard.AllowedSkewSeconds, trustForwarded).Handle,
		OpenAPIRateLimitMiddleware:    middleware.NewOpenAPIRateLimitMiddleware(rateLimiter, ratePrefix, openAPILimit, openAPIWindow, trustForwarded).Handle,
		LoginRateLimitMiddleware:      middleware.NewLoginRateLimitMiddleware(rateLimiter, ratePrefix, loginLimit, loginWindow, trustForwarded).Handle,
		AdminAuthMiddleware:           middleware.NewAdminAuthMiddleware(c.AdminToken, c.JwtSecret).Handle,
		AdminRBACMiddleware:           middleware.NewAdminRBACMiddleware(servicehubclient.New(serviceHubCli), 10*time.Second).Handle,
		MerchantConsoleAuthMiddleware: middleware.NewMerchantConsoleAuthMiddleware(c.JwtSecret, merchantclient.NewMerchant(coreCli)).Handle,

		ServiceHub: servicehubclient.New(serviceHubCli),

		OrderRpc:    orderclient.NewOrder(tradeCli),
		SettleRpc:   settleclient.NewSettle(coreCli),
		ChannelRpc:  channelclient.NewChannel(tradeCli),
		MerchantRpc: merchantclient.NewMerchant(coreCli),

		NsqProducer:    producer,
		RuntimeConfig:  runtimeCfg,
		ReplayRedis:    replayRedis,
		RateRedis:      rateRedis,
		TradeConn:      tradeCli.Conn(),
		CoreConn:       coreCli.Conn(),
		ServiceHubConn: serviceHubCli.Conn(),
		ChannelDrivers: chReg,
	}
}
