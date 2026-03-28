package kvcache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/zeromicro/go-zero/core/logx"
)

// ChannelSnapshot holds full channel row JSON from Consul KV (mirrors channels table).
type ChannelSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]*configkv.ChannelKV
}

func NewChannelSnapshot(store *consulx.ConfigStore) *ChannelSnapshot {
	return &ChannelSnapshot{
		store:  store,
		prefix: configkv.ChannelSnapshotKVPrefix(),
		byID:   make(map[int64]*configkv.ChannelKV),
	}
}

func (c *ChannelSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *ChannelSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache channel snapshot SyncPrefixOnce: %v", err)
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

func (c *ChannelSnapshot) applyKV(key string, data []byte) {
	id, ok := parseChannelSnapshotID(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byID, id)
		return
	}
	var kv configkv.ChannelKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache channel snapshot bad json key=%s: %v", key, err)
		return
	}
	if kv.ID <= 0 {
		return
	}
	c.byID[id] = &kv
}

func parseChannelSnapshotID(fullKey, prefix string) (int64, bool) {
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

// Get returns (snapshot, true) if Consul has a valid blob for this channel.
func (c *ChannelSnapshot) Get(channelID int64) (*configkv.ChannelKV, bool) {
	if c == nil || channelID <= 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[channelID]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

// ForEach invokes fn for every cached channel (read lock).
func (c *ChannelSnapshot) ForEach(fn func(id int64, ch *configkv.ChannelKV)) {
	if c == nil || fn == nil {
		return
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for id, ch := range c.byID {
		if ch != nil {
			fn(id, ch)
		}
	}
}

// PickChannelConfig returns channel_config from the snapshot when present, otherwise dbValue.
func PickChannelConfig(cache *ChannelSnapshot, channelID int64, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(channelID); ok {
			return v.ChannelConfig
		}
	}
	return dbValue
}
