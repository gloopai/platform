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
	Mysql    struct {
		DataSource string
	}
	ReplayGuard struct {
		RedisAddr          string `json:",optional"`
		RedisPassword      string `json:",optional"`
		RedisDB            int    `json:",optional"`
		KeyPrefix          string `json:",optional"`
		AllowedSkewSeconds int64  `json:",optional"`
		TTLSeconds         int64  `json:",optional"`
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
	CoreRpc  zrpc.RpcClientConf
	TradeRpc zrpc.RpcClientConf
	Nsq      struct {
		NsqdTCPAddr string
		Topic       string
	}
}
