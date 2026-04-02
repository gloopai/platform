package gatewaymw

import (
	"context"
	"net/http"
	"strings"

	"github.com/gloopai/platform/common/jwtutil"
)

type merchantIDKey struct{}

// MerchantIDFromContext returns the merchant id set by [MerchantConsoleAuth], or empty.
func MerchantIDFromContext(ctx context.Context) string {
	v := ctx.Value(merchantIDKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// MerchantConsoleAuthOptions configures JWT validation for merchant console APIs.
type MerchantConsoleAuthOptions struct {
	JWTSecret string
	// Fail writes a JSON error envelope (e.g. apiresp.Fail).
	Fail func(w http.ResponseWriter, code int, message string)
	CodeUnauthorized int
	// VerifyMerchant optional: if non-nil, called after JWT parse; any error → unauthorized.
	VerifyMerchant func(ctx context.Context, merchantID string) error
}

// MerchantConsoleAuth validates X-Merchant-Token (HS256 JWT with merchant subject).
type MerchantConsoleAuth struct {
	jwtSecret        string
	fail             func(w http.ResponseWriter, code int, message string)
	codeUnauthorized int
	verifyMerchant   func(ctx context.Context, merchantID string) error
}

// NewMerchantConsoleAuth builds merchant console auth middleware.
func NewMerchantConsoleAuth(opt MerchantConsoleAuthOptions) *MerchantConsoleAuth {
	return &MerchantConsoleAuth{
		jwtSecret:        opt.JWTSecret,
		fail:             opt.Fail,
		codeUnauthorized: opt.CodeUnauthorized,
		verifyMerchant:   opt.VerifyMerchant,
	}
}

func (m *MerchantConsoleAuth) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(r.Header.Get("X-Merchant-Token"))
		if tok == "" || strings.TrimSpace(m.jwtSecret) == "" {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}
		merchantID, err := jwtutil.ParseMerchantJWT(m.jwtSecret, tok)
		if err != nil {
			m.fail(w, m.codeUnauthorized, "unauthorized")
			return
		}
		if m.verifyMerchant != nil {
			if err := m.verifyMerchant(r.Context(), merchantID); err != nil {
				m.fail(w, m.codeUnauthorized, "unauthorized")
				return
			}
		}
		ctx := context.WithValue(r.Context(), merchantIDKey{}, merchantID)
		next(w, r.WithContext(ctx))
	}
}
