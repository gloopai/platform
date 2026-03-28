package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/gloopai/pay/common/consulx"
	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/common/timex"
	"github.com/gloopai/pay/trade/internal/config"
	"github.com/gloopai/pay/trade/internal/server"
	"github.com/gloopai/pay/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/trade.yaml", "the config file")

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

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		orderpb.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)

	reg, err := consulx.RegisterService(c.Consul.Addr, c.Consul.Service, c.Consul.ID, c.ListenOn, c.Consul.Host)
	if err != nil {
		panic(err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go s.Start()
	<-signalCtx.Done()
	_ = reg.Deregister()
	s.Stop()
}
