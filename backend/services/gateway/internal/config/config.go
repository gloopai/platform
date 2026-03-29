// scaffold/platform-admin：网关仅对接 service-hub（管理端平台能力）。

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	ServiceName string `json:",optional"`
	AdminServer rest.RestConf
	// 以下 Rest 配置段保留为零值兼容旧 YAML；进程仅使用 AdminServer。
	MerchantServer rest.RestConf `json:",optional"`
	OpenAPIServer  rest.RestConf `json:",optional"`
	CheckoutServer rest.RestConf `json:",optional"`
	Timezone       string        `json:",optional"`
	OpenAPI        struct {
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
	AdminToken      string `json:",optional"`
	JwtSecret       string `json:",optional"`
	CheckoutBaseUrl string `json:",optional"`
	Consul          struct {
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
