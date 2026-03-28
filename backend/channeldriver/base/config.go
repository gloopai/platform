package base

import "encoding/json"

// ChannelConfig is a snapshot of one channels table row (plus optional extras) used per request.
// Multiple rows may share the same DriverKey with different credentials.
type ChannelConfig struct {
	ChannelID int64  `json:"channel_id"`
	DriverKey string `json:"driver_key"` // protocol implementation key, e.g. "psp_india_a"

	// GatewayBaseURL is the origin for upstream HTTP API (e.g. https://api.example.com), without trailing slash.
	GatewayBaseURL string `json:"gateway_base_url"`

	// AppID is the upstream application / merchant id (e.g. appId in PSP docs).
	AppID string `json:"app_id,omitempty"`

	// ChannelMerchantNo is an alternate merchant identifier if the PSP uses both.
	ChannelMerchantNo string `json:"channel_merchant_no,omitempty"`

	SignSecret       string `json:"-"` // do not log in plaintext in production
	RSAPrivateKeyPEM string `json:"-"`

	SupportsPayin  bool `json:"supports_payin"`
	SupportsPayout bool `json:"supports_payout"`

	// Extra holds PSP-specific JSON or string map (future channels.extra_json).
	Extra json.RawMessage `json:"extra,omitempty"`
}
