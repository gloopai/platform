package requestx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const HeaderRequestID = "X-Request-Id"

type requestIDKey struct{}

func Ensure(r *http.Request, w http.ResponseWriter) *http.Request {
	reqID := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if reqID == "" {
		reqID = newID()
	}
	w.Header().Set(HeaderRequestID, reqID)
	ctx := context.WithValue(r.Context(), requestIDKey{}, reqID)
	return r.WithContext(ctx)
}

func FromContext(ctx context.Context) string {
	v := ctx.Value(requestIDKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func newID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "rid_fallback"
	}
	return hex.EncodeToString(b[:])
}
