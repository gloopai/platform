package base

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ConfigFieldsFromChannelJSON extracts common string fields from channels.channel_config JSON object.
// Recognized keys: gateway_url, channel_merchant_no (legacy: upstream_merchant_no), sign_secret, rsa_private_key.
// Non-object JSON returns empty fields.
func ConfigFieldsFromChannelJSON(raw string) (gatewayURL, merchantNo, signSecret, rsaPEM string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return
	}
	mer := coerceString(m["channel_merchant_no"])
	if mer == "" {
		mer = coerceString(m["upstream_merchant_no"])
	}
	return coerceString(m["gateway_url"]), mer,
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

// LegacyChannelFields holds split DB columns when channels.channel_config was empty (legacy layout).
type LegacyChannelFields struct {
	GatewayURL        string
	ChannelMerchantNo string
	SignSecret        string
	RSAPrivateKey     string
}

// ChannelConfigJSONForAPI returns the channel_config blob for admin APIs: prefer the stored JSON;
// if empty, synthesize a JSON object from legacy split columns (migration / old admin).
func ChannelConfigJSONForAPI(channelConfig string, leg LegacyChannelFields) string {
	channelConfig = strings.TrimSpace(channelConfig)
	if channelConfig != "" {
		return channelConfig
	}
	legMap := map[string]string{
		"gateway_url":         strings.TrimSpace(leg.GatewayURL),
		"channel_merchant_no": strings.TrimSpace(leg.ChannelMerchantNo),
		"sign_secret":         strings.TrimSpace(leg.SignSecret),
		"rsa_private_key":     strings.TrimSpace(leg.RSAPrivateKey),
	}
	b, err := json.Marshal(legMap)
	if err != nil {
		return ""
	}
	return string(b)
}

// ValidateChannelConfigJSON returns an error if s is non-empty and not valid JSON.
func ValidateChannelConfigJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return fmt.Errorf("channel_config must be valid JSON")
	}
	return nil
}
