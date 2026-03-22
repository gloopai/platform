package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
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
}
