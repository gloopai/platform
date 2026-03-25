package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/common/timex"
	"github.com/gloopai/pay/notice-consumer/internal/config"
	"github.com/gloopai/pay/notice-consumer/internal/notice"
	"github.com/nsqio/go-nsq"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// notice-consumer:
	// - consume merchant_notice events from NSQ
	// - call merchant notify_url with a signed payload
	// - retry with a time-based schedule and record every attempt
	var configFile = flag.String("f", "etc/notice-consumer.yaml", "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	if err := timex.ApplyProcessTimezone(c.Timezone); err != nil {
		panic(err)
	}

	consulSvc := c.Consul.Service
	if consulSvc == "" {
		consulSvc = c.Name
	}
	if consulSvc == "" {
		consulSvc = "payment.worker.notice-consumer"
	}
	consulx.SetBaseConfig(consulx.BaseConfig{Addr: c.Consul.Addr})

	gdb, err := gorm.Open(mysql.Open(dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone)), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if sqlDB, err := gdb.DB(); err != nil {
		panic(err)
	} else if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	timeout := c.Http.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	httpClient := &http.Client{Timeout: timeout}

	// Health endpoint is used for Consul check.
	healthSrv := startHealthServer(c.Health.ListenOn)

	reg, err := consulx.RegisterService(c.Consul.Addr, consulSvc, c.Consul.ID, healthSrv.Addr, c.Consul.Host)
	if err != nil {
		panic(err)
	}

	cfg := nsq.NewConfig()
	if c.Nsq.MaxAttempts > 0 {
		cfg.MaxAttempts = uint16(c.Nsq.MaxAttempts)
	} else {
		cfg.MaxAttempts = 6
	}
	consumer, err := nsq.NewConsumer(c.Nsq.Topic, c.Nsq.Channel, cfg)
	if err != nil {
		panic(err)
	}

	processor := notice.NewProcessor(gdb, httpClient, nil)
	consumer.AddHandler(nsq.HandlerFunc(processor.HandleNSQMessage))

	if err := consumer.ConnectToNSQD(c.Nsq.NsqdTCPAddr); err != nil {
		panic(err)
	}

	// Graceful shutdown: stop consumer, deregister from Consul, shutdown health server.
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	consumer.Stop()
	select {
	case <-consumer.StopChan:
	case <-time.After(3 * time.Second):
	}
	_ = reg.Deregister()
	_ = healthSrv.Shutdown(context.Background())
}

func startHealthServer(listenOn string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Use a dedicated server with short header timeout to minimize resource usage.
	srv := &http.Server{
		Addr:              listenOn,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		// Bind explicitly to fail fast on port conflicts.
		ln, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			panic(err)
		}
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return srv
}
