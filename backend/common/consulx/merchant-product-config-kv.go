package consulx

import (
	"fmt"
	"net/url"
	"strings"
)

// MerchantSnapshotKVPrefix is the Consul KV prefix for full merchant row JSON snapshots.
// Full key: pay/config/global/merchants/snapshot/{url_path_escape(merchant_id)}
func MerchantSnapshotKVPrefix() string {
	return GlobalConfigPrefix() + "merchants/snapshot/"
}

// MerchantSnapshotKVKey returns the Consul KV key for a merchant snapshot blob.
func MerchantSnapshotKVKey(merchantID string) string {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return ""
	}
	seg := url.PathEscape(merchantID)
	if seg == "" {
		return ""
	}
	return MerchantSnapshotKVPrefix() + seg
}

// PayinProductSnapshotKVPrefix is the prefix for full payin_products row JSON.
func PayinProductSnapshotKVPrefix() string {
	return GlobalConfigPrefix() + "payin_products/snapshot/"
}

// PayinProductSnapshotKVKey returns the Consul KV key for a payin product snapshot.
func PayinProductSnapshotKVKey(productID int64) string {
	if productID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayinProductSnapshotKVPrefix(), productID)
}

// PayoutProductSnapshotKVPrefix is the prefix for full payout_products row JSON.
func PayoutProductSnapshotKVPrefix() string {
	return GlobalConfigPrefix() + "payout_products/snapshot/"
}

// PayoutProductSnapshotKVKey returns the Consul KV key for a payout product snapshot.
func PayoutProductSnapshotKVKey(productID int64) string {
	if productID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayoutProductSnapshotKVPrefix(), productID)
}
