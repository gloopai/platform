// Package gatewaymw holds HTTP middleware shared by platform and product gateways (RBAC, etc.).
package gatewaymw

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

// RbacRule is satisfied by service-hub AdminApiRule protobuf messages.
type RbacRule interface {
	GetStatus() int64
	GetMethod() string
	GetPathPattern() string
	GetPermKey() string
}

// RbacHub is the minimal service-hub surface for admin API RBAC.
type RbacHub[TRow RbacRule] interface {
	GetAdminRbacMyPerms(ctx context.Context, adminUserID int64) (isSuper bool, permKeys []string, err error)
	ListAdminApiRules(ctx context.Context, page, pageSize int64, q, permKey string) ([]TRow, int64, error)
}

// AdminRBACOptions configures [AdminRBAC].
type AdminRBACOptions[TRow RbacRule] struct {
	Hub RbacHub[TRow]
	TTL time.Duration
	// Fail writes a JSON error envelope (e.g. apiresp.Fail).
	Fail func(w http.ResponseWriter, code int, message string)
	// AdminIDFromCtx reads the authenticated admin user id (e.g. from JWT middleware).
	AdminIDFromCtx func(ctx context.Context) int64
	CodeUnauthorized int
	CodeForbidden    int
}

// AdminRBAC enforces permission keys for admin APIs.
//
// Behavior:
// - For admin APIs without a registered permission key: deny (fail-closed)
// - For super_admin: allow all
// - Cache perms per admin_user_id for a short TTL
type AdminRBAC[TRow RbacRule] struct {
	svcHub RbacHub[TRow]
	ttl    time.Duration

	fail             func(w http.ResponseWriter, code int, message string)
	adminIDFromCtx   func(ctx context.Context) int64
	codeUnauthorized int
	codeForbidden    int

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

// NewAdminRBAC builds RBAC middleware. TTL defaults to 10s when <= 0.
func NewAdminRBAC[TRow RbacRule](opt AdminRBACOptions[TRow]) *AdminRBAC[TRow] {
	ttl := opt.TTL
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return &AdminRBAC[TRow]{
		svcHub:           opt.Hub,
		ttl:              ttl,
		fail:             opt.Fail,
		adminIDFromCtx:   opt.AdminIDFromCtx,
		codeUnauthorized: opt.CodeUnauthorized,
		codeForbidden:    opt.CodeForbidden,
		cache:            make(map[int64]permCache),
	}
}

func (m *AdminRBAC[TRow]) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/admin/") {
			next(w, r)
			return
		}
		if r.URL.Path == "/v1/admin/login" {
			next(w, r)
			return
		}

		adminID := m.adminIDFromCtx(r.Context())
		if adminID <= 0 {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}

		if adminSessionBaselineOK(r) {
			next(w, r)
			return
		}

		isSuper, keys, err := m.getPerms(r, adminID)
		if err != nil {
			m.fail(w, m.codeForbidden, err.Error())
			return
		}
		if isSuper {
			next(w, r)
			return
		}

		required, err := m.requiredPerm(r)
		if err != nil {
			m.fail(w, m.codeForbidden, err.Error())
			return
		}
		if required == "" {
			m.fail(w, m.codeForbidden, "forbidden: no api rule for this path")
			return
		}
		if _, ok := keys[required]; ok {
			next(w, r)
			return
		}
		m.fail(w, m.codeForbidden, "forbidden")
	}
}

func adminSessionBaselineOK(r *http.Request) bool {
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	path := strings.TrimSpace(r.URL.Path)
	if method == http.MethodPost && path == "/v1/admin/logout" {
		return true
	}
	if method != http.MethodGet {
		return false
	}
	switch path {
	case "/v1/admin/me", "/v1/admin/rbac/my_menu", "/v1/admin/display_settings":
		return true
	default:
		return false
	}
}

func (m *AdminRBAC[TRow]) requiredPerm(r *http.Request) (string, error) {
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
		if MatchPathPattern(ru.pattern, path) {
			return ru.permKey, nil
		}
	}
	if method == http.MethodGet && path == "/v1/admin/psp_driver_channel_config_schema" {
		return "admin.channels.read", nil
	}
	return "", nil
}

func (m *AdminRBAC[TRow]) getApiRules(r *http.Request) ([]apiRule, error) {
	now := time.Now()
	m.ruleMu.Lock()
	if now.Before(m.ruleCache.expiresAt) {
		out := m.ruleCache.rules
		m.ruleMu.Unlock()
		return out, nil
	}
	m.ruleMu.Unlock()

	rows, _, err := m.svcHub.ListAdminApiRules(r.Context(), 0, 0, "", "")
	if err != nil {
		return nil, err
	}
	out := make([]apiRule, 0, len(rows))
	for _, rr := range rows {
		if rbacRowIsNil(rr) || rr.GetStatus() != 1 {
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

func (m *AdminRBAC[TRow]) getPerms(r *http.Request, adminID int64) (bool, map[string]struct{}, error) {
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

func rbacRowIsNil[TRow RbacRule](rr TRow) bool {
	v := reflect.ValueOf(rr)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}
