package store

import "github.com/gloopai/pay/common/model"

// ChannelToKV builds a Consul snapshot JSON DTO from a channels row.
func ChannelToKV(c *model.Channel) *model.ChannelKV {
	if c == nil {
		return nil
	}
	return &model.ChannelKV{
		ID:                    c.ID,
		Name:                  c.Name,
		PayinType:             c.PayinType,
		GatewayURL:            c.GatewayUrl,
		ChannelMerchantNo:     c.ChannelMerchantNo,
		RsaPrivateKey:         c.RsaPrivateKey,
		SignSecret:            c.SignSecret,
		ChannelConfig:         c.ChannelConfig,
		Weight:                c.Weight,
		MinAmount:             c.MinAmount,
		MaxAmount:             c.MaxAmount,
		SupportsPayin:         c.SupportsPayin,
		SupportsPayout:        c.SupportsPayout,
		ChannelPayinRateBps:   c.ChannelPayinRateBps,
		ChannelPayoutRateBps:  c.ChannelPayoutRateBps,
		ChannelPayoutFeeMode:  c.ChannelPayoutFeeMode,
		ChannelPayoutFixedFee: c.ChannelPayoutFixedFee,
		Enabled:               c.Enabled,
		FuseEnabled:           c.FuseEnabled,
	}
}

// PayinProductAdminToKV builds a Consul snapshot from a payin product admin row.
func PayinProductAdminToKV(p *model.PayinProductAdmin) *model.PayinProductKV {
	if p == nil {
		return nil
	}
	return &model.PayinProductKV{
		ID:            p.ID,
		Code:          p.Code,
		Name:          p.Name,
		SortOrder:     p.SortOrder,
		Enabled:       p.Enabled,
		ProductConfig: p.ProductConfig,
	}
}

// PayoutProductAdminToKV builds a Consul snapshot from a payout product admin row.
func PayoutProductAdminToKV(p *model.PayoutProductAdmin) *model.PayoutProductKV {
	if p == nil {
		return nil
	}
	return &model.PayoutProductKV{
		ID:            p.ID,
		Code:          p.Code,
		Name:          p.Name,
		SortOrder:     p.SortOrder,
		Enabled:       p.Enabled,
		ProductConfig: p.ProductConfig,
	}
}
