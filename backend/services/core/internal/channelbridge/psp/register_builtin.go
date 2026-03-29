package psp

import (
	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
)

// registerBuiltinDrivers registers built-in driver_key → constructor. Add one r.Register(...) per PSP.
func registerBuiltinDrivers(r *Registry, ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) {
	if r == nil || ch == nil {
		return
	}
	r.Register(hexmeta.DriverKey, func(channelID int64) (contracts.ChannelDriver, error) {
		return hexmeta.NewDriver(channelID, ch, snap)
	})
}
