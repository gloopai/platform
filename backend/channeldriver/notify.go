package channeldriver

import "net/http"

const DefaultUpstreamNotifyContentType = "text/plain; charset=utf-8"

// NotifyContentType returns the HTTP Content-Type for the response body returned to the PSP.
// If drv implements PayinNotifyContentTyper / PayoutNotifyContentTyper, that value is used.
func NotifyContentType(drv any) string {
	if drv == nil {
		return DefaultUpstreamNotifyContentType
	}
	if t, ok := drv.(PayinNotifyContentTyper); ok {
		if s := t.PayinNotifyContentType(); s != "" {
			return s
		}
	}
	if t, ok := drv.(PayoutNotifyContentTyper); ok {
		if s := t.PayoutNotifyContentType(); s != "" {
			return s
		}
	}
	return DefaultUpstreamNotifyContentType
}

// WriteUpstreamNotify writes body and Content-Type for a successful notify handling path.
// Typical body comes from PayinUpstream.PayinNotifyResponse / PayoutUpstream.PayoutNotifyResponse.
func WriteUpstreamNotify(w http.ResponseWriter, drv any, body []byte) {
	w.Header().Set("Content-Type", NotifyContentType(drv))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
