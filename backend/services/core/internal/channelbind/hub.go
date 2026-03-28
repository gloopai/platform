// Hub centralizes payin routing (KV vs DB) and [channeldriver.ChannelDriver] access for core.
package channelbind

import (
	"context"
	"fmt"

	"github.com/gloopai/pay/core/channeldriver"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HubConfig wires dependencies for [Hub]. Snapshot fields may be nil when Consul is not used.
type HubConfig struct {
	Channels *store.ChannelsStore
	// PayinProducts is used when memory routing is not ready (merchant product allowlist DB path).
	PayinProducts *store.PayinProductsStore

	Registry *channeldriver.Registry
	Resolver channeldriver.ChannelResolver

	RuntimeConfig *consulx.ConfigStore

	PayinProductSnapshot         *kvcache.PayinProductSnapshot
	PayinProductBindingsSnapshot *kvcache.PayinProductBindingsSnapshot
	ChannelSnapshot              *kvcache.ChannelSnapshot
	MerchantPayinGrantsSnapshot  *kvcache.MerchantPayinGrantsSnapshot
}

// Hub combines payin routing and channel driver resolution for one core process.
type Hub struct {
	cfg HubConfig
}

// NewHub returns a hub; cfg.Registry and cfg.Resolver must be non-nil for GetDriver.
func NewHub(cfg HubConfig) *Hub {
	return &Hub{cfg: cfg}
}

// MemoryReady matches OpenAPI hot-path: Consul routing snapshots are wired.
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

// RoutePayin selects channel_id and payin_product_id by product code (payin_type) and amount.
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
	channelID, payProductID, err := h.cfg.Channels.Route(ctx, payinProductCode, amount)
	if err != nil {
		return 0, 0, status.Error(codes.FailedPrecondition, err.Error())
	}
	return channelID, payProductID, nil
}

// MerchantPayinProductAllowed returns nil if the merchant may use this payin product code (OpenAPI payin_type).
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

// GetDriver returns a cached [channeldriver.ChannelDriver] for channel_id with merged channel_config (KV + DB).
func (h *Hub) GetDriver(ctx context.Context, channelID int64) (channeldriver.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, fmt.Errorf("channelbind: registry not configured")
	}
	if h.cfg.Resolver == nil {
		return nil, fmt.Errorf("channelbind: channel resolver not configured")
	}
	return h.cfg.Registry.GetChannelDriver(ctx, channelID, h.cfg.Resolver)
}

// InvalidateDriverCache drops the in-process driver after channel_config / KV changes.
func (h *Hub) InvalidateDriverCache(channelID int64) {
	if h == nil || h.cfg.Registry == nil {
		return
	}
	h.cfg.Registry.InvalidateChannelDriver(channelID)
}

// OpenPayin builds a driver from an explicit [channeldriver.BindInput].
func (h *Hub) OpenPayin(in channeldriver.BindInput) (channeldriver.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, channeldriver.ErrNoDriver
	}
	return h.cfg.Registry.OpenPayin(in)
}

// OpenPayout builds a payout driver from [channeldriver.BindInput].
func (h *Hub) OpenPayout(in channeldriver.BindInput) (channeldriver.ChannelDriver, error) {
	if h == nil || h.cfg.Registry == nil {
		return nil, channeldriver.ErrNoDriver
	}
	return h.cfg.Registry.OpenPayout(in)
}
