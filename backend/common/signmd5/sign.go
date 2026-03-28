// Package signmd5 implements the OpenAPI / merchant-notify style MD5 signature:
// normalize keys to lower case, skip sign/signature, sort, join k=v&…, append &key=secret, MD5 hex.
package signmd5

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
)

// SignSortedKV builds the sorted key=value string (non-empty values), appends &key=secret, returns MD5 hex.
// Keys are normalized to lower case; "sign" and "signature" are ignored. Values are trimmed.
func SignSortedKV(params map[string]string, secret string) string {
	norm := make(map[string]string, len(params))
	for k, v := range params {
		norm[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	keys := make([]string, 0, len(norm))
	for k := range norm {
		if k == "sign" || k == "signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		v := norm[k]
		if v == "" {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	if b.Len() > 0 {
		b.WriteByte('&')
	}
	b.WriteString("key=")
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
