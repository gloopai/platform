// Package channelbridge bridges platform channel rows (DB + KV) to PSP drivers: BindInput resolution, payin routing, and the driver registry.
package channelbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/core/internal/channelbridge/psp"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	hm "github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Resolver loads channel rows and merged channel_config (KV overrides DB column).
type Resolver struct {
	Ch   *store.ChannelsStore
	Snap *kvcache.ChannelSnapshot
}

func NewResolver(ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) *Resolver {
	return &Resolver{Ch: ch, Snap: snap}
}

func (r *Resolver) ResolveBindInput(ctx context.Context, channelID int64) (contracts.BindInput, error) {
	if r == nil || r.Ch == nil {
		return contracts.BindInput{}, fmt.Errorf("channelbridge: nil resolver")
	}
	if channelID <= 0 {
		return contracts.BindInput{}, fmt.Errorf("channelbridge: invalid channel_id")
	}
	row, err := r.Ch.AdminGetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contracts.BindInput{}, fmt.Errorf("channelbridge: channel %d not found", channelID)
		}
		return contracts.BindInput{}, err
	}
	cc := kvcache.PickChannelConfig(r.Snap, channelID, row.ChannelConfig)
	raw := strings.TrimSpace(cc)
	dk := strings.TrimSpace(row.PayinType)
	if dk != hm.DriverKey {
		if err := validateChannelConfigJSON(raw); err != nil {
			return contracts.BindInput{}, err
		}
	}
	return contracts.BindInput{
		ChannelID:         channelID,
		DriverKey:         dk,
		ChannelConfigJSON: raw,
	}, nil
}

// BridgeConfig wires routing snapshots + registry for one process.
type BridgeConfig struct {
	Channels      *store.ChannelsStore
	PayinProducts *store.PayinProductsStore
	Registry      *psp.Registry
	Resolver      psp.ChannelResolver
	RuntimeConfig *consulx.ConfigStore

	PayinProductSnapshot         *kvcache.PayinProductSnapshot
	PayinProductBindingsSnapshot *kvcache.PayinProductBindingsSnapshot
	ChannelSnapshot              *kvcache.ChannelSnapshot
	MerchantPayinGrantsSnapshot  *kvcache.MerchantPayinGrantsSnapshot
}

// Bridge: OpenAPI payin routing (memory vs DB) + cached drivers.
type Bridge struct {
	cfg BridgeConfig
}

func NewBridge(cfg BridgeConfig) *Bridge { return &Bridge{cfg: cfg} }

func (b *Bridge) MemoryReady() bool {
	if b == nil {
		return false
	}
	c := b.cfg
	if c.RuntimeConfig == nil {
		return false
	}
	return c.PayinProductSnapshot != nil &&
		c.PayinProductBindingsSnapshot != nil &&
		c.ChannelSnapshot != nil &&
		c.MerchantPayinGrantsSnapshot != nil
}

func (b *Bridge) RoutePayin(ctx context.Context, payinProductCode string, amount int64) (channelID, payinProductID int64, err error) {
	if b == nil || b.cfg.Channels == nil {
		return 0, 0, status.Error(codes.FailedPrecondition, "channels not configured")
	}
	if b.MemoryReady() {
		ch, pid, e := kvcache.RoutePayinFromMemory(
			payinProductCode,
			amount,
			b.cfg.PayinProductSnapshot,
			b.cfg.PayinProductBindingsSnapshot,
			b.cfg.ChannelSnapshot,
		)
		if e != nil {
			return 0, 0, status.Error(codes.FailedPrecondition, e.Error())
		}
		return ch, pid, nil
	}
	cid, ppid, err := b.cfg.Channels.Route(ctx, payinProductCode, amount)
	if err != nil {
		return 0, 0, status.Error(codes.FailedPrecondition, err.Error())
	}
	return cid, ppid, nil
}

func (b *Bridge) MerchantPayinProductAllowed(ctx context.Context, merchantID, payinProductCode string) error {
	if b == nil {
		return status.Error(codes.Internal, "channel bridge not configured")
	}
	if b.MemoryReady() {
		ok := kvcache.MerchantHasPayinProductCodeMemory(
			merchantID,
			payinProductCode,
			b.cfg.MerchantPayinGrantsSnapshot,
			b.cfg.PayinProductSnapshot,
		)
		if !ok {
			return status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
		}
		return nil
	}
	if b.cfg.PayinProducts == nil {
		return status.Error(codes.Internal, "payin products store not configured")
	}
	ok, err := b.cfg.PayinProducts.MerchantHasPayinProductCode(ctx, merchantID, payinProductCode)
	if err != nil {
		return status.Error(codes.Internal, "check merchant pay products failed")
	}
	if !ok {
		return status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
	}
	return nil
}

func (b *Bridge) GetDriver(ctx context.Context, channelID int64) (contracts.ChannelDriver, error) {
	if b == nil || b.cfg.Registry == nil {
		return nil, fmt.Errorf("channelbridge: registry not configured")
	}
	if b.cfg.Resolver == nil {
		return nil, fmt.Errorf("channelbridge: channel resolver not configured")
	}
	return b.cfg.Registry.GetChannelDriver(ctx, channelID, b.cfg.Resolver)
}

func (b *Bridge) InvalidateDriverCache(channelID int64) {
	if b == nil || b.cfg.Registry == nil {
		return
	}
	b.cfg.Registry.InvalidateChannelDriver(channelID)
}

func (b *Bridge) OpenPayin(in contracts.BindInput) (contracts.ChannelDriver, error) {
	if b == nil || b.cfg.Registry == nil {
		return nil, contracts.ErrNoDriver
	}
	return b.cfg.Registry.OpenPayin(in)
}

func (b *Bridge) OpenPayout(in contracts.BindInput) (contracts.ChannelDriver, error) {
	if b == nil || b.cfg.Registry == nil {
		return nil, contracts.ErrNoDriver
	}
	return b.cfg.Registry.OpenPayout(in)
}

func validateChannelConfigJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return fmt.Errorf("channel_config must be valid JSON")
	}
	return nil
}
