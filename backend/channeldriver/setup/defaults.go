// Package setup provides optional wiring helpers (e.g. dev mock PSP drivers) without import cycles on the root channeldriver package.
package setup

import (
	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/mockpsp"
	"github.com/gloopai/pay/channeldriver/mockpsp2"
)

// RegisterDefaultMockPSPs registers mock_psp and mock_psp2 drivers (gateway / core / trade dev wiring).
func RegisterDefaultMockPSPs(reg *channeldriver.Registry) error {
	if err := mockpsp.RegisterAll(reg, mockpsp.New(mockpsp.DefaultDriverKey)); err != nil {
		return err
	}
	return mockpsp2.RegisterAll(reg, mockpsp2.New(mockpsp2.DefaultDriverKey))
}
