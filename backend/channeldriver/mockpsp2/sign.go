package mockpsp2

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
)

// SignMd5SortedKV 与 gateway internal/middleware.Md5Sign 一致：键名统一小写、排序、非空值参与，末尾 &key=secret，再 MD5 hex。
// 与 mock_psp 的 HMAC-SHA256 区分；不参与签名的字段不要放入 params。
func SignMd5SortedKV(params map[string]string, secret string) string {
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
