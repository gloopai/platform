// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Timezone string `json:",optional"`
	OpenAPI  struct {
		// MaxBodyBytes caps JSON body size for signed OpenAPI and login param parsing (default 262144).
		MaxBodyBytes int64 `json:",optional"`
		// TrustForwardedFor: when true, IP whitelist and rate limits use X-Forwarded-For / X-Real-IP. Only enable behind a trusted reverse proxy.
		TrustForwardedFor bool `json:",optional"`
	}
	Mysql struct {
		DataSource             string
		MaxOpenConns           int   `json:",optional"`
		MaxIdleConns           int   `json:",optional"`
		ConnMaxLifetimeSeconds int64 `json:",optional"`
	}
	ReplayGuard struct {
		RedisAddr          string `json:",optional"`
		RedisPassword      string `json:",optional"`
		RedisDB            int    `json:",optional"`
		KeyPrefix          string `json:",optional"`
		AllowedSkewSeconds int64  `json:",optional"`
		TTLSeconds         int64  `json:",optional"`
	}
	RateLimit struct {
		RedisAddr     string `json:",optional"`
		RedisPassword string `json:",optional"`
		RedisDB       int    `json:",optional"`
		KeyPrefix     string `json:",optional"`
		// OpenAPILimitPerWindow: per-window cap; 0 = default 600; -1 = disable (see NewServiceContext).
		OpenAPILimitPerWindow int64 `json:",optional"`
		OpenAPIWindowSeconds  int64 `json:",optional"`
		LoginLimitPerWindow   int64 `json:",optional"`
		LoginWindowSeconds    int64 `json:",optional"`
	}
	AdminToken      string `json:",optional"`
	JwtSecret       string `json:",optional"`
	CheckoutBaseUrl string `json:",optional"`
	Consul          struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
	CoreRpc    zrpc.RpcClientConf
	TradeRpc   zrpc.RpcClientConf
	OpsMonitor struct {
		// Extra Consul service names to include on admin ops page, e.g. notice-consumer.
		Services []string `json:",optional"`
	}
	Nsq struct {
		NsqdTCPAddr string
		Topic       string
	}
}
