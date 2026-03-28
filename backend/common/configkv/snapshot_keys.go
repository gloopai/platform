package configkv

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

// ChannelSnapshotKVPrefix is the Consul KV prefix for full channel row JSON (under global config).
// Full key: pay/config/global/channels/snapshot/{channel_id}
func ChannelSnapshotKVPrefix() string {
	return GlobalConfigPrefix() + "channels/snapshot/"
}

// ChannelSnapshotKVKey returns the Consul KV key for a channel snapshot blob.
func ChannelSnapshotKVKey(channelID int64) string {
	if channelID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", ChannelSnapshotKVPrefix(), channelID)
}

// MerchantPayinGrantsKVPrefix is pay/config/global/merchants/payin_grants/snapshot/
func MerchantPayinGrantsKVPrefix() string {
	return GlobalConfigPrefix() + "merchants/payin_grants/snapshot/"
}

// MerchantPayinGrantsKVKey returns Consul KV key for merchant payin product grants snapshot.
func MerchantPayinGrantsKVKey(merchantID string) string {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return ""
	}
	seg := url.PathEscape(merchantID)
	if seg == "" {
		return ""
	}
	return MerchantPayinGrantsKVPrefix() + seg
}

// MerchantPayoutGrantsKVPrefix is pay/config/global/merchants/payout_grants/snapshot/
func MerchantPayoutGrantsKVPrefix() string {
	return GlobalConfigPrefix() + "merchants/payout_grants/snapshot/"
}

// MerchantPayoutGrantsKVKey returns Consul KV key for merchant payout product grants snapshot.
func MerchantPayoutGrantsKVKey(merchantID string) string {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return ""
	}
	seg := url.PathEscape(merchantID)
	if seg == "" {
		return ""
	}
	return MerchantPayoutGrantsKVPrefix() + seg
}

// PayinProductChannelBindingsKVPrefix is pay/config/global/payin_product_channels/snapshot/
func PayinProductChannelBindingsKVPrefix() string {
	return GlobalConfigPrefix() + "payin_product_channels/snapshot/"
}

// PayinProductChannelBindingsKVKey returns Consul KV for payin product → channel bindings.
func PayinProductChannelBindingsKVKey(payinProductID int64) string {
	if payinProductID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayinProductChannelBindingsKVPrefix(), payinProductID)
}

// PayoutProductChannelBindingsKVPrefix is pay/config/global/payout_product_channels/snapshot/
func PayoutProductChannelBindingsKVPrefix() string {
	return GlobalConfigPrefix() + "payout_product_channels/snapshot/"
}

// PayoutProductChannelBindingsKVKey returns Consul KV for payout product → channel bindings.
func PayoutProductChannelBindingsKVKey(payoutProductID int64) string {
	if payoutProductID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", PayoutProductChannelBindingsKVPrefix(), payoutProductID)
}
