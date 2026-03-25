package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Timezone string `json:",optional"`
	Mysql    struct {
		DataSource             string
		MaxOpenConns           int   `json:",optional"`
		MaxIdleConns           int   `json:",optional"`
		ConnMaxLifetimeSeconds int64 `json:",optional"`
	}
	Consul struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
}
