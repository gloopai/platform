package handler

import (
	"net/http"
	"time"

	"github.com/gloopai/platform/gateway/internal/apiresp"
)

// HealthHandler 供运维探活与本地管理台「运维监控」页联调，无需鉴权。
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiresp.OK(w, map[string]any{
			"status":       "ok",
			"service":      "gateway",
			"timestamp_ms": time.Now().UnixMilli(),
		})
	}
}
