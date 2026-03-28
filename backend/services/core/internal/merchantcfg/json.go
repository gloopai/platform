package merchantcfg

import (
	"encoding/json"
	"strings"
)

// AppSecretFromMergedJSON returns app_secret from merged merchant_config JSON when non-empty;
// otherwise returns columnSecret (same idea as channel_config overriding sign_secret).
func AppSecretFromMergedJSON(columnSecret, mergedJSON string) string {
	mergedJSON = strings.TrimSpace(mergedJSON)
	if mergedJSON == "" {
		return columnSecret
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mergedJSON), &m); err != nil {
		return columnSecret
	}
	raw, ok := m["app_secret"]
	if !ok {
		return columnSecret
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return columnSecret
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return columnSecret
	}
	return s
}
