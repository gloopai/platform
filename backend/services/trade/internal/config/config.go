package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	// CoreRpc 用于通道/支付产品等配置读接口（Channel gRPC 已注册在 core）。
	CoreRpc zrpc.RpcClientConf `json:",optional"`
	Timezone string `json:",optional"`
	Mysql    struct {
		DataSource string
	}
	BizRedis struct {
		Addr     string
		Password string
		DB       int
	}
	Consul struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
	// Channel async callbacks: base URL of gateway OpenAPIServer (same host/port as signed OpenAPI) for notifyUrl (e.g. http://127.0.0.1:8090).
	Channel struct {
		CheckoutNotifyBaseURL string `json:",optional"`
	}
}
