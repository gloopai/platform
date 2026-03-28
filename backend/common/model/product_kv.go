package model

// PayinProductKV is the JSON body for Consul key pay/config/global/payin_products/snapshot/{id}.
type PayinProductKV struct {
	ID            int64  `json:"id,omitempty"`
	Code          string `json:"code,omitempty"`
	Name          string `json:"name,omitempty"`
	SortOrder     int64  `json:"sort_order,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	ProductConfig string `json:"product_config,omitempty"`
}

// PayoutProductKV is the JSON body for Consul key pay/config/global/payout_products/snapshot/{id}.
type PayoutProductKV struct {
	ID            int64  `json:"id,omitempty"`
	Code          string `json:"code,omitempty"`
	Name          string `json:"name,omitempty"`
	SortOrder     int64  `json:"sort_order,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	ProductConfig string `json:"product_config,omitempty"`
}
