package hexmeta

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Config is parsed from channels.channel_config (merged with legacy columns via channelconfig).
type Config struct {
	GatewayURL string `json:"gateway_url"`
	AppID      string `json:"app_id"`
	Secret     string `json:"sign_secret"`
}

func parseConfig(raw string) (*Config, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("hexmeta: empty channel_config")
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("hexmeta: channel_config: %w", err)
	}
	cfg := &Config{
		GatewayURL: stringField(m["gateway_url"]),
		AppID:      stringField(m["app_id"]),
		Secret:     stringField(m["sign_secret"]),
	}
	if cfg.AppID == "" {
		cfg.AppID = stringField(m["channel_merchant_no"])
	}
	if cfg.GatewayURL == "" || cfg.AppID == "" || cfg.Secret == "" {
		return nil, fmt.Errorf("hexmeta: gateway_url, app_id (or channel_merchant_no), and sign_secret are required")
	}
	return cfg, nil
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
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.0f", t), "0"), ".")
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
