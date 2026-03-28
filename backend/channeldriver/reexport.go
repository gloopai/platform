package channeldriver

import (
	"github.com/gloopai/pay/channeldriver/base"
)

// Re-exports of github.com/gloopai/pay/channeldriver/base. Application wiring of
// [Registry] / [ChannelResolver] belongs in the core service; other binaries should prefer
// channel gRPC to core (gateway/trade may still import types/helpers during migration).

type (
	BindInput          = base.BindInput
	BaseChannelDriver  = base.BaseChannelDriver
	ChannelDriver      = base.ChannelDriver
	CreatePaymentReq   = base.CreatePaymentReq
	CreatePaymentResp  = base.CreatePaymentResp
	QueryPaymentReq    = base.QueryPaymentReq
	QueryPaymentResp   = base.QueryPaymentResp
	MakeupReq          = base.MakeupReq
	PayinNotifyParsed  = base.PayinNotifyParsed
	CreatePayoutReq    = base.CreatePayoutReq
	CreatePayoutResp   = base.CreatePayoutResp
	QueryPayoutReq     = base.QueryPayoutReq
	QueryPayoutResp    = base.QueryPayoutResp
	PayoutNotifyParsed = base.PayoutNotifyParsed
	BalanceSnapshot    = base.BalanceSnapshot
	PayinOrderStatus   = base.PayinOrderStatus
	PayoutOrderStatus  = base.PayoutOrderStatus
	PayoutWayCode      = base.PayoutWayCode
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

	WriteChannelNotify = base.WriteChannelNotify
	NotifyContentType  = base.NotifyContentType
)
