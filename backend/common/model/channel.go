package model

// Channel is a row from table channels (PSP / routing configuration).
type Channel struct {
	ID                    int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name                  string `json:"name,omitempty" gorm:"column:name"`
	PayinType             string `json:"payin_type,omitempty" gorm:"column:payin_type"`
	GatewayUrl            string `json:"gateway_url,omitempty" gorm:"column:gateway_url"`
	ChannelMerchantNo     string `json:"channel_merchant_no,omitempty" gorm:"column:channel_merchant_no"`
	RsaPrivateKey         string `json:"rsa_private_key,omitempty" gorm:"column:rsa_private_key"`
	SignSecret            string `json:"sign_secret,omitempty" gorm:"column:sign_secret"`
	// ChannelConfig JSON: common keys (gateway_url, channel_merchant_no, sign_secret, rsa_private_key)
	// plus driver_config (PSP-specific JSON; platform may merge legacy columns before passing JSON to drivers).
	ChannelConfig         string `json:"channel_config,omitempty" gorm:"column:channel_config"`
	Weight                int64  `json:"weight,omitempty" gorm:"column:weight"`
	MinAmount             int64  `json:"min_amount,omitempty" gorm:"column:min_amount"`
	MaxAmount             int64  `json:"max_amount,omitempty" gorm:"column:max_amount"`
	SupportsPayin         bool   `json:"supports_payin,omitempty" gorm:"column:supports_payin"`
	SupportsPayout        bool   `json:"supports_payout,omitempty" gorm:"column:supports_payout"`
	ChannelPayinRateBps   int64  `json:"channel_payin_rate_bps,omitempty" gorm:"column:channel_payin_rate_bps"`
	ChannelPayoutRateBps  int64  `json:"channel_payout_rate_bps,omitempty" gorm:"column:channel_payout_rate_bps"`
	ChannelPayoutFeeMode  int64  `json:"channel_payout_fee_mode,omitempty" gorm:"column:channel_payout_fee_mode"`
	ChannelPayoutFixedFee int64  `json:"channel_payout_fixed_fee,omitempty" gorm:"column:channel_payout_fixed_fee"`
	Enabled               bool   `json:"enabled,omitempty" gorm:"column:enabled"`
	FuseEnabled           bool   `json:"fuse_enabled,omitempty" gorm:"column:fuse_enabled"`
}
