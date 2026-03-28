package model

// PayinGrant describes one enabled merchant_payin_products row (optional merchant_rate_bps override).
type PayinGrant struct {
	PayinProductID int64  `json:"payin_product_id,omitempty" gorm:"column:payin_product_id"`
	RateBps        *int64 `json:"merchant_rate_bps,omitempty" gorm:"column:merchant_rate_bps"`
}

// PayoutGrant describes one enabled merchant_payout_products row.
type PayoutGrant struct {
	PayoutProductID int64  `json:"payout_product_id,omitempty" gorm:"column:payout_product_id"`
	FeeMode         int64  `json:"fee_mode,omitempty" gorm:"column:fee_mode"`
	RateBps         *int64 `json:"merchant_rate_bps,omitempty" gorm:"column:merchant_rate_bps"`
	FixedFeeAmount  int64  `json:"fee_fixed_amount,omitempty" gorm:"column:fee_fixed_amount"`
}
