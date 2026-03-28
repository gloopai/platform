package mockpsp

import "github.com/gloopai/pay/channeldriver"

// RegisterAll registers this driver as payin, payout, and balance for the same driver key.
func RegisterAll(reg *channeldriver.Registry, d *Driver) error {
	if d == nil {
		return nil
	}
	if err := reg.RegisterPayin(d); err != nil {
		return err
	}
	if err := reg.RegisterPayout(d); err != nil {
		return err
	}
	if err := reg.RegisterBalance(d); err != nil {
		return err
	}
	return nil
}
