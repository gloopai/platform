package kvcache

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/zeromicro/go-zero/core/logx"
)

// ChannelConfig holds the per-channel config JSON from Consul KV
// (pay/config/global/channels/config/{id}), mirroring DB channels.channel_config.
// For OpenAPI checkout / PrepareTerminalPay / gateway GetSignSecret only.
type ChannelConfig struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]string
}

func NewChannelConfig(store *consulx.ConfigStore) *ChannelConfig {
	return &ChannelConfig{
		store:  store,
		prefix: consulx.ChannelConfigKVPrefix(),
		byID:   make(map[int64]string),
	}
}

// Start loads an initial snapshot then subscribes to ConfigStore events for this prefix.
func (c *ChannelConfig) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *ChannelConfig) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache channel config SyncPrefixOnce: %v", err)
	}
	c.store.ForEachPrefix(p, func(key string, data []byte) {
		c.applyKV(key, data)
	})

	sub := c.store.Subscribe(256)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sub:
			if !ok {
				return
			}
			if strings.HasPrefix(ev.Key, p) {
				c.applyKV(ev.Key, ev.Data)
			}
		}
	}
}

func (c *ChannelConfig) applyKV(key string, data []byte) {
	id, ok := parseChannelID(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byID, id)
		return
	}
	c.byID[id] = string(data)
}

func parseChannelID(fullKey, prefix string) (int64, bool) {
	suffix := strings.TrimPrefix(fullKey, prefix)
	suffix = strings.Trim(suffix, "/")
	if suffix == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// Get returns (json, true) if Consul has a non-empty blob for this channel.
func (c *ChannelConfig) Get(channelID int64) (string, bool) {
	if c == nil || channelID <= 0 {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[channelID]
	if !ok || strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}

// PickChannelConfig returns Consul value if present, otherwise dbValue (channels.channel_config).
func PickChannelConfig(cache *ChannelConfig, channelID int64, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(channelID); ok {
			return v
		}
	}
	return dbValue
}
