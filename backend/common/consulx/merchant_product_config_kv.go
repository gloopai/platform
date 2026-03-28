package consulx

import (
	"fmt"
	"net/url"
	"strings"
)

// MerchantConfigKVPrefix is the Consul KV prefix for per-merchant config JSON (mirrors merchants.merchant_config).
// Full key: pay/config/global/merchants/config/{url_path_escape(merchant_id)}
func MerchantConfigKVPrefix() string {
	return GlobalConfigPrefix() + "merchants/config/"
}

// MerchantConfigKVKey returns the Consul KV key for a merchant's config blob.
func MerchantConfigKVKey(merchantID string) string {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return ""
	}
	seg := url.PathEscape(merchantID)
	if seg == "" {
		return ""
	}
	return MerchantConfigKVPrefix() + seg
}

// PayinProductConfigKVPrefix mirrors payin_products.product_config JSON.
func PayinProductConfigKVPrefix() string {
	return GlobalConfigPrefix() + "payin_products/config/"
}

// PayinProductConfigKVKey returns the Consul KV key for a payin product's config blob.
func PayinProductConfigKVKey(productID int64) string {
	if productID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayinProductConfigKVPrefix(), productID)
}

// PayoutProductConfigKVPrefix mirrors payout_products.product_config JSON.
func PayoutProductConfigKVPrefix() string {
	return GlobalConfigPrefix() + "payout_products/config/"
}

// PayoutProductConfigKVKey returns the Consul KV key for a payout product's config blob.
func PayoutProductConfigKVKey(productID int64) string {
	if productID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayoutProductConfigKVPrefix(), productID)
}
