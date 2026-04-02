package gatewaymw

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements a fixed-window counter in Redis.
type RedisRateLimiter struct {
	cli *redis.Client
}

// NewRedisRateLimiter builds a Redis-backed limiter.
func NewRedisRateLimiter(cli *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{cli: cli}
}

// Allow increments key and returns whether the request is within limit for the window.
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

// OpenAPIRateLimitOptions configures [OpenAPIRateLimit].
type OpenAPIRateLimitOptions struct {
	Limiter *RedisRateLimiter
	// KeyPrefix is prepended to Redis keys (e.g. pay:openapi:ratelimit).
	KeyPrefix string
	Limit     int64
	Window    time.Duration
	// TrustForwarded enables X-Forwarded-For / X-Real-IP for client IP (only behind a trusted proxy).
	TrustForwarded bool

	Fail func(w http.ResponseWriter, code int, message string)
	CodeInvalidParams   int
	CodeUnavailable     int
	CodeTooManyRequests int
}

// OpenAPIRateLimit limits OpenAPI traffic per client IP + app/merchant account.
type OpenAPIRateLimit struct {
	limiter *RedisRateLimiter
	keyPrefix string
	limit int64
	window time.Duration
	trustForwarded bool

	fail                func(w http.ResponseWriter, code int, message string)
	codeInvalidParams   int
	codeUnavailable     int
	codeTooManyRequests int
}

// NewOpenAPIRateLimit builds OpenAPI rate-limit middleware.
func NewOpenAPIRateLimit(opt OpenAPIRateLimitOptions) *OpenAPIRateLimit {
	return &OpenAPIRateLimit{
		limiter:             opt.Limiter,
		keyPrefix:           opt.KeyPrefix,
		limit:               opt.Limit,
		window:              opt.Window,
		trustForwarded:      opt.TrustForwarded,
		fail:                opt.Fail,
		codeInvalidParams:   opt.CodeInvalidParams,
		codeUnavailable:     opt.CodeUnavailable,
		codeTooManyRequests: opt.CodeTooManyRequests,
	}
}

func (m *OpenAPIRateLimit) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := ReadMergedParams(r)
		if err != nil {
			m.fail(w, m.codeInvalidParams, "invalid params")
			return
		}
		account := strings.TrimSpace(params["app_id"])
		if account == "" {
			account = strings.TrimSpace(params["merchant_id"])
		}
		if account == "" {
			account = "unknown"
		}
		ip := ClientHost(r, m.trustForwarded)
		key := m.keyPrefix + ":openapi:" + ip + ":" + account
		ok, err := m.limiter.Allow(r.Context(), key, m.limit, m.window)
		if err != nil {
			m.fail(w, m.codeUnavailable, "rate limiter unavailable")
			return
		}
		if !ok {
			m.fail(w, m.codeTooManyRequests, "too many requests")
			return
		}
		next(w, r)
	}
}

// LoginRateLimitOptions configures [LoginRateLimit].
type LoginRateLimitOptions struct {
	Limiter *RedisRateLimiter
	KeyPrefix string
	Limit int64
	Window time.Duration
	TrustForwarded bool

	Fail func(w http.ResponseWriter, code int, message string)
	CodeInvalidParams   int
	CodeUnavailable     int
	CodeTooManyRequests int
}

// LoginRateLimit limits login-style endpoints per IP + account fields in the body/query.
type LoginRateLimit struct {
	limiter *RedisRateLimiter
	keyPrefix string
	limit int64
	window time.Duration
	trustForwarded bool

	fail                func(w http.ResponseWriter, code int, message string)
	codeInvalidParams   int
	codeUnavailable     int
	codeTooManyRequests int
}

// NewLoginRateLimit builds login rate-limit middleware.
func NewLoginRateLimit(opt LoginRateLimitOptions) *LoginRateLimit {
	return &LoginRateLimit{
		limiter:             opt.Limiter,
		keyPrefix:           opt.KeyPrefix,
		limit:               opt.Limit,
		window:              opt.Window,
		trustForwarded:      opt.TrustForwarded,
		fail:                opt.Fail,
		codeInvalidParams:   opt.CodeInvalidParams,
		codeUnavailable:     opt.CodeUnavailable,
		codeTooManyRequests: opt.CodeTooManyRequests,
	}
}

func (m *LoginRateLimit) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := ReadMergedParams(r)
		if err != nil {
			m.fail(w, m.codeInvalidParams, "invalid params")
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
		ip := ClientHost(r, m.trustForwarded)
		key := m.keyPrefix + ":login:" + r.URL.Path + ":" + ip + ":" + account
		ok, err := m.limiter.Allow(r.Context(), key, m.limit, m.window)
		if err != nil {
			m.fail(w, m.codeUnavailable, "rate limiter unavailable")
			return
		}
		if !ok {
			m.fail(w, m.codeTooManyRequests, "too many requests")
			return
		}
		next(w, r)
	}
}
