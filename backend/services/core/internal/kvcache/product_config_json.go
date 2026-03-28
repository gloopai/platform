package kvcache

import (
	"encoding/json"
	"strings"
)

// DisplayNameFromProductJSON returns display_name from a product_config JSON blob when non-empty.
func DisplayNameFromProductJSON(mergedJSON string) string {
	mergedJSON = strings.TrimSpace(mergedJSON)
	if mergedJSON == "" {
		return ""
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mergedJSON), &m); err != nil {
		return ""
	}
	raw, ok := m["display_name"]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}
