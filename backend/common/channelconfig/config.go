// Package channelconfig helps build and validate the channels.channel_config JSON blob for this
// product (legacy column merge, admin APIs). PSP-specific keys beyond these conventions are
// parsed only inside each channeldriver implementation.
package channelconfig

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

// ChannelConfigJSONForBind builds the effective channel_config JSON for drivers: starts from the
// stored column (or legacy split columns when empty via ChannelConfigJSONForAPI), then merges
// supports_payin / supports_payout from the row when those keys are absent.
func ChannelConfigJSONForBind(channelConfig string, leg LegacyChannelFields, supportsPayin, supportsPayout bool) (string, error) {
	base := ChannelConfigJSONForAPI(channelConfig, leg)
	base = strings.TrimSpace(base)
	if base == "" {
		m := map[string]interface{}{
			"supports_payin":  supportsPayin,
			"supports_payout": supportsPayout,
		}
		b, err := json.Marshal(m)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(base), &m); err != nil {
		return "", err
	}
	if _, ok := m["supports_payin"]; !ok {
		m["supports_payin"] = supportsPayin
	}
	if _, ok := m["supports_payout"]; !ok {
		m["supports_payout"] = supportsPayout
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// StringFromJSONObject returns a string value for one key from a JSON object (top-level).
// Unknown shape or missing key returns "".
func StringFromJSONObject(rawJSON, key string) string {
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" || key == "" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &m); err != nil || m == nil {
		return ""
	}
	return stringField(m[key])
}

func stringField(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
