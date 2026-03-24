package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientHost returns the client IP (host only) for rate limiting and IP whitelist checks.
// When trustForwarded is false, only RemoteAddr is used (safe default).
// When true, X-Forwarded-For (first hop) or X-Real-IP is used if present — only enable behind a trusted reverse proxy.
func ClientHost(r *http.Request, trustForwarded bool) string {
	if trustForwarded {
		if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				host := strings.TrimSpace(parts[0])
				if host != "" {
					return stripZone(host)
				}
			}
		}
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
			return stripZone(xri)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func stripZone(ip string) string {
	// IPv6 zone id: fe80::1%en0
	if i := strings.IndexByte(ip, '%'); i >= 0 {
		return ip[:i]
	}
	return ip
}
