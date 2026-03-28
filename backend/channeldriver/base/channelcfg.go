package base

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ConfigFieldsFromUpstreamJSON extracts common string fields from channels.upstream_config JSON object.
// Recognized keys: gateway_url, upstream_merchant_no, sign_secret, rsa_private_key.
// Non-object JSON returns empty fields.
func ConfigFieldsFromUpstreamJSON(raw string) (gatewayURL, merchantNo, signSecret, rsaPEM string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return
	}
	return coerceString(m["gateway_url"]), coerceString(m["upstream_merchant_no"]),
		coerceString(m["sign_secret"]), coerceString(m["rsa_private_key"])
}

func coerceString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

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
