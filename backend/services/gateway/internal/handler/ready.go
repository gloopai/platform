package handler

import (
	"net/http"
	"time"

	"github.com/gloopai/pay/common/healthx"
	"github.com/gloopai/pay/gateway/internal/svc"
)

func ReadyHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	checks := []healthx.Check{
		healthx.GormPing(ctx.Gorm),
	}
	if ctx.ReplayRedis != nil {
		checks = append(checks, healthx.RedisPing("redis_replay", ctx.ReplayRedis))
	}
	if ctx.RateRedis != nil {
		checks = append(checks, healthx.RedisPing("redis_ratelimit", ctx.RateRedis))
	}
	if ctx.TradeConn != nil {
		checks = append(checks, healthx.GRPCHealthCheck("trade_grpc", ctx.TradeConn, "", 1200*time.Millisecond))
	}
	if ctx.CoreConn != nil {
		checks = append(checks, healthx.GRPCHealthCheck("core_grpc", ctx.CoreConn, "", 1200*time.Millisecond))
	}
	return healthx.HTTPHandler(checks...)
}
