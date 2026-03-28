package middleware

import "github.com/gloopai/pay/common/signmd5"

// Md5Sign builds the OpenAPI channel-notify style sign string (sorted key=value& + key=secret, then MD5 hex).
func Md5Sign(params map[string]string, secret string) string {
	return signmd5.SignSortedKV(params, secret)
}
