package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/grpcclient/servicehubclient"
)

// AdminRBACMiddleware enforces permission keys for admin APIs.
//
// Behavior:
// - For admin APIs without a registered permission key: deny (fail-closed)
// - For super_admin: allow all
// - Cache perms per admin_user_id for a short TTL
type AdminRBACMiddleware struct {
	svcHub servicehubclient.ServiceHub
	ttl    time.Duration

	mu    sync.Mutex
	cache map[int64]permCache

	ruleMu    sync.Mutex
	ruleCache apiRuleCache
}

type permCache struct {
	expiresAt time.Time
	isSuper   bool
	keys      map[string]struct{}
}

type apiRuleCache struct {
	expiresAt time.Time
	rules     []apiRule
}

func NewAdminRBACMiddleware(svcHub servicehubclient.ServiceHub, ttl time.Duration) *AdminRBACMiddleware {
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return &AdminRBACMiddleware{svcHub: svcHub, ttl: ttl, cache: make(map[int64]permCache)}
}

func (m *AdminRBACMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only protect admin APIs; login handled elsewhere.
		if !strings.HasPrefix(r.URL.Path, "/v1/admin/") {
			next(w, r)
			return
		}
		if r.URL.Path == "/v1/admin/login" {
			next(w, r)
			return
		}

		adminID := AdminIdFromContext(r.Context())
		if adminID <= 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		isSuper, keys, err := m.getPerms(r, adminID)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if isSuper {
			next(w, r)
			return
		}

		required, err := m.requiredPerm(r)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if required == "" {
			// fail-closed: endpoint exists but no rule configured
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if _, ok := keys[required]; ok {
			next(w, r)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	}
}

func (m *AdminRBACMiddleware) requiredPerm(r *http.Request) (string, error) {
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	path := strings.TrimSpace(r.URL.Path)
	rules, err := m.getApiRules(r)
	if err != nil {
		return "", err
	}
	for _, ru := range rules {
		if ru.method != method {
			continue
		}
		if matchPattern(ru.pattern, path) {
			return ru.permKey, nil
		}
	}
	return "", nil
}

func (m *AdminRBACMiddleware) getApiRules(r *http.Request) ([]apiRule, error) {
	now := time.Now()
	m.ruleMu.Lock()
	if now.Before(m.ruleCache.expiresAt) {
		out := m.ruleCache.rules
		m.ruleMu.Unlock()
		return out, nil
	}
	m.ruleMu.Unlock()

	rows, err := m.svcHub.ListAdminApiRules(r.Context())
	if err != nil {
		return nil, err
	}
	out := make([]apiRule, 0, len(rows))
	for _, rr := range rows {
		if rr == nil || rr.GetStatus() != 1 {
			continue
		}
		out = append(out, apiRule{
			method:  strings.ToUpper(strings.TrimSpace(rr.GetMethod())),
			pattern: strings.TrimSpace(rr.GetPathPattern()),
			permKey: strings.TrimSpace(rr.GetPermKey()),
		})
	}
	m.ruleMu.Lock()
	m.ruleCache = apiRuleCache{expiresAt: now.Add(m.ttl), rules: out}
	m.ruleMu.Unlock()
	return out, nil
}

func (m *AdminRBACMiddleware) getPerms(r *http.Request, adminID int64) (bool, map[string]struct{}, error) {
	now := time.Now()
	m.mu.Lock()
	if c, ok := m.cache[adminID]; ok && now.Before(c.expiresAt) {
		m.mu.Unlock()
		return c.isSuper, c.keys, nil
	}
	m.mu.Unlock()

	isSuper, permKeys, err := m.svcHub.GetAdminRbacMyPerms(r.Context(), adminID)
	if err != nil {
		return false, nil, err
	}
	keys := make(map[string]struct{}, len(permKeys))
	for _, k := range permKeys {
		k = strings.TrimSpace(k)
		if k == "" || k == "*" {
			continue
		}
		keys[k] = struct{}{}
	}

	m.mu.Lock()
	m.cache[adminID] = permCache{
		expiresAt: now.Add(m.ttl),
		isSuper:   isSuper,
		keys:      keys,
	}
	m.mu.Unlock()
	return isSuper, keys, nil
}

type apiRule struct {
	method  string
	pattern string // supports ":param" segments
	permKey string
}

func matchPattern(pattern, path string) bool {
	if pattern == path {
		return true
	}
	ps := splitPath(pattern)
	as := splitPath(path)
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

func splitPath(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}
