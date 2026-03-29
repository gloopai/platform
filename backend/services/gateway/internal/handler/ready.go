package handler

import (
	"net/http"
	"time"

	"github.com/gloopai/pay/common/healthx"
	"github.com/gloopai/pay/gateway/internal/svc"
)

func ReadyHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	checks := []healthx.Check{}
	if ctx.ServiceHubConn != nil {
		checks = append(checks, healthx.GRPCHealthCheck("service_hub_grpc", ctx.ServiceHubConn, "", 1200*time.Millisecond))
	}
	if ctx.RateRedis != nil {
		checks = append(checks, healthx.RedisPing("redis_ratelimit", ctx.RateRedis))
	}
	return healthx.HTTPHandler(checks...)
}
