package base

import (
	"context"
	"net/http"
)

// ChannelDriver is the single bound handle for one channels row: identity, normalized payin/payout
// RPCs, and async notify verification. Implementations parse BindInput.ChannelConfigJSON in
// OpenChannel / OpenPayin (same object for both when using a unified factory).
type ChannelDriver interface {
	DriverKey() string
	ChannelID() int64

	CreatePayment(ctx context.Context, req *CreatePaymentReq) (*CreatePaymentResp, error)
	QueryPayment(ctx context.Context, req *QueryPaymentReq) (*QueryPaymentResp, error)
	Makeup(ctx context.Context, req *MakeupReq) error
	VerifyPayinNotify(ctx context.Context, r *http.Request) (*PayinNotifyParsed, error)
	PayinNotifyResponse(success bool) []byte

	CreatePayout(ctx context.Context, req *CreatePayoutReq) (*CreatePayoutResp, error)
	QueryPayout(ctx context.Context, req *QueryPayoutReq) (*QueryPayoutResp, error)
	VerifyPayoutNotify(ctx context.Context, r *http.Request) (*PayoutNotifyParsed, error)
	PayoutNotifyResponse(success bool) []byte

	QueryBalance(ctx context.Context) (*BalanceSnapshot, error)
}

type BaseChannelDriver struct {
	DriverKey string
	ChannelID int64
}

// NewBaseChannelDriver builds identity for embedding; override methods on the outer type.
func NewBaseChannelDriver(channelID int64, driverKey string) BaseChannelDriver {
	return BaseChannelDriver{DriverKey: driverKey, ChannelID: channelID}
}

func (*BaseChannelDriver) CreatePayment(context.Context, *CreatePaymentReq) (*CreatePaymentResp, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) QueryPayment(context.Context, *QueryPaymentReq) (*QueryPaymentResp, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) Makeup(context.Context, *MakeupReq) error {
	return ErrUnsupported
}

func (*BaseChannelDriver) VerifyPayinNotify(context.Context, *http.Request) (*PayinNotifyParsed, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) PayinNotifyResponse(success bool) []byte {
	if success {
		return []byte("SUCCESS")
	}
	return []byte("FAIL")
}

func (*BaseChannelDriver) CreatePayout(context.Context, *CreatePayoutReq) (*CreatePayoutResp, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) QueryPayout(context.Context, *QueryPayoutReq) (*QueryPayoutResp, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) VerifyPayoutNotify(context.Context, *http.Request) (*PayoutNotifyParsed, error) {
	return nil, ErrUnsupported
}

func (*BaseChannelDriver) PayoutNotifyResponse(success bool) []byte {
	if success {
		return []byte("SUCCESS")
	}
	return []byte("FAIL")
}

func (*BaseChannelDriver) QueryBalance(context.Context) (*BalanceSnapshot, error) {
	return nil, ErrUnsupported
}
