package store

import (
	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/model"
)

// ChannelToKV builds a Consul snapshot JSON DTO from a channels row.
func ChannelToKV(c *model.Channel) *configkv.ChannelKV {
	if c == nil {
		return nil
	}
	return &configkv.ChannelKV{
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

// KVToChannel maps a Consul channel snapshot back to model.Channel.
func KVToChannel(kv *configkv.ChannelKV) *model.Channel {
	if kv == nil {
		return nil
	}
	return &model.Channel{
		ID:                    kv.ID,
		Name:                  kv.Name,
		PayinType:             kv.PayinType,
		GatewayUrl:            kv.GatewayURL,
		ChannelMerchantNo:     kv.ChannelMerchantNo,
		RsaPrivateKey:         kv.RsaPrivateKey,
		SignSecret:            kv.SignSecret,
		ChannelConfig:         kv.ChannelConfig,
		Weight:                kv.Weight,
		MinAmount:             kv.MinAmount,
		MaxAmount:             kv.MaxAmount,
		SupportsPayin:         kv.SupportsPayin,
		SupportsPayout:        kv.SupportsPayout,
		ChannelPayinRateBps:   kv.ChannelPayinRateBps,
		ChannelPayoutRateBps:  kv.ChannelPayoutRateBps,
		ChannelPayoutFeeMode:  kv.ChannelPayoutFeeMode,
		ChannelPayoutFixedFee: kv.ChannelPayoutFixedFee,
		Enabled:               kv.Enabled,
		FuseEnabled:           kv.FuseEnabled,
	}
}

// PayinProductAdminToKV builds a Consul snapshot from a payin product admin row.
func PayinProductAdminToKV(p *model.PayinProductAdmin) *configkv.PayinProductKV {
	if p == nil {
		return nil
	}
	return &configkv.PayinProductKV{
		ID:            p.ID,
		Code:          p.Code,
		Name:          p.Name,
		SortOrder:     p.SortOrder,
		Enabled:       p.Enabled,
		ProductConfig: p.ProductConfig,
	}
}

// PayoutProductAdminToKV builds a Consul snapshot from a payout product admin row.
func PayoutProductAdminToKV(p *model.PayoutProductAdmin) *configkv.PayoutProductKV {
	if p == nil {
		return nil
	}
	return &configkv.PayoutProductKV{
		ID:            p.ID,
		Code:          p.Code,
		Name:          p.Name,
		SortOrder:     p.SortOrder,
		Enabled:       p.Enabled,
		ProductConfig: p.ProductConfig,
	}
}
