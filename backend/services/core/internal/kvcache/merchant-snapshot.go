package kvcache

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// MerchantSnapshot holds per-merchant full-row JSON from Consul KV (mirrors merchants table; no password_hash in KV).
type MerchantSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu      sync.RWMutex
	byMerID map[string]*model.MerchantKV
}

func NewMerchantSnapshot(store *consulx.ConfigStore) *MerchantSnapshot {
	return &MerchantSnapshot{
		store:   store,
		prefix:  consulx.MerchantSnapshotKVPrefix(),
		byMerID: make(map[string]*model.MerchantKV),
	}
}

func (c *MerchantSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *MerchantSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache merchant snapshot SyncPrefixOnce: %v", err)
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

func (c *MerchantSnapshot) applyKV(key string, data []byte) {
	id, ok := parseMerchantSnapshotKey(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byMerID, id)
		return
	}
	var kv model.MerchantKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache merchant snapshot bad json key=%s: %v", key, err)
		return
	}
	if strings.TrimSpace(kv.MerchantID) == "" {
		return
	}
	c.byMerID[id] = &kv
}

func parseMerchantSnapshotKey(fullKey, prefix string) (string, bool) {
	suffix := strings.TrimPrefix(fullKey, prefix)
	suffix = strings.Trim(suffix, "/")
	if suffix == "" {
		return "", false
	}
	id, err := url.PathUnescape(suffix)
	if err != nil || strings.TrimSpace(id) == "" {
		return "", false
	}
	return strings.TrimSpace(id), true
}

// Get returns (snapshot, true) if Consul has a valid blob for this merchant.
func (c *MerchantSnapshot) Get(merchantID string) (*model.MerchantKV, bool) {
	if c == nil {
		return nil, false
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byMerID[merchantID]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

// PickMerchantConfig returns merchant_config from the snapshot when present, otherwise dbValue.
func PickMerchantConfig(cache *MerchantSnapshot, merchantID string, dbValue string) string {
	if cache != nil {
		if s, ok := cache.Get(merchantID); ok {
			return s.MerchantConfig
		}
	}
	return dbValue
}
