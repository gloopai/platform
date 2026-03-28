package base

import (
	"context"
	"net/http"
)

// PayoutWayCode matches common PSP payout rails (India doc: 1=bank, 2=UPI).
type PayoutWayCode int

const (
	PayoutWayUnknown PayoutWayCode = iota
	PayoutWayBankCard
	PayoutWayUPI
)

// PayoutUpstream is one registered protocol for payout (代付).
type PayoutUpstream interface {
	Key() string

	CreatePayout(ctx context.Context, cfg *ChannelConfig, req *CreatePayoutReq) (*CreatePayoutResp, error)
	QueryPayout(ctx context.Context, cfg *ChannelConfig, req *QueryPayoutReq) (*QueryPayoutResp, error)

	VerifyPayoutNotify(ctx context.Context, cfg *ChannelConfig, r *http.Request) (*PayoutNotifyParsed, error)
	PayoutNotifyResponse(success bool) []byte
}

// PayoutNotifyContentTyper is optionally implemented for Content-Type on payout notify.
type PayoutNotifyContentTyper interface {
	PayoutNotifyContentType() string
}

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
	UpstreamOrderNo string // sysOrderNo
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
	UpstreamOrderNo string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string // UTR
	RawStatus       string
}

// PayoutNotifyParsed is verified payout callback for platform.
type PayoutNotifyParsed struct {
	MerchantOrderNo string
	UpstreamOrderNo string
	AmountMinor     int64
	Status          PayoutOrderStatus
	ReferenceNo     string // UTR
	RawStatus       string
}
