package mockpsp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// SignHMAC builds a deterministic signature for mock callbacks: HMAC-SHA256(secret, canonical).
// canonical is "k1=v1&k2=v2&..." with keys sorted excluding "sign".
func SignHMAC(secret string, fields map[string]string) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fields[k])
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil))
}
