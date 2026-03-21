// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/handler"
	"github.com/gloopai/pay/gateway/internal/registry"
	"github.com/gloopai/pay/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)

	reg, err := registry.Register(c.Consul.Addr, c.Consul.Service, c.Consul.ID, fmt.Sprintf("%s:%d", c.Host, c.Port), c.Consul.Host)
	if err != nil {
		panic(err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go server.Start()
	<-signalCtx.Done()
	_ = reg.Deregister()
	server.Stop()
}
