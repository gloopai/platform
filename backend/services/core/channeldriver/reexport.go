// Package channeldriver is the stable import path for PSP drivers and helpers. Implementations live
// under [github.com/gloopai/pay/core/internal/channelbind/channeldriver]; gateway/trade depend on
// this package (not on internal/) so they can share types with core.
package channeldriver

import icd "github.com/gloopai/pay/core/internal/channelbind/channeldriver"

type (
	BindInput          = icd.BindInput
	BaseChannelDriver  = icd.BaseChannelDriver
	ChannelDriver      = icd.ChannelDriver
	CreatePaymentReq   = icd.CreatePaymentReq
	CreatePaymentResp  = icd.CreatePaymentResp
	QueryPaymentReq    = icd.QueryPaymentReq
	QueryPaymentResp   = icd.QueryPaymentResp
	MakeupReq          = icd.MakeupReq
	PayinNotifyParsed  = icd.PayinNotifyParsed
	CreatePayoutReq    = icd.CreatePayoutReq
	CreatePayoutResp   = icd.CreatePayoutResp
	QueryPayoutReq     = icd.QueryPayoutReq
	QueryPayoutResp    = icd.QueryPayoutResp
	PayoutNotifyParsed = icd.PayoutNotifyParsed
	BalanceSnapshot    = icd.BalanceSnapshot
	PayinOrderStatus   = icd.PayinOrderStatus
	PayoutOrderStatus  = icd.PayoutOrderStatus
	PayoutWayCode      = icd.PayoutWayCode
	Registry           = icd.Registry
	ChannelResolver    = icd.ChannelResolver
)

const (
	DefaultChannelNotifyContentType = icd.DefaultChannelNotifyContentType

	PayinStatusUnknown    = icd.PayinStatusUnknown
	PayinStatusProcessing = icd.PayinStatusProcessing
	PayinStatusSuccess    = icd.PayinStatusSuccess
	PayinStatusFailed     = icd.PayinStatusFailed

	PayoutStatusUnknown    = icd.PayoutStatusUnknown
	PayoutStatusProcessing = icd.PayoutStatusProcessing
	PayoutStatusSuccess    = icd.PayoutStatusSuccess
	PayoutStatusFailed     = icd.PayoutStatusFailed

	PayoutWayUnknown  = icd.PayoutWayUnknown
	PayoutWayBankCard = icd.PayoutWayBankCard
	PayoutWayUPI      = icd.PayoutWayUPI
)

var (
	ErrNoDriver     = icd.ErrNoDriver
	ErrUnsupported  = icd.ErrUnsupported
	ErrVerifyNotify = icd.ErrVerifyNotify

	WriteChannelNotify = icd.WriteChannelNotify
	NotifyContentType  = icd.NotifyContentType
)

// NewRegistry delegates to the internal implementation.
func NewRegistry() *Registry { return icd.NewRegistry() }

// RegisterBuiltInDrivers registers default PSP implementations.
func RegisterBuiltInDrivers(r *Registry) error { return icd.RegisterBuiltInDrivers(r) }
