package hexmeta

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
)

// signMD5 builds the upstream MD5 hex-lowercase signature (ASCII key sort, empty values omitted, append &key=secret).
func signMD5(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
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
		b.WriteString(params[k])
	}
	b.WriteString("&key=")
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
