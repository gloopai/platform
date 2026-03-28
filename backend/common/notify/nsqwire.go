package notify

import "encoding/json"

// PortalNotifyTopic 是 portal 通知 NSQ topic（service-hub 发布；如有消费者需独立 channel）。
const PortalNotifyTopic = "pay.notify.portal"

// PortalNSQMessage 是 NSQ 消息体 JSON；Envelope 为 SSE data 负载。
type PortalNSQMessage struct {
	Portal       string          `json:"portal"` // "admin" | "merchant"
	Broadcast    bool            `json:"broadcast"`
	AdminUserIDs []int64         `json:"admin_user_ids,omitempty"`
	MerchantIDs  []string        `json:"merchant_ids,omitempty"`
	Envelope     json.RawMessage `json:"envelope"`
}
