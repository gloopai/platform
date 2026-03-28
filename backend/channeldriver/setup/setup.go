package setup

import (
	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/hexmeta"
)

// RegisterDefaultMockPSPs registers built-in PSP drivers (demo / production implementations).
func RegisterDefaultMockPSPs(reg *channeldriver.Registry) error {
	if reg == nil {
		return nil
	}
	reg.Register(hexmeta.DriverKey, func(channelID int64, cfgJSON string) (channeldriver.ChannelDriver, error) {
		return hexmeta.New(channelID, cfgJSON)
	})
	return nil
}
