package psp

import (
	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
)

// RegisterBuiltInDrivers registers all drivers under drivers/. Call once at process startup.
func RegisterBuiltInDrivers(r *Registry, ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) error {
	if r == nil {
		return nil
	}
	r.channels = ch
	r.channelSnap = snap
	r.Register(hexmeta.DriverKey, func(channelID int64) (contracts.ChannelDriver, error) {
		return hexmeta.NewDriver(channelID, ch, snap)
	})
	return nil
}
