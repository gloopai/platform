package contracts

const DefaultChannelNotifyContentType = "text/plain; charset=utf-8"

type PayinNotifyContentTyper interface {
	PayinNotifyContentType() string
}

type PayoutNotifyContentTyper interface {
	PayoutNotifyContentType() string
}

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
