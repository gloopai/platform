package model

// MerchantKV is the JSON body for Consul key pay/config/global/merchants/snapshot/{merchant_id}.
// It mirrors merchants except password_hash is never stored in KV.
// Used by core kvcache and sync after Create/Update merchant.
type MerchantKV struct {
	ID               int64  `json:"id,omitempty"`
	MerchantID       string `json:"merchant_id,omitempty"`
	AppID            string `json:"app_id,omitempty"`
	Email            string `json:"email,omitempty"`
	AppSecret        string `json:"app_secret,omitempty"`
	Status           int64  `json:"status,omitempty"`
	IpWhitelist      string `json:"ip_whitelist,omitempty"`
	PayinBalance     int64  `json:"payin_balance,omitempty"`
	AvailableBalance int64  `json:"available_balance,omitempty"`
	FrozenBalance    int64  `json:"frozen_balance,omitempty"`
	WithdrawnAmount  int64  `json:"withdrawn_amount,omitempty"`
	NotifyURL        string `json:"notify_url,omitempty"`
	ReturnURL        string `json:"return_url,omitempty"`
	MerchantConfig   string `json:"merchant_config,omitempty"`
}
