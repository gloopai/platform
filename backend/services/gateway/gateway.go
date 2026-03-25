// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/timex"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/handler"
	adminhandler "github.com/gloopai/pay/gateway/internal/handler/admin"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var c config.Config
	conf.MustLoad(*configFile, &c)
	if err := timex.ApplyProcessTimezone(c.Timezone); err != nil {
		panic(err)
	}
	consulx.SetBaseConfig(consulx.BaseConfig{Addr: c.Consul.Addr})

	server := rest.MustNewServer(c.RestConf)
	server.Use(middleware.NewTraceHeaderMiddleware().Handle)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  "GET",
				Path:    "/ready",
				Handler: handler.ReadyHandler(ctx),
			},
		},
	)
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{ctx.AdminAuthMiddleware},
			[]rest.Route{
				{
					Method:  "GET",
					Path:    "/v1/admin/ops/services",
					Handler: adminhandler.AdminOpsServicesHandler(ctx),
				},
			}...,
		),
	)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)

	reg, err := consulx.RegisterService(c.Consul.Addr, c.Consul.Service, c.Consul.ID, fmt.Sprintf("%s:%d", c.Host, c.Port), c.Consul.Host)
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
