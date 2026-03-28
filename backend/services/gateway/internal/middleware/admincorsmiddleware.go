package middleware

import (
	"net/http"
	"strings"
)

// AdminCorsMiddleware 可选：管理台前端直连 gateway（跨域）时放行 CORS。
// 本地 Vite 代理对 fetch + SSE 流式响应可能缓冲，导致收不到 notification 帧；直连 8080 可避免。
type AdminCorsMiddleware struct {
	allowed map[string]struct{}
}

func NewAdminCorsMiddleware(allowedOrigins []string) *AdminCorsMiddleware {
	m := make(map[string]struct{})
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			m[o] = struct{}{}
		}
	}
	return &AdminCorsMiddleware{allowed: m}
}

func applyAdminCorsHeaders(w http.ResponseWriter, r *http.Request, allowed map[string]struct{}) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	if _, ok := allowed[origin]; !ok {
		return false
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	// 预检里浏览器会带 Access-Control-Request-Headers，必须原样回显，否则 Chrome 会报 CORS 失败
	if reqHdr := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers")); reqHdr != "" {
		w.Header().Set("Access-Control-Allow-Headers", reqHdr)
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "X-Admin-Token, Content-Type, Accept, Accept-Language, Authorization")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	// Chrome Private Network Access：从公网页访问 localhost/内网时预检会带此头
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Access-Control-Request-Private-Network")), "true") {
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
	}
	return true
}

func (m *AdminCorsMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	if len(m.allowed) == 0 {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		applyAdminCorsHeaders(w, r, m.allowed)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
