package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/platform/common/grpcclient/servicehubclient"
	"github.com/gloopai/platform/gateway/internal/requestx"
)

type AdminOpLog struct {
	svcHub          servicehubclient.ServiceHub
	trustForwarded  bool
	excludeRules    []opExcludeRule
	userCacheTTL    time.Duration
	userCacheMu     sync.Mutex
	userCacheByID   map[int64]cachedAdminUser
	ruleCacheTTL    time.Duration
	ruleCacheMu     sync.Mutex
	ruleCacheExpire time.Time
	rules           []opApiRule
}

type cachedAdminUser struct {
	expiresAt time.Time
	username  string
}

type opApiRule struct {
	method  string
	pattern string
	permKey string
}

type opExcludeRule struct {
	method  string // empty means all methods
	pattern string
}

func NewAdminOpLog(svcHub servicehubclient.ServiceHub, trustForwarded bool, excludes []string) *AdminOpLog {
	return &AdminOpLog{
		svcHub:         svcHub,
		trustForwarded: trustForwarded,
		excludeRules:   parseExcludeRules(excludes),
		userCacheTTL:   30 * time.Second,
		userCacheByID:  make(map[int64]cachedAdminUser),
		ruleCacheTTL:   10 * time.Second,
	}
}

func (m *AdminOpLog) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/admin/") || r.URL.Path == "/v1/admin/login" {
			next(w, r)
			return
		}
		method := strings.ToUpper(strings.TrimSpace(r.Method))
		path := strings.TrimSpace(r.URL.Path)
		if m.shouldExclude(method, path) {
			next(w, r)
			return
		}
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(sw, r)

		adminID := AdminIdFromContext(r.Context())
		if adminID <= 0 {
			return
		}

		username := m.cachedAdminUsername(r.Context(), adminID)
		permKey := m.matchPermKey(r.Context(), method, path)
		reqID := strings.TrimSpace(requestx.FromContext(r.Context()))
		if reqID == "" {
			reqID = strings.TrimSpace(r.Header.Get(requestx.HeaderRequestID))
		}
		durationMs := time.Since(start).Milliseconds()
		statusCode := sw.statusCode
		success := statusCode >= 200 && statusCode < 400
		errMsg := ""
		if !success {
			errMsg = http.StatusText(statusCode)
		}

		row := &servicehubclient.RecordAdminOperationLogReq{
			RequestId:     reqID,
			AdminUserId:   adminID,
			AdminUsername: username,
			OperatorIp:    ClientHost(r, m.trustForwarded),
			UserAgent:     strings.TrimSpace(r.UserAgent()),
			Method:        method,
			Path:          path,
			QueryString:   strings.TrimSpace(r.URL.RawQuery),
			PermKey:       permKey,
			HttpStatus:    int32(statusCode),
			Success:       success,
			DurationMs:    durationMs,
			ErrorMessage:  errMsg,
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = m.svcHub.RecordAdminOperationLog(ctx, row)
		}()
	}
}

func (m *AdminOpLog) shouldExclude(method, path string) bool {
	for _, r := range m.excludeRules {
		if r.method != "" && r.method != method {
			continue
		}
		if matchApiRulePattern(r.pattern, path) {
			return true
		}
	}
	return false
}

func (m *AdminOpLog) cachedAdminUsername(ctx context.Context, adminID int64) string {
	now := time.Now()
	m.userCacheMu.Lock()
	if c, ok := m.userCacheByID[adminID]; ok && now.Before(c.expiresAt) {
		m.userCacheMu.Unlock()
		return c.username
	}
	m.userCacheMu.Unlock()

	username := ""
	if u, err := m.svcHub.GetAdminUserById(ctx, adminID); err == nil && u != nil {
		username = strings.TrimSpace(u.GetUsername())
	}
	m.userCacheMu.Lock()
	m.userCacheByID[adminID] = cachedAdminUser{expiresAt: now.Add(m.userCacheTTL), username: username}
	m.userCacheMu.Unlock()
	return username
}

func (m *AdminOpLog) matchPermKey(ctx context.Context, method, path string) string {
	rules := m.getRules(ctx)
	for _, ru := range rules {
		if ru.method != method {
			continue
		}
		if matchApiRulePattern(ru.pattern, path) {
			return ru.permKey
		}
	}
	return ""
}

func (m *AdminOpLog) getRules(ctx context.Context) []opApiRule {
	now := time.Now()
	m.ruleCacheMu.Lock()
	if now.Before(m.ruleCacheExpire) {
		out := m.rules
		m.ruleCacheMu.Unlock()
		return out
	}
	m.ruleCacheMu.Unlock()

	apiRules, _, err := m.svcHub.ListAdminApiRules(ctx, 0, 0, "", "")
	if err != nil {
		return nil
	}
	out := make([]opApiRule, 0, len(apiRules))
	for _, rr := range apiRules {
		if rr == nil || rr.GetStatus() != 1 {
			continue
		}
		out = append(out, opApiRule{
			method:  strings.ToUpper(strings.TrimSpace(rr.GetMethod())),
			pattern: strings.TrimSpace(rr.GetPathPattern()),
			permKey: strings.TrimSpace(rr.GetPermKey()),
		})
	}
	m.ruleCacheMu.Lock()
	m.rules = out
	m.ruleCacheExpire = now.Add(m.ruleCacheTTL)
	m.ruleCacheMu.Unlock()
	return out
}

func matchApiRulePattern(pattern, path string) bool {
	if pattern == path {
		return true
	}
	ps := splitApiPath(pattern)
	as := splitApiPath(path)
	if len(ps) != len(as) {
		return false
	}
	for i := 0; i < len(ps); i++ {
		if strings.HasPrefix(ps[i], ":") {
			if as[i] == "" {
				return false
			}
			continue
		}
		if ps[i] != as[i] {
			return false
		}
	}
	return true
}

func splitApiPath(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}

func parseExcludeRules(items []string) []opExcludeRule {
	out := make([]opExcludeRule, 0, len(items))
	for _, raw := range items {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		parts := strings.Fields(v)
		if len(parts) >= 2 && strings.HasPrefix(parts[1], "/") {
			out = append(out, opExcludeRule{
				method:  strings.ToUpper(parts[0]),
				pattern: strings.TrimSpace(parts[1]),
			})
			continue
		}
		if strings.HasPrefix(v, "/") {
			out = append(out, opExcludeRule{pattern: v})
		}
	}
	return out
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
