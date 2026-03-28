package consulx

import "fmt"

// ChannelConfigKVPrefix is the Consul KV prefix for per-channel config JSON (under global config).
// Full key: pay/config/global/channels/config/{channel_id}
func ChannelConfigKVPrefix() string {
	return GlobalConfigPrefix() + "channels/config/"
}

// ChannelConfigKVKey returns the Consul KV key for a channel's config blob (DB column channel_config).
func ChannelConfigKVKey(channelID int64) string {
	if channelID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", ChannelConfigKVPrefix(), channelID)
}
