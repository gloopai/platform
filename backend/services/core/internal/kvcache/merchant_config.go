package kvcache

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/zeromicro/go-zero/core/logx"
)

// MerchantConfig holds merchants.merchant_config JSON from Consul KV (OpenAPI GetAuthInfo / GetMerchant hot path).
type MerchantConfig struct {
	store  *consulx.ConfigStore
	prefix string

	mu      sync.RWMutex
	byMerID map[string]string
}

func NewMerchantConfig(store *consulx.ConfigStore) *MerchantConfig {
	return &MerchantConfig{
		store:   store,
		prefix:  consulx.MerchantConfigKVPrefix(),
		byMerID: make(map[string]string),
	}
}

func (c *MerchantConfig) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *MerchantConfig) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache merchant config SyncPrefixOnce: %v", err)
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

func (c *MerchantConfig) applyKV(key string, data []byte) {
	id, ok := parseMerchantIDKey(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byMerID, id)
		return
	}
	c.byMerID[id] = string(data)
}

func parseMerchantIDKey(fullKey, prefix string) (string, bool) {
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

// Get returns (json, true) if Consul has a non-empty blob for this merchant.
func (c *MerchantConfig) Get(merchantID string) (string, bool) {
	if c == nil {
		return "", false
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byMerID[merchantID]
	if !ok || strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}

// PickMerchantConfig returns Consul value if present, otherwise dbValue (merchants.merchant_config).
func PickMerchantConfig(cache *MerchantConfig, merchantID string, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(merchantID); ok {
			return v
		}
	}
	return dbValue
}
