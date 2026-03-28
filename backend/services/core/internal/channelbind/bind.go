// Package channelbind: DB+KV → BindInput, payin routing, and PSP driver registry.
package channelbind

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/core/internal/channelbind/psp"
	"github.com/gloopai/pay/core/internal/channelbind/psp/contracts"
	hm "github.com/gloopai/pay/core/internal/channelbind/psp/drivers/hexmeta"
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
		return contracts.BindInput{}, fmt.Errorf("channelbind: nil resolver")
	}
	if channelID <= 0 {
		return contracts.BindInput{}, fmt.Errorf("channelbind: invalid channel_id")
	}
	row, err := r.Ch.AdminGetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contracts.BindInput{}, fmt.Errorf("channelbind: channel %d not found", channelID)
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

// HubConfig wires routing snapshots + registry for one process.
type HubConfig struct {
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

// Hub: OpenAPI payin routing (memory vs DB) + cached drivers.
type Hub struct {
	cfg HubConfig
}

func NewHub(cfg HubConfig) *Hub { return &Hub{cfg: cfg} }

func (h *Hub) MemoryReady() bool {
	if h == nil {
		return false
	}
	c := h.cfg
	if c.RuntimeConfig == nil {
		return false
	}
	return c.PayinProductSnapshot != nil &&
		c.PayinProductBindingsSnapshot != nil &&
		c.ChannelSnapshot != nil &&
		c.MerchantPayinGrantsSnapshot != nil
}

func (h *Hub) RoutePayin(ctx context.Context, payinProductCode string, amount int64) (channelID, payinProductID int64, err error) {
	if h == nil || h.cfg.Channels == nil {
		return 0, 0, status.Error(codes.FailedPrecondition, "channels not configured")
	}
	if h.MemoryReady() {
		ch, pid, e := kvcache.RoutePayinFromMemory(
			payinProductCode,
			amount,
			h.cfg.PayinProductSnapshot,
			h.cfg.PayinProductBindingsSnapshot,
			h.cfg.ChannelSnapshot,
		)
		if e != nil {
			return 0, 0, status.Error(codes.FailedPrecondition, e.Error())
		}
		return ch, pid, nil
	}
	cid, ppid, err := h.cfg.Channels.Route(ctx, payinProductCode, amount)
	if err != nil {
		return 0, 0, status.Error(codes.FailedPrecondition, err.Error())
	}
	return cid, ppid, nil
}

func (h *Hub) MerchantPayinProductAllowed(ctx context.Context, merchantID, payinProductCode string) error {
	if h == nil {
		return status.Error(codes.Internal, "channel hub not configured")
	}
	if h.MemoryReady() {
		ok := kvcache.MerchantHasPayinProductCodeMemory(
			merchantID,
			payinProductCode,
			h.cfg.MerchantPayinGrantsSnapshot,
			h.cfg.PayinProductSnapshot,
		)
		if !ok {
			return status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
		}
		return nil
	}
	if h.cfg.PayinProducts == nil {
		return status.Error(codes.Internal, "payin products store not configured")
	}
	ok, err := h.cfg.PayinProducts.MerchantHasPayinProductCode(ctx, merchantID, payinProductCode)
	if err != nil {
		return status.Error(codes.Internal, "check merchant pay products failed")
	}
	if !ok {
		return status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
	}
	return nil
}

func (h *Hub) GetDriver(ctx context.Context, channelID int64) (contracts.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, fmt.Errorf("channelbind: registry not configured")
	}
	if h.cfg.Resolver == nil {
		return nil, fmt.Errorf("channelbind: channel resolver not configured")
	}
	return h.cfg.Registry.GetChannelDriver(ctx, channelID, h.cfg.Resolver)
}

func (h *Hub) InvalidateDriverCache(channelID int64) {
	if h == nil || h.cfg.Registry == nil {
		return
	}
	h.cfg.Registry.InvalidateChannelDriver(channelID)
}

func (h *Hub) OpenPayin(in contracts.BindInput) (contracts.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, contracts.ErrNoDriver
	}
	return h.cfg.Registry.OpenPayin(in)
}

func (h *Hub) OpenPayout(in contracts.BindInput) (contracts.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, contracts.ErrNoDriver
	}
	return h.cfg.Registry.OpenPayout(in)
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
