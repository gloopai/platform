// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"strings"
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

	ctx := svc.NewServiceContext(c)

	// scaffold/platform-admin: 仅启动 Admin HTTP（商户 / OpenAPI / Checkout 端口不监听；配置文件中仍可保留对应段供参考）
	adminServer := rest.MustNewServer(c.AdminServer)
	adminServer.Use(middleware.NewTraceHeader().Handle)
	handler.RegisterCommonHandlers(adminServer, ctx)
	handler.RegisterAdminHandlers(adminServer, ctx)
	servers := []*rest.Server{adminServer}

	adminServer.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ready",
				Handler: handler.ReadyHandler(ctx),
			},
		},
	)
	adminServer.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{ctx.AdminAuth},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/v1/admin/ops/services",
					Handler: adminhandler.AdminOpsServicesHandler(ctx),
				},
			}...,
		),
	)

	fmt.Printf("Starting admin API at %s:%d...\n", c.AdminServer.Host, c.AdminServer.Port)

	regService := strings.TrimSpace(c.Consul.Service)
	if regService == "" {
		regService = "gateway-admin-api"
	}
	reg, err := consulx.RegisterService(c.Consul.Addr, regService, c.Consul.ID, fmt.Sprintf("%s:%d", c.AdminServer.Host, c.AdminServer.Port), c.Consul.Host)
	if err != nil {
		panic(err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, s := range servers {
		go s.Start()
	}
	<-signalCtx.Done()
	_ = reg.Deregister()
	for _, s := range servers {
		s.Stop()
	}
}
