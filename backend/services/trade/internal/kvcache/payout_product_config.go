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

// PayoutProductConfig holds payout_products.product_config JSON from Consul KV.
type PayoutProductConfig struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]string
}

func NewPayoutProductConfig(store *consulx.ConfigStore) *PayoutProductConfig {
	return &PayoutProductConfig{
		store:  store,
		prefix: consulx.PayoutProductConfigKVPrefix(),
		byID:   make(map[int64]string),
	}
}

func (c *PayoutProductConfig) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *PayoutProductConfig) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache payout product config SyncPrefixOnce: %v", err)
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

func (c *PayoutProductConfig) applyKV(key string, data []byte) {
	id, ok := parsePayoutProductID(key, c.prefix)
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

func parsePayoutProductID(fullKey, prefix string) (int64, bool) {
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

func (c *PayoutProductConfig) Get(productID int64) (string, bool) {
	if c == nil || productID <= 0 {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[productID]
	if !ok || strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}

// PickPayoutProductConfig returns Consul value if present, otherwise dbValue (payout_products.product_config).
func PickPayoutProductConfig(cache *PayoutProductConfig, productID int64, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(productID); ok {
			return v
		}
	}
	return dbValue
}
