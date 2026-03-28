package portalnotify

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gloopai/pay/gateway/internal/middleware"
)

// AdminNotificationsSSE streams admin notifications via in-process hub + NSQ. Requires AdminAuthMiddleware.
func AdminNotificationsSSE(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil {
			http.Error(w, "notify hub not configured", http.StatusServiceUnavailable)
			return
		}
		aid := middleware.AdminIdFromContext(r.Context())
		master := aid <= 0
		ch, unregister := hub.RegisterAdmin(master, aid)
		defer unregister()
		serveSSE(w, r, ch)
	}
}

// MerchantNotificationsSSE streams merchant notifications. Requires MerchantConsoleAuthMiddleware.
func MerchantNotificationsSSE(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil {
			http.Error(w, "notify hub not configured", http.StatusServiceUnavailable)
			return
		}
		mid := middleware.MerchantIdFromContext(r.Context())
		if mid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ch, unregister := hub.RegisterMerchant(mid)
		defer unregister()
		serveSSE(w, r, ch)
	}
}

func serveSSE(w http.ResponseWriter, r *http.Request, ch <-chan []byte) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()
	fmt.Fprintf(w, "event: connected\ndata: {\"ok\":true}\n\n")
	fl.Flush()

	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			fmt.Fprintf(w, ": ping\n\n")
			fl.Flush()
		case data, ok := <-ch:
			if !ok {
				return
			}
			if len(data) == 0 {
				continue
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			fl.Flush()
		}
	}
}
