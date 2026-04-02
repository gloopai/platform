package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gloopai/platform/gateway/internal/apiresp"
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

type OpenAPIRateLimit struct {
	limiter              *RedisRateLimiter
	keyPrefix            string
	limit                int64
	window               time.Duration
	trustForwardedForIPs bool
}

func NewOpenAPIRateLimit(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration, trustForwardedForIPs bool) *OpenAPIRateLimit {
	return &OpenAPIRateLimit{
		limiter:              limiter,
		keyPrefix:            keyPrefix,
		limit:                limit,
		window:               window,
		trustForwardedForIPs: trustForwardedForIPs,
	}
}

func (m *OpenAPIRateLimit) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, "invalid params")
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
			apiresp.Fail(w, apiresp.CodeUnavailable, "rate limiter unavailable")
			return
		}
		if !ok {
			apiresp.Fail(w, apiresp.CodeTooManyRequests, "too many requests")
			return
		}
		next(w, r)
	}
}

type LoginRateLimit struct {
	limiter              *RedisRateLimiter
	keyPrefix            string
	limit                int64
	window               time.Duration
	trustForwardedForIPs bool
}

func NewLoginRateLimit(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration, trustForwardedForIPs bool) *LoginRateLimit {
	return &LoginRateLimit{
		limiter:              limiter,
		keyPrefix:            keyPrefix,
		limit:                limit,
		window:               window,
		trustForwardedForIPs: trustForwardedForIPs,
	}
}

func (m *LoginRateLimit) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, "invalid params")
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
			apiresp.Fail(w, apiresp.CodeUnavailable, "rate limiter unavailable")
			return
		}
		if !ok {
			apiresp.Fail(w, apiresp.CodeTooManyRequests, "too many requests")
			return
		}
		next(w, r)
	}
}
