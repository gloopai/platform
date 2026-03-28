// Package channelbridge: OpenAPI payin routing (memory vs DB). PSP driver cache lives in psp.Registry (see svc.ServiceContext.DriverRegistry).
package channelbridge

import (
	"context"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BridgeConfig wires routing snapshots for one process.
type BridgeConfig struct {
	Channels      *store.ChannelsStore
	PayinProducts *store.PayinProductsStore
	RuntimeConfig *consulx.ConfigStore

	PayinProductSnapshot         *kvcache.PayinProductSnapshot
	PayinProductBindingsSnapshot *kvcache.PayinProductBindingsSnapshot
	ChannelSnapshot              *kvcache.ChannelSnapshot
	MerchantPayinGrantsSnapshot  *kvcache.MerchantPayinGrantsSnapshot
}

// Bridge: OpenAPI payin routing (memory vs DB).
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
