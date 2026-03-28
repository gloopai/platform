package model

import "time"

// Order status values for payin_orders.status / payout_orders.status (TINYINT).
const (
	OrderStatusPending int32 = 0
	OrderStatusPaid    int32 = 1
	OrderStatusFailed  int32 = 2
	OrderStatusClosed  int32 = 3
)

// OrderRecord is the shared shape for payin_orders and payout_orders in this codebase.
// Payout store queries alias payout_product_id/code into PayinProductId/PayinProductCode.
// ChannelName is only set when listing with JOIN channels.
type OrderRecord struct {
	OrderNo          string    `json:"order_no,omitempty" gorm:"column:order_no"`
	MerchantId       string    `json:"merchant_id,omitempty" gorm:"column:merchant_id"`
	MerchantOrderNo  string    `json:"merchant_order_no,omitempty" gorm:"column:merchant_order_no"`
	Amount           int64     `json:"amount,omitempty" gorm:"column:amount"`
	Currency         string    `json:"currency,omitempty" gorm:"column:currency"`
	Status           int32     `json:"status,omitempty" gorm:"column:status"`
	ChannelId        int64     `json:"channel_id,omitempty" gorm:"column:channel_id"`
	PayinProductId   int64     `json:"payin_product_id,omitempty" gorm:"column:payin_product_id"`
	PayinProductCode string    `json:"payin_product_code,omitempty" gorm:"column:payin_product_code"`
	ChannelLocked    int32     `json:"channel_locked,omitempty" gorm:"column:channel_locked"`
	ReturnUrl        string    `json:"return_url,omitempty" gorm:"column:return_url"`
	NotifyUrl        string    `json:"notify_url,omitempty" gorm:"column:notify_url"`
	ChannelTradeNo   string    `json:"channel_trade_no,omitempty" gorm:"column:channel_trade_no"`
	PaidAmount       int64     `json:"paid_amount,omitempty" gorm:"column:paid_amount"`
	FeeMode          int64     `json:"fee_mode,omitempty" gorm:"column:fee_mode"`
	FeeRateBps       int64     `json:"fee_rate_bps,omitempty" gorm:"column:fee_rate_bps"`
	FeeFixedAmount   int64     `json:"fee_fixed_amount,omitempty" gorm:"column:fee_fixed_amount"`
	FeeAmount        int64     `json:"fee_amount,omitempty" gorm:"column:fee_amount"`
	NetAmount        int64     `json:"net_amount,omitempty" gorm:"column:net_amount"`
	CreatedAt        time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
	ChannelName      string    `json:"channel_name,omitempty" gorm:"column:channel_name"`
}
