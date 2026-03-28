package model

// ChannelKV is the JSON body for Consul key pay/config/global/channels/snapshot/{id}.
// Mirrors table channels; used by trade kvcache and sync after Create/Update channel.
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
