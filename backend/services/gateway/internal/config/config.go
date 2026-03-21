// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Mysql struct {
		DataSource string
	}
	AdminToken      string `json:",optional"`
	CheckoutBaseUrl string `json:",optional"`
	Consul          struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
	OrderRpc   zrpc.RpcClientConf
	SettleRpc  zrpc.RpcClientConf
	ChannelRpc zrpc.RpcClientConf
	Nsq        struct {
		NsqdTCPAddr string
		Topic       string
	}
}
