package channeldriver

import "github.com/gloopai/pay/core/internal/channelbind/channeldriver/hexmeta"

// RegisterBuiltInDrivers registers default PSP implementations (e.g. hexmeta). Call once at process startup.
func RegisterBuiltInDrivers(r *Registry) error {
	if r == nil {
		return nil
	}
	r.Register(hexmeta.DriverKey, func(channelID int64, cfgJSON string) (ChannelDriver, error) {
		return hexmeta.New(channelID, cfgJSON)
	})
	return nil
}
