package base

// BindInput is the platform payload for opening a bound channel: stable row identity plus the full
// channel_config JSON (shape is PSP-specific; drivers parse it in OpenPayin / OpenPayout / OpenBalance).
type BindInput struct {
	ChannelID int64
	DriverKey string

	ChannelConfigJSON string
}

// CreatePaymentReq maps to upstream create collection order parameters (normalized).
type CreatePaymentReq struct {
	MerchantOrderNo string // maps to orderNo
	AmountMinor     int64  // amount in minor units (e.g. paise / fen)
	PayerName       string
	PayerPhone      string
	PayerEmail      string
	UserIP          string
	NotifyURL       string // platform URL the PSP will call
}

// CreatePaymentResp is upstream create response mapped to platform fields.
type CreatePaymentResp struct {
	ChannelOrderNo string // sysOrderNo
	PayURL         string // payUrl / cashier
}

// QueryPaymentReq queries by merchant order number.
type QueryPaymentReq struct {
	MerchantOrderNo string
}

// PayinOrderStatus is normalized from upstream status strings (e.g. "1","2","3").
type PayinOrderStatus int

const (
	PayinStatusUnknown PayinOrderStatus = iota
	PayinStatusProcessing
	PayinStatusSuccess
	PayinStatusFailed
)

// QueryPaymentResp is normalized query result.
type QueryPaymentResp struct {
	AppID           string
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayinOrderStatus
	ReferenceNo     string // UTR when present
	FailReason      string
	RawStatus       string // upstream status string if needed for audit
}

// MakeupReq is UTR-based makeup / reconciliation.
type MakeupReq struct {
	MerchantOrderNo string
	ReferenceNo     string // UTR
}

// PayinNotifyParsed is the verified, normalized callback for platform order updates.
type PayinNotifyParsed struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	PaidAmountMinor int64
	Status          PayinOrderStatus
	RawStatus       string
}

// PayoutWayCode matches common PSP payout rails (e.g. India: bank vs UPI).
type PayoutWayCode int

const (
	PayoutWayUnknown PayoutWayCode = iota
	PayoutWayBankCard
	PayoutWayUPI
)

// CreatePayoutReq maps upstream payout create request (normalized).
type CreatePayoutReq struct {
	MerchantOrderNo string
	AmountMinor     int64
	WayCode         PayoutWayCode
	BankName        string
	BankCode        string // IFSC when WayCode is bank
	AccountNo       string
	HolderName      string
	Phone           string
	Email           string
	NotifyURL       string
}

// CreatePayoutResp is upstream response after create payout.
type CreatePayoutResp struct {
	ChannelOrderNo string // sysOrderNo
}

// QueryPayoutReq queries by merchant order number.
type QueryPayoutReq struct {
	MerchantOrderNo string
}

// PayoutOrderStatus normalized.
type PayoutOrderStatus int

const (
	PayoutStatusUnknown PayoutOrderStatus = iota
	PayoutStatusProcessing
	PayoutStatusSuccess
	PayoutStatusFailed
)

// QueryPayoutResp is normalized payout query result.
type QueryPayoutResp struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string // UTR
	RawStatus       string
}

// PayoutNotifyParsed is verified payout callback for platform.
type PayoutNotifyParsed struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string // UTR
	RawStatus       string
}

// BalanceSnapshot maps PSP balance fields (amounts in minor units).
type BalanceSnapshot struct {
	AvailableMinor int64 // available for payout
	UnsettledMinor int64 // collected but not yet available for payout
	FrozenMinor    int64 // in-flight payout
}
