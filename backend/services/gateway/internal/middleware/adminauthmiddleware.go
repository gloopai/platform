package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gloopai/pay/gateway/internal/store"
)

type AdminAuthMiddleware struct {
	masterToken string
	sessions    *store.SessionsStore
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

func NewAdminAuthMiddleware(masterToken string, sessions *store.SessionsStore) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{masterToken: masterToken, sessions: sessions}
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
		if m.sessions == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sum := sha256.Sum256([]byte(tok))
		hash := hex.EncodeToString(sum[:])
		sess, err := m.sessions.GetAdminSession(r.Context(), hash)
		if err != nil || sess == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), adminIdKey{}, sess.AdminId)
		next(w, r.WithContext(ctx))
	}
}
