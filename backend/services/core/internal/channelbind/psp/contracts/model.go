// Package contracts: shared types and enums for all PSP drivers (no registry, no implementations).
package contracts

// BindInput is the platform payload for opening a bound channel (channel_config JSON is PSP-specific).
type BindInput struct {
	ChannelID         int64
	DriverKey         string
	ChannelConfigJSON string
}

type CreatePaymentReq struct {
	MerchantOrderNo string
	AmountMinor     int64
	PayerName       string
	PayerPhone      string
	PayerEmail      string
	UserIP          string
	NotifyURL       string
}

type CreatePaymentResp struct {
	ChannelOrderNo string
	PayURL         string
}

type QueryPaymentReq struct {
	MerchantOrderNo string
}

type PayinOrderStatus int

const (
	PayinStatusUnknown PayinOrderStatus = iota
	PayinStatusProcessing
	PayinStatusSuccess
	PayinStatusFailed
)

type QueryPaymentResp struct {
	AppID           string
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayinOrderStatus
	ReferenceNo     string
	FailReason      string
	RawStatus       string
}

type MakeupReq struct {
	MerchantOrderNo string
	ReferenceNo     string
}

type PayinNotifyParsed struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	PaidAmountMinor int64
	Status          PayinOrderStatus
	RawStatus       string
}

type PayoutWayCode int

const (
	PayoutWayUnknown PayoutWayCode = iota
	PayoutWayBankCard
	PayoutWayUPI
)

type CreatePayoutReq struct {
	MerchantOrderNo string
	AmountMinor     int64
	WayCode         PayoutWayCode
	BankName        string
	BankCode        string
	AccountNo       string
	HolderName      string
	Phone           string
	Email           string
	NotifyURL       string
}

type CreatePayoutResp struct {
	ChannelOrderNo string
}

type QueryPayoutReq struct {
	MerchantOrderNo string
}

type PayoutOrderStatus int

const (
	PayoutStatusUnknown PayoutOrderStatus = iota
	PayoutStatusProcessing
	PayoutStatusSuccess
	PayoutStatusFailed
)

type QueryPayoutResp struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string
	RawStatus       string
}

type PayoutNotifyParsed struct {
	MerchantOrderNo string
	ChannelOrderNo  string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string
	RawStatus       string
}

type BalanceSnapshot struct {
	AvailableMinor int64
	UnsettledMinor int64
	FrozenMinor    int64
}
