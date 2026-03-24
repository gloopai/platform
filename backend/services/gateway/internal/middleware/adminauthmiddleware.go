package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gloopai/pay/gateway/internal/logic/shared"
)

type AdminAuthMiddleware struct {
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

func NewAdminAuthMiddleware(masterToken, jwtSecret string) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{masterToken: masterToken, jwtSecret: jwtSecret}
}

func (m *AdminAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
		if tok == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if m.masterToken != "" && tok == m.masterToken {
			next(w, r)
			return
		}
		if strings.TrimSpace(m.jwtSecret) == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		adminID, err := shared.ParseAdminJWT(m.jwtSecret, tok)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), adminIdKey{}, adminID)
		next(w, r.WithContext(ctx))
	}
}
