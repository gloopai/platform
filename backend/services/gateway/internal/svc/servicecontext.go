// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/common/grpcclient/settleclient"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/store"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	MerchantSignMiddleware        rest.Middleware
	OpenAPIRateLimitMiddleware    rest.Middleware
	LoginRateLimitMiddleware      rest.Middleware
	AdminAuthMiddleware           rest.Middleware
	MerchantConsoleAuthMiddleware rest.Middleware

	// 仅管理台账号；业务数据经 Trade/Core RPC。
	AdminUsers     *store.AdminUsersStore
	GlobalSettings *store.GlobalSettingsStore
	PayoutOrders   *store.PayoutOrdersStore

	OrderRpc    orderclient.Order
	SettleRpc   settleclient.Settle
	ChannelRpc  channelclient.Channel
	MerchantRpc merchantclient.Merchant

	NsqProducer *nsq.Producer

	RuntimeConfig *consulx.ConfigStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone))
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	consulx.RegisterResolver()

	tradeCli := zrpc.MustNewClient(c.TradeRpc)
	coreCli := zrpc.MustNewClient(c.CoreRpc)

	producer, err := nsq.NewProducer(c.Nsq.NsqdTCPAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	if err := producer.Ping(); err != nil {
		panic(err)
	}

	adminUsersStore := store.NewAdminUsersStore(sqlDB)
	globalSettingsStore := store.NewGlobalSettingsStore(sqlDB)
	payoutOrdersStore := store.NewPayoutOrdersStore(sqlDB)
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
	openAPILimit := c.RateLimit.OpenAPILimitPerWindow
	if openAPILimit <= 0 {
		openAPILimit = 600
	}
	loginLimit := c.RateLimit.LoginLimitPerWindow
	if loginLimit <= 0 {
		loginLimit = 60
	}
	var runtimeCfg *consulx.ConfigStore
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}

	return &ServiceContext{
		Config: c,

		MerchantSignMiddleware:        middleware.NewMerchantSignMiddleware(merchantclient.NewMerchant(coreCli), replayGuard, c.ReplayGuard.AllowedSkewSeconds).Handle,
		OpenAPIRateLimitMiddleware:    middleware.NewOpenAPIRateLimitMiddleware(rateLimiter, ratePrefix, openAPILimit, openAPIWindow).Handle,
		LoginRateLimitMiddleware:      middleware.NewLoginRateLimitMiddleware(rateLimiter, ratePrefix, loginLimit, loginWindow).Handle,
		AdminAuthMiddleware:           middleware.NewAdminAuthMiddleware(c.AdminToken, c.JwtSecret).Handle,
		MerchantConsoleAuthMiddleware: middleware.NewMerchantConsoleAuthMiddleware(c.JwtSecret, merchantclient.NewMerchant(coreCli)).Handle,

		AdminUsers:     adminUsersStore,
		GlobalSettings: globalSettingsStore,
		PayoutOrders:   payoutOrdersStore,

		OrderRpc:    orderclient.NewOrder(tradeCli),
		SettleRpc:   settleclient.NewSettle(coreCli),
		ChannelRpc:  channelclient.NewChannel(tradeCli),
		MerchantRpc: merchantclient.NewMerchant(coreCli),

		NsqProducer:   producer,
		RuntimeConfig: runtimeCfg,
	}
}
