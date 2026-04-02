// Package requestx carries X-Request-Id through request context for gateways and downstream logic.
package requestx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// HeaderRequestID is the canonical request ID header name.
const HeaderRequestID = "X-Request-Id"

type requestIDKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	id = strings.TrimSpace(id)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

// Ensure ensures a request ID is present (from header or generated), sets the response header, and stores it in context.
func Ensure(r *http.Request, w http.ResponseWriter) *http.Request {
	reqID := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if reqID == "" {
		reqID = newID()
	}
	w.Header().Set(HeaderRequestID, reqID)
	return r.WithContext(withRequestID(r.Context(), reqID))
}

// FromContext returns the request ID stored by Ensure, or empty.
func FromContext(ctx context.Context) string {
	v := ctx.Value(requestIDKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// TraceIDFromContext returns the OpenTelemetry trace id string when present.
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
