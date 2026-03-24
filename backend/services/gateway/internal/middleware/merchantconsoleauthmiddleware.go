package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
)

type MerchantConsoleAuthMiddleware struct {
	jwtSecret string
	merchants merchantclient.Merchant
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

func NewMerchantConsoleAuthMiddleware(jwtSecret string, merchants merchantclient.Merchant) *MerchantConsoleAuthMiddleware {
	return &MerchantConsoleAuthMiddleware{
		jwtSecret: jwtSecret,
		merchants: merchants,
	}
}

func (m *MerchantConsoleAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Merchant-Token"))
		if tok == "" || strings.TrimSpace(m.jwtSecret) == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		merchantID, err := shared.ParseMerchantJWT(m.jwtSecret, tok)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if m.merchants != nil {
			auth, err := m.merchants.GetAuthInfo(r.Context(), &merchantclient.GetAuthInfoReq{MerchantId: merchantID})
			if err != nil || auth.GetStatus() != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		ctx := context.WithValue(r.Context(), merchantIdKey{}, merchantID)
		next(w, r.WithContext(ctx))
	}
}
