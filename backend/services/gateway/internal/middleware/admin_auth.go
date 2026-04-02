package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gloopai/platform/gateway/internal/apiresp"
	"github.com/gloopai/platform/gateway/internal/logic/shared"
)

type AdminAuth struct {
	masterToken string
	jwtSecret   string
}

type adminIdKey struct{}

func AdminIdFromContext(ctx context.Context) int64 {
	v := ctx.Value(adminIdKey{})
	if v == nil {
		return 0
	}
	id, _ := v.(int64)
	return id
}

func NewAdminAuth(masterToken, jwtSecret string) *AdminAuth {
	return &AdminAuth{masterToken: masterToken, jwtSecret: jwtSecret}
}

func (m *AdminAuth) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
		if tok == "" {
			apiresp.Fail(w, apiresp.CodeUnauthorized, "unauthorized")
			return
		}
		if m.masterToken != "" && tok == m.masterToken {
			next(w, r)
			return
		}
		if strings.TrimSpace(m.jwtSecret) == "" {
			apiresp.Fail(w, apiresp.CodeUnauthorized, "unauthorized")
			return
		}
		adminID, err := shared.ParseAdminJWT(m.jwtSecret, tok)
		if err != nil {
			apiresp.Fail(w, apiresp.CodeUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), adminIdKey{}, adminID)
		next(w, r.WithContext(ctx))
	}
}
