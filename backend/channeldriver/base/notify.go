package base

import "net/http"

const DefaultChannelNotifyContentType = "text/plain; charset=utf-8"

// PayinNotifyContentTyper optionally sets Content-Type on payin notify response.
type PayinNotifyContentTyper interface {
	PayinNotifyContentType() string
}

// PayoutNotifyContentTyper optionally sets Content-Type on payout notify response.
type PayoutNotifyContentTyper interface {
	PayoutNotifyContentType() string
}

// NotifyContentType returns the HTTP Content-Type for the response body returned to the PSP.
func NotifyContentType(drv any) string {
	if drv == nil {
		return DefaultChannelNotifyContentType
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
	return DefaultChannelNotifyContentType
}

// WriteChannelNotify writes body and Content-Type for a notify handling path.
func WriteChannelNotify(w http.ResponseWriter, drv any, body []byte) {
	w.Header().Set("Content-Type", NotifyContentType(drv))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
