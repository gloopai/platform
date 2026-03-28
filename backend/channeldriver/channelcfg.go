package channeldriver

import "strings"

// ConfigFromDriverKey builds a ChannelConfig using DB fields. DriverKey is typically channels.payin_type.
func ConfigFromDriverKey(channelID int64, driverKey, gatewayBaseURL, appID, signSecret string, rsaPEM string, payin, payout bool) *ChannelConfig {
	return &ChannelConfig{
		ChannelID:        channelID,
		DriverKey:        strings.TrimSpace(driverKey),
		GatewayBaseURL:   strings.TrimSpace(gatewayBaseURL),
		AppID:            strings.TrimSpace(appID),
		SignSecret:       signSecret,
		RSAPrivateKeyPEM: rsaPEM,
		SupportsPayin:    payin,
		SupportsPayout:   payout,
	}
}
