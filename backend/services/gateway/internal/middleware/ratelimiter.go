package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
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
	limiter   *RedisRateLimiter
	keyPrefix string
	limit     int64
	window    time.Duration
}

func NewOpenAPIRateLimitMiddleware(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration) *OpenAPIRateLimitMiddleware {
	return &OpenAPIRateLimitMiddleware{
		limiter:   limiter,
		keyPrefix: keyPrefix,
		limit:     limit,
		window:    window,
	}
}

func (m *OpenAPIRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", "invalid params")
			return
		}
		merchantID := strings.TrimSpace(params["merchant_id"])
		if merchantID == "" {
			merchantID = "unknown"
		}
		ip := clientIP(r.RemoteAddr)
		key := m.keyPrefix + ":openapi:" + ip + ":" + merchantID
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
	limiter   *RedisRateLimiter
	keyPrefix string
	limit     int64
	window    time.Duration
}

func NewLoginRateLimitMiddleware(limiter *RedisRateLimiter, keyPrefix string, limit int64, window time.Duration) *LoginRateLimitMiddleware {
	return &LoginRateLimitMiddleware{
		limiter:   limiter,
		keyPrefix: keyPrefix,
		limit:     limit,
		window:    window,
	}
}

func (m *LoginRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", "invalid params")
			return
		}
		account := strings.TrimSpace(params["merchant_id"])
		if account == "" {
			account = strings.TrimSpace(params["username"])
		}
		if account == "" {
			account = "unknown"
		}
		ip := clientIP(r.RemoteAddr)
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

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return host
}

func parseWindowSeconds(v int64, fallback int64) time.Duration {
	if v <= 0 {
		v = fallback
	}
	return time.Duration(v) * time.Second
}

func mustInt64(v int64, fallback int64) int64 {
	if v <= 0 {
		return fallback
	}
	return v
}

func joinParts(parts ...string) string {
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteString(p)
	}
	return b.String()
}

func withWindowSuffix(prefix string, window time.Duration) string {
	return joinParts(prefix, strconv.FormatInt(int64(window/time.Second), 10))
}
