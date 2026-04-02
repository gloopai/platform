package gatewaymw

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OpLogRecord is the audit payload sent to service-hub (field names align with RecordAdminOperationLogReq).
type OpLogRecord struct {
	RequestID     string
	AdminUserID   int64
	AdminUsername string
	OperatorIP    string
	UserAgent     string
	Method        string
	Path          string
	QueryString   string
	PermKey       string
	HTTPStatus    int32
	Success       bool
	DurationMs    int64
	ErrorMessage  string
}

// OpLogHub is the minimal service-hub surface for admin operation logging.
type OpLogHub[TRow RbacRule] interface {
	ListAdminApiRules(ctx context.Context, page, pageSize int64, q, permKey string) ([]TRow, int64, error)
	FetchAdminUsername(ctx context.Context, adminUserID int64) string
	RecordOpLog(ctx context.Context, rec OpLogRecord) error
}

// AdminOpLogOptions configures [AdminOpLog].
type AdminOpLogOptions[TRow RbacRule] struct {
	Hub OpLogHub[TRow]

	TrustForwarded bool
	Excludes       []string

	RequestIDFromCtx func(context.Context) string
	HeaderRequestID  string

	AdminIDFromCtx func(context.Context) int64

	UserCacheTTL time.Duration
	RuleCacheTTL time.Duration
}

// AdminOpLog records admin API calls asynchronously via service-hub.
type AdminOpLog[TRow RbacRule] struct {
	hub OpLogHub[TRow]

	trustForwarded bool
	excludeRules   []opExcludeRule

	userCacheTTL  time.Duration
	userCacheMu   sync.Mutex
	userCacheByID map[int64]cachedAdminUser

	ruleCacheTTL    time.Duration
	ruleCacheMu     sync.Mutex
	ruleCacheExpire time.Time
	rules           []opLogApiRule

	requestIDFromCtx func(context.Context) string
	headerRequestID  string
	adminIDFromCtx   func(context.Context) int64
}

type cachedAdminUser struct {
	expiresAt time.Time
	username  string
}

type opLogApiRule struct {
	method  string
	pattern string
	permKey string
}

type opExcludeRule struct {
	method  string // empty means all methods
	pattern string
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// NewAdminOpLog builds admin operation-log middleware.
func NewAdminOpLog[TRow RbacRule](opt AdminOpLogOptions[TRow]) *AdminOpLog[TRow] {
	userTTL := opt.UserCacheTTL
	if userTTL <= 0 {
		userTTL = 30 * time.Second
	}
	ruleTTL := opt.RuleCacheTTL
	if ruleTTL <= 0 {
		ruleTTL = 10 * time.Second
	}
	return &AdminOpLog[TRow]{
		hub:              opt.Hub,
		trustForwarded:   opt.TrustForwarded,
		excludeRules:     parseOpLogExcludeRules(opt.Excludes),
		userCacheTTL:     userTTL,
		userCacheByID:    make(map[int64]cachedAdminUser),
		ruleCacheTTL:     ruleTTL,
		requestIDFromCtx: opt.RequestIDFromCtx,
		headerRequestID:  strings.TrimSpace(opt.HeaderRequestID),
		adminIDFromCtx:   opt.AdminIDFromCtx,
	}
}

func (m *AdminOpLog[TRow]) Handle(next http.HandlerFunc) http.HandlerFunc {
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

		adminID := m.adminIDFromCtx(r.Context())
		if adminID <= 0 {
			return
		}

		username := m.cachedAdminUsername(r.Context(), adminID)
		permKey := m.matchPermKey(r.Context(), method, path)
		reqID := ""
		if m.requestIDFromCtx != nil {
			reqID = strings.TrimSpace(m.requestIDFromCtx(r.Context()))
		}
		if reqID == "" && m.headerRequestID != "" {
			reqID = strings.TrimSpace(r.Header.Get(m.headerRequestID))
		}
		durationMs := time.Since(start).Milliseconds()
		statusCode := sw.statusCode
		success := statusCode >= 200 && statusCode < 400
		errMsg := ""
		if !success {
			errMsg = http.StatusText(statusCode)
		}

		rec := OpLogRecord{
			RequestID:     reqID,
			AdminUserID:   adminID,
			AdminUsername: username,
			OperatorIP:    ClientHost(r, m.trustForwarded),
			UserAgent:     strings.TrimSpace(r.UserAgent()),
			Method:        method,
			Path:          path,
			QueryString:   strings.TrimSpace(r.URL.RawQuery),
			PermKey:       permKey,
			HTTPStatus:    int32(statusCode),
			Success:       success,
			DurationMs:    durationMs,
			ErrorMessage:  errMsg,
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = m.hub.RecordOpLog(ctx, rec)
		}()
	}
}

func (m *AdminOpLog[TRow]) shouldExclude(method, path string) bool {
	for _, ex := range m.excludeRules {
		if ex.method != "" && ex.method != method {
			continue
		}
		if MatchPathPattern(ex.pattern, path) {
			return true
		}
	}
	return false
}

func (m *AdminOpLog[TRow]) cachedAdminUsername(ctx context.Context, adminID int64) string {
	now := time.Now()
	m.userCacheMu.Lock()
	if c, ok := m.userCacheByID[adminID]; ok && now.Before(c.expiresAt) {
		m.userCacheMu.Unlock()
		return c.username
	}
	m.userCacheMu.Unlock()

	username := m.hub.FetchAdminUsername(ctx, adminID)
	m.userCacheMu.Lock()
	m.userCacheByID[adminID] = cachedAdminUser{expiresAt: now.Add(m.userCacheTTL), username: username}
	m.userCacheMu.Unlock()
	return username
}

func (m *AdminOpLog[TRow]) matchPermKey(ctx context.Context, method, path string) string {
	rules := m.getRules(ctx)
	for _, ru := range rules {
		if ru.method != method {
			continue
		}
		if MatchPathPattern(ru.pattern, path) {
			return ru.permKey
		}
	}
	return ""
}

func (m *AdminOpLog[TRow]) getRules(ctx context.Context) []opLogApiRule {
	now := time.Now()
	m.ruleCacheMu.Lock()
	if now.Before(m.ruleCacheExpire) {
		out := m.rules
		m.ruleCacheMu.Unlock()
		return out
	}
	m.ruleCacheMu.Unlock()

	apiRules, _, err := m.hub.ListAdminApiRules(ctx, 0, 0, "", "")
	if err != nil {
		return nil
	}
	out := make([]opLogApiRule, 0, len(apiRules))
	for _, rr := range apiRules {
		if rbacRowIsNil(rr) || rr.GetStatus() != 1 {
			continue
		}
		out = append(out, opLogApiRule{
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

func parseOpLogExcludeRules(items []string) []opExcludeRule {
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
