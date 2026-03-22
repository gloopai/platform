package logic

import (
	"github.com/gloopai/pay/merchant/internal/store"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
)

func toMerchantInfo(m *store.Merchant) *merchantpb.MerchantInfo {
	if m == nil {
		return nil
	}
	return &merchantpb.MerchantInfo{
		MerchantId:      m.MerchantId,
		ApiSecret:       m.ApiSecret,
		Status:          m.Status,
		RateBps:         m.RateBps,
		IpWhitelist:     m.IpWhitelist,
		Balance:         m.Balance,
		FrozenBalance:   m.FrozenBalance,
		WithdrawnAmount: m.WithdrawnAmount,
		NotifyUrl:       m.NotifyUrl,
		ReturnUrl:       m.ReturnUrl,
	}
}
