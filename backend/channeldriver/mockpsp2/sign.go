package mockpsp2

import "github.com/gloopai/pay/common/signmd5"

// SignMd5SortedKV 与 gateway OpenAPI 验签一致：键名统一小写、排序、非空值参与，末尾 &key=secret，再 MD5 hex。
// 与 mock_psp 的 HMAC-SHA256 区分；不参与签名的字段不要放入 params。
func SignMd5SortedKV(params map[string]string, secret string) string {
	return signmd5.SignSortedKV(params, secret)
}
