// scaffold/platform-admin：网关仅对接 service-hub（管理端平台能力）。

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	ServiceName string `json:",optional"`
	AdminServer rest.RestConf
	Timezone    string `json:",optional"`
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
	AdminToken string `json:",optional"`
	JwtSecret  string `json:",optional"`
	AdminOpLog struct {
		// Exclude 支持两种格式：
		// 1) "/v1/admin/op_logs"（匹配任意方法）
		// 2) "GET /v1/admin/op_logs"（匹配指定方法）
		// 路径支持 :param 段，如 /v1/admin/rbac/menus/:id
		Exclude []string `json:",optional"`
	} `json:",optional"`
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
