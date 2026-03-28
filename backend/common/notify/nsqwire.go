package notify

import "encoding/json"

// PortalNotifyTopic 是 NSQ topic；各 gateway 实例用不同 channel 名订阅以收全量副本。
const PortalNotifyTopic = "pay.notify.portal"

// PortalNSQMessage 是 NSQ 消息体 JSON；Envelope 为 SSE data 负载。
type PortalNSQMessage struct {
	Portal       string          `json:"portal"` // "admin" | "merchant"
	Broadcast    bool            `json:"broadcast"`
	AdminUserIDs []int64         `json:"admin_user_ids,omitempty"`
	MerchantIDs  []string        `json:"merchant_ids,omitempty"`
	Envelope     json.RawMessage `json:"envelope"`
}
