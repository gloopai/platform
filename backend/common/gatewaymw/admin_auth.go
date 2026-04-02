package gatewaymw

import (
	"context"
	"net/http"
	"strings"

	"github.com/gloopai/platform/common/jwtutil"
)

type adminIDKey struct{}

// AdminIDFromContext returns the admin user id set by [AdminAuth], or 0.
func AdminIDFromContext(ctx context.Context) int64 {
	v := ctx.Value(adminIDKey{})
	if v == nil {
		return 0
	}
	id, _ := v.(int64)
	return id
}

// AdminAuthOptions configures [AdminAuth].
type AdminAuthOptions struct {
	MasterToken string
	JWTSecret   string
	// Fail writes a JSON error envelope (e.g. apiresp.Fail).
	Fail func(w http.ResponseWriter, code int, message string)
	// CodeUnauthorized matches gateway apiresp (e.g. 4010).
	CodeUnauthorized int
}

// AdminAuth validates X-Admin-Token (master token or JWT).
type AdminAuth struct {
	masterToken      string
	jwtSecret        string
	fail             func(w http.ResponseWriter, code int, message string)
	codeUnauthorized int
}

// NewAdminAuth builds admin auth middleware.
func NewAdminAuth(opt AdminAuthOptions) *AdminAuth {
	return &AdminAuth{
		masterToken:      opt.MasterToken,
		jwtSecret:        opt.JWTSecret,
		fail:             opt.Fail,
		codeUnauthorized: opt.CodeUnauthorized,
	}
}

func (m *AdminAuth) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
		if tok == "" {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}
		if m.masterToken != "" && tok == m.masterToken {
			next(w, r)
			return
		}
		if strings.TrimSpace(m.jwtSecret) == "" {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}
		adminID, err := jwtutil.ParseAdminJWT(m.jwtSecret, tok)
		if err != nil {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), adminIDKey{}, adminID)
		next(w, r.WithContext(ctx))
	}
}
