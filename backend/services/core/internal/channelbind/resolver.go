// Package channelbind wires platform channel rows into channeldriver.BindInput and hosts [Hub]
// (routing + driver registry). Gateway/trade should call core via gRPC instead of embedding drivers long term.
package channelbind

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gloopai/pay/common/channelconfig"
	"github.com/gloopai/pay/core/channeldriver"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"gorm.io/gorm"
)

// Resolver implements [channeldriver.ChannelResolver] using DB + optional Consul channel snapshot.
type Resolver struct {
	Ch   *store.ChannelsStore
	Snap *kvcache.ChannelSnapshot
}

// NewResolver builds a resolver; snap may be nil (uses DB channel_config only).
func NewResolver(ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) *Resolver {
	return &Resolver{Ch: ch, Snap: snap}
}

// ResolveBindInput loads one channels row and builds merged JSON for drivers.
func (r *Resolver) ResolveBindInput(ctx context.Context, channelID int64) (channeldriver.BindInput, error) {
	if r == nil || r.Ch == nil {
		return channeldriver.BindInput{}, fmt.Errorf("channelbind: nil resolver")
	}
	if channelID <= 0 {
		return channeldriver.BindInput{}, fmt.Errorf("channelbind: invalid channel_id")
	}
	row, err := r.Ch.AdminGetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return channeldriver.BindInput{}, fmt.Errorf("channelbind: channel %d not found", channelID)
		}
		return channeldriver.BindInput{}, err
	}
	cc := kvcache.PickChannelConfig(r.Snap, channelID, row.ChannelConfig)
	raw, err := channelconfig.ChannelConfigJSONForBind(cc, channelconfig.LegacyChannelFields{
		GatewayURL:        row.GatewayUrl,
		ChannelMerchantNo: row.ChannelMerchantNo,
		SignSecret:        row.SignSecret,
		RSAPrivateKey:     row.RsaPrivateKey,
	}, row.SupportsPayin, row.SupportsPayout)
	if err != nil {
		return channeldriver.BindInput{}, err
	}
	if err := channelconfig.ValidateChannelConfigJSON(raw); err != nil {
		return channeldriver.BindInput{}, err
	}
	return channeldriver.BindInput{
		ChannelID:         channelID,
		DriverKey:         strings.TrimSpace(row.PayinType),
		ChannelConfigJSON: raw,
	}, nil
}
