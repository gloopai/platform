package middleware

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type TraceHeader struct{}

func NewTraceHeader() *TraceHeader {
	return &TraceHeader{}
}

func (m *TraceHeader) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.HasTraceID() {
			w.Header().Set("X-Trace-Id", spanCtx.TraceID().String())
		}
		if spanCtx.HasSpanID() {
			w.Header().Set("X-Span-Id", spanCtx.SpanID().String())
		}
		if spanCtx.HasTraceID() || spanCtx.HasSpanID() {
			expose := w.Header().Get("Access-Control-Expose-Headers")
			w.Header().Set("Access-Control-Expose-Headers", mergeExposeHeaders(expose,
				"Traceparent",
				"X-Trace-Id",
				"X-Span-Id",
			))
		}

		next(w, r)
	}
}

func mergeExposeHeaders(exist string, headers ...string) string {
	set := map[string]struct{}{}
	var out []string

	add := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" {
			return
		}
		key := strings.ToLower(h)
		if _, ok := set[key]; ok {
			return
		}
		set[key] = struct{}{}
		out = append(out, h)
	}

	for _, part := range strings.Split(exist, ",") {
		add(part)
	}
	for _, h := range headers {
		add(h)
	}
	return strings.Join(out, ", ")
}
