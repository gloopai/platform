package consulx

import "fmt"

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
