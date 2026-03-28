package configkv

// Snapshot JSON blobs written under keys from snapshot_keys.go (pay/config/global/.../snapshot/...).

// MerchantKV is the JSON body for merchants/snapshot/{merchant_id}.
// It mirrors merchants except password_hash is never stored in KV.
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

// ChannelKV is the JSON body for channels/snapshot/{id}. Mirrors table channels.
type ChannelKV struct {
	ID                    int64  `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	PayinType             string `json:"payin_type,omitempty"`
	GatewayURL            string `json:"gateway_url,omitempty"`
	ChannelMerchantNo     string `json:"channel_merchant_no,omitempty"`
	RsaPrivateKey         string `json:"rsa_private_key,omitempty"`
	SignSecret            string `json:"sign_secret,omitempty"`
	ChannelConfig         string `json:"channel_config,omitempty"`
	Weight                int64  `json:"weight,omitempty"`
	MinAmount             int64  `json:"min_amount,omitempty"`
	MaxAmount             int64  `json:"max_amount,omitempty"`
	SupportsPayin         bool   `json:"supports_payin,omitempty"`
	SupportsPayout        bool   `json:"supports_payout,omitempty"`
	ChannelPayinRateBps   int64  `json:"channel_payin_rate_bps,omitempty"`
	ChannelPayoutRateBps  int64  `json:"channel_payout_rate_bps,omitempty"`
	ChannelPayoutFeeMode  int64  `json:"channel_payout_fee_mode,omitempty"`
	ChannelPayoutFixedFee int64  `json:"channel_payout_fixed_fee,omitempty"`
	Enabled               bool   `json:"enabled,omitempty"`
	FuseEnabled           bool   `json:"fuse_enabled,omitempty"`
}

// PayinProductKV is the JSON body for payin_products/snapshot/{id}.
type PayinProductKV struct {
	ID            int64  `json:"id,omitempty"`
	Code          string `json:"code,omitempty"`
	Name          string `json:"name,omitempty"`
	SortOrder     int64  `json:"sort_order,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	ProductConfig string `json:"product_config,omitempty"`
}

// PayoutProductKV is the JSON body for payout_products/snapshot/{id}.
type PayoutProductKV struct {
	ID            int64  `json:"id,omitempty"`
	Code          string `json:"code,omitempty"`
	Name          string `json:"name,omitempty"`
	SortOrder     int64  `json:"sort_order,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	ProductConfig string `json:"product_config,omitempty"`
}

// MerchantPayinGrantsKV is the JSON body for merchants/payin_grants/snapshot/{merchant_id}.
type MerchantPayinGrantsKV struct {
	MerchantID string         `json:"merchant_id,omitempty"`
	Grants     []PayinGrantKV `json:"grants"`
}

// PayinGrantKV mirrors enabled merchant_payin_products rows (ordered).
type PayinGrantKV struct {
	PayinProductID  int64  `json:"payin_product_id"`
	MerchantRateBps *int64 `json:"merchant_rate_bps,omitempty"`
}

// MerchantPayoutGrantsKV is the JSON body for merchants/payout_grants/snapshot/{merchant_id}.
type MerchantPayoutGrantsKV struct {
	MerchantID string          `json:"merchant_id,omitempty"`
	Grants     []PayoutGrantKV `json:"grants"`
}

// PayoutGrantKV mirrors enabled merchant_payout_products rows (ordered).
type PayoutGrantKV struct {
	PayoutProductID int64  `json:"payout_product_id"`
	FeeMode         int64  `json:"fee_mode,omitempty"`
	MerchantRateBps *int64 `json:"merchant_rate_bps,omitempty"`
	FeeFixedAmount  int64  `json:"fee_fixed_amount,omitempty"`
}

// PayinProductBindingsKV lists channel bindings for one payin product (for routing / OpenAPI).
type PayinProductBindingsKV struct {
	PayinProductID int64                   `json:"payin_product_id,omitempty"`
	Bindings       []PayinProductChannelKV `json:"bindings"`
}

// PayinProductChannelKV is one row in payin_product_channels.
type PayinProductChannelKV struct {
	ID        int64 `json:"id,omitempty"`
	ChannelID int64 `json:"channel_id"`
	Weight    int64 `json:"weight"`
	Enabled   bool  `json:"enabled"`
}

// PayoutProductBindingsKV lists channel bindings for one payout product.
type PayoutProductBindingsKV struct {
	PayoutProductID int64                    `json:"payout_product_id,omitempty"`
	Bindings        []PayoutProductChannelKV `json:"bindings"`
}

// PayoutProductChannelKV is one row in payout_product_channels.
type PayoutProductChannelKV struct {
	ID        int64 `json:"id,omitempty"`
	ChannelID int64 `json:"channel_id"`
	Weight    int64 `json:"weight"`
	Enabled   bool  `json:"enabled"`
}
