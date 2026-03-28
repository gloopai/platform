package requestx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

const HeaderRequestID = "X-Request-Id"

type requestIDKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	id = strings.TrimSpace(id)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

func Ensure(r *http.Request, w http.ResponseWriter) *http.Request {
	reqID := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if reqID == "" {
		reqID = newID()
	}
	w.Header().Set(HeaderRequestID, reqID)
	return r.WithContext(withRequestID(r.Context(), reqID))
}

func FromContext(ctx context.Context) string {
	v := ctx.Value(requestIDKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func TraceIDFromContext(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

func newID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "rid_fallback"
	}
	return hex.EncodeToString(b[:])
}
