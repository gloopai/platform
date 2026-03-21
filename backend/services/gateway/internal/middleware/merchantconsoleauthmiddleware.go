package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gloopai/pay/gateway/internal/store"
)

type MerchantConsoleAuthMiddleware struct {
	sessions  *store.SessionsStore
	merchants *store.MerchantsStore
}

type merchantIdKey struct{}

func MerchantIdFromContext(ctx context.Context) string {
	v := ctx.Value(merchantIdKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func NewMerchantConsoleAuthMiddleware(sessions *store.SessionsStore, merchants *store.MerchantsStore) *MerchantConsoleAuthMiddleware {
	return &MerchantConsoleAuthMiddleware{
		sessions:  sessions,
		merchants: merchants,
	}
}

func (m *MerchantConsoleAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Merchant-Token"))
		if tok == "" || m.sessions == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sum := sha256.Sum256([]byte(tok))
		hash := hex.EncodeToString(sum[:])
		sess, err := m.sessions.GetMerchantSession(r.Context(), hash)
		if err != nil || sess == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if m.merchants != nil {
			me, err := m.merchants.GetByMerchantId(r.Context(), sess.MerchantId)
			if err != nil || me.Status != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		ctx := context.WithValue(r.Context(), merchantIdKey{}, sess.MerchantId)
		next(w, r.WithContext(ctx))
	}
}

