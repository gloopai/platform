package channeldriver

import (
	"context"
	"net/http"

	"github.com/gloopai/pay/channeldriver/base"
)

// Re-exports of github.com/gloopai/pay/channeldriver/base for stable import paths on trade/gateway.

type (
	ChannelConfig       = base.ChannelConfig
	CreatePaymentReq    = base.CreatePaymentReq
	CreatePaymentResp   = base.CreatePaymentResp
	QueryPaymentReq     = base.QueryPaymentReq
	QueryPaymentResp    = base.QueryPaymentResp
	MakeupReq           = base.MakeupReq
	PayinNotifyParsed   = base.PayinNotifyParsed
	PayinChannel       = base.PayinChannel
	PayoutChannel      = base.PayoutChannel
	BalanceChannel     = base.BalanceChannel
	CreatePayoutReq     = base.CreatePayoutReq
	CreatePayoutResp    = base.CreatePayoutResp
	QueryPayoutReq      = base.QueryPayoutReq
	QueryPayoutResp     = base.QueryPayoutResp
	PayoutNotifyParsed  = base.PayoutNotifyParsed
	BalanceSnapshot     = base.BalanceSnapshot
	Registry            = base.Registry
	PayinNotifyRoute    = base.PayinNotifyRoute
	PayoutNotifyRoute   = base.PayoutNotifyRoute
	PayinOrderStatus    = base.PayinOrderStatus
	PayoutOrderStatus   = base.PayoutOrderStatus
	PayoutWayCode       = base.PayoutWayCode
)

const (
	DefaultChannelNotifyContentType = base.DefaultChannelNotifyContentType

	PayinStatusUnknown    = base.PayinStatusUnknown
	PayinStatusProcessing = base.PayinStatusProcessing
	PayinStatusSuccess    = base.PayinStatusSuccess
	PayinStatusFailed     = base.PayinStatusFailed

	PayoutStatusUnknown    = base.PayoutStatusUnknown
	PayoutStatusProcessing = base.PayoutStatusProcessing
	PayoutStatusSuccess    = base.PayoutStatusSuccess
	PayoutStatusFailed     = base.PayoutStatusFailed

	PayoutWayUnknown  = base.PayoutWayUnknown
	PayoutWayBankCard = base.PayoutWayBankCard
	PayoutWayUPI      = base.PayoutWayUPI
)

var (
	ErrNoDriver     = base.ErrNoDriver
	ErrUnsupported  = base.ErrUnsupported
	ErrVerifyNotify = base.ErrVerifyNotify
)

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return base.NewRegistry() }

// ConfigFromDriverKey builds a ChannelConfig using DB fields. DriverKey is typically channels.payin_type.
func ConfigFromDriverKey(channelID int64, driverKey, gatewayBaseURL, appID, signSecret string, rsaPEM string, payin, payout bool) *ChannelConfig {
	return base.ConfigFromDriverKey(channelID, driverKey, gatewayBaseURL, appID, signSecret, rsaPEM, payin, payout)
}

// ConfigFieldsFromChannelJSON extracts common keys from channels.channel_config JSON.
func ConfigFieldsFromChannelJSON(raw string) (gatewayURL, merchantNo, signSecret, rsaPEM string) {
	return base.ConfigFieldsFromChannelJSON(raw)
}

// NotifyContentType returns the HTTP Content-Type for the response body returned to the PSP.
func NotifyContentType(drv any) string { return base.NotifyContentType(drv) }

// WriteChannelNotify writes body and Content-Type for a successful notify handling path.
func WriteChannelNotify(w http.ResponseWriter, drv any, body []byte) {
	base.WriteChannelNotify(w, drv, body)
}

// HandlePayinNotify is a helper: resolve config, load driver, verify, return body bytes for the PSP.
func HandlePayinNotify(ctx context.Context, reg *Registry, route PayinNotifyRoute, r *http.Request,
	onSuccess func(*PayinNotifyParsed) (ok bool, err error),
) (body []byte, drv PayinChannel, err error) {
	return base.HandlePayinNotify(ctx, reg, route, r, onSuccess)
}

// HandlePayoutNotify is the payout analogue of HandlePayinNotify.
func HandlePayoutNotify(ctx context.Context, reg *Registry, route PayoutNotifyRoute, r *http.Request,
	onSuccess func(*PayoutNotifyParsed) (ok bool, err error),
) (body []byte, drv PayoutChannel, err error) {
	return base.HandlePayoutNotify(ctx, reg, route, r, onSuccess)
}
