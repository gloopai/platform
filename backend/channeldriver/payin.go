package channeldriver

import (
	"context"
	"net/http"
)

// PayinUpstream is one registered protocol for collection (代收). Implementations must be
// safe for concurrent use; each call receives the ChannelConfig for the selected channel row.
type PayinUpstream interface {
	// Key returns the driver_key this implementation handles (e.g. "psp_india_a").
	Key() string

	// CreatePayment calls upstream create order API (e.g. POST .../order/payment).
	CreatePayment(ctx context.Context, cfg *ChannelConfig, req *CreatePaymentReq) (*CreatePaymentResp, error)

	// QueryPayment calls upstream query API (e.g. POST .../query/payment).
	QueryPayment(ctx context.Context, cfg *ChannelConfig, req *QueryPaymentReq) (*QueryPaymentResp, error)

	// Makeup requests manual reconciliation with UTR (e.g. POST .../makeup). May return ErrUnsupported.
	Makeup(ctx context.Context, cfg *ChannelConfig, req *MakeupReq) error

	// VerifyPayinNotify validates signature and parses body for an async payin callback.
	VerifyPayinNotify(ctx context.Context, cfg *ChannelConfig, r *http.Request) (*PayinNotifyParsed, error)

	// PayinNotifyResponse is the raw HTTP body to return to the PSP (e.g. SUCCESS / FAIL bytes).
	PayinNotifyResponse(success bool) []byte
}

// PayinNotifyContentTyper is optionally implemented to set Content-Type on the notify response.
type PayinNotifyContentTyper interface {
	PayinNotifyContentType() string
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
	UpstreamOrderNo string // sysOrderNo
	PayURL          string // payUrl / cashier
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
	AppID            string
	MerchantOrderNo  string
	UpstreamOrderNo  string
	AmountMinor      int64
	Status           PayinOrderStatus
	ReferenceNo      string // UTR when present
	FailReason       string
	RawStatus        string // upstream status string if needed for audit
}

// MakeupReq is UTR-based makeup / reconciliation.
type MakeupReq struct {
	MerchantOrderNo string
	ReferenceNo     string // UTR
}

// PayinNotifyParsed is the verified, normalized callback for platform order updates.
type PayinNotifyParsed struct {
	MerchantOrderNo string
	UpstreamOrderNo string
	PaidAmountMinor int64
	Status          PayinOrderStatus
	RawStatus       string
}
