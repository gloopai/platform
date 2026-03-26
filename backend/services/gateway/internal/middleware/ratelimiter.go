package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gloopai/pay/gateway/internal/openapi"
	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	cli *redis.Client
}

func NewRedisRateLimiter(cli *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{cli: cli}
}

func (l *RedisRateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	if limit <= 0 || window <= 0 {
		return true, nil
	}
	n, err := l.cli.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = l.cli.Expire(ctx, key, window).Err()
	}
	return n <= limit, nil
}

type OpenAPIRateLimitMiddleware struct {
	limiter              *RedisRateLimiter
	keyPrefix            string
	limit                int64
	window               time.Duration
	trustForwardedForIPs bool
}

func NewOpenAPIRateLimitMiddleware(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration, trustForwardedForIPs bool) *OpenAPIRateLimitMiddleware {
	return &OpenAPIRateLimitMiddleware{
		limiter:              limiter,
		keyPrefix:            keyPrefix,
		limit:                limit,
		window:               window,
		trustForwardedForIPs: trustForwardedForIPs,
	}
}

func (m *OpenAPIRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", "invalid params")
			return
		}
		account := strings.TrimSpace(params["app_id"])
		if account == "" {
			account = strings.TrimSpace(params["merchant_id"])
		}
		if account == "" {
			account = "unknown"
		}
		ip := ClientHost(r, m.trustForwardedForIPs)
		key := m.keyPrefix + ":openapi:" + ip + ":" + account
		ok, err := m.limiter.Allow(r.Context(), key, m.limit, m.window)
		if err != nil {
			openapi.Write(w, http.StatusServiceUnavailable, "UNAVAILABLE", "rate limiter unavailable")
			return
		}
		if !ok {
			openapi.Write(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "too many requests")
			return
		}
		next(w, r)
	}
}

type LoginRateLimitMiddleware struct {
	limiter              *RedisRateLimiter
	keyPrefix            string
	limit                int64
	window               time.Duration
	trustForwardedForIPs bool
}

func NewLoginRateLimitMiddleware(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration, trustForwardedForIPs bool) *LoginRateLimitMiddleware {
	return &LoginRateLimitMiddleware{
		limiter:              limiter,
		keyPrefix:            keyPrefix,
		limit:                limit,
		window:               window,
		trustForwardedForIPs: trustForwardedForIPs,
	}
}

func (m *LoginRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", "invalid params")
			return
		}
		account := strings.TrimSpace(params["email"])
		if account == "" {
			account = strings.TrimSpace(params["merchant_id"])
		}
		if account == "" {
			account = strings.TrimSpace(params["username"])
		}
		if account == "" {
			account = "unknown"
		}
		ip := ClientHost(r, m.trustForwardedForIPs)
		key := m.keyPrefix + ":login:" + r.URL.Path + ":" + ip + ":" + account
		ok, err := m.limiter.Allow(r.Context(), key, m.limit, m.window)
		if err != nil {
			openapi.Write(w, http.StatusServiceUnavailable, "UNAVAILABLE", "rate limiter unavailable")
			return
		}
		if !ok {
			openapi.Write(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "too many requests")
			return
		}
		next(w, r)
	}
}
