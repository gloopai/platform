package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
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
	// Upstream PSP callbacks: base URL of gateway OpenAPIServer (same host/port as signed OpenAPI) for notifyUrl (e.g. http://127.0.0.1:8090).
	Upstream struct {
		CheckoutNotifyBaseURL string `json:",optional"`
	}
}
