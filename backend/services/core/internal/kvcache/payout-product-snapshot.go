package kvcache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// PayoutProductSnapshot holds full payout_products row JSON from Consul KV.
type PayoutProductSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]*model.PayoutProductKV
}

func NewPayoutProductSnapshot(store *consulx.ConfigStore) *PayoutProductSnapshot {
	return &PayoutProductSnapshot{
		store:  store,
		prefix: consulx.PayoutProductSnapshotKVPrefix(),
		byID:   make(map[int64]*model.PayoutProductKV),
	}
}

func (c *PayoutProductSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *PayoutProductSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache payout product snapshot SyncPrefixOnce: %v", err)
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

func (c *PayoutProductSnapshot) applyKV(key string, data []byte) {
	id, ok := parsePayoutProductSnapshotID(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byID, id)
		return
	}
	var kv model.PayoutProductKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache payout product snapshot bad json key=%s: %v", key, err)
		return
	}
	if kv.ID <= 0 {
		return
	}
	c.byID[id] = &kv
}

func parsePayoutProductSnapshotID(fullKey, prefix string) (int64, bool) {
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

// Get returns (snapshot, true) if Consul has a valid blob for this product id.
func (c *PayoutProductSnapshot) Get(productID int64) (*model.PayoutProductKV, bool) {
	if c == nil || productID <= 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[productID]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

// PickPayoutProductConfig returns product_config from the snapshot when present, otherwise dbValue.
func PickPayoutProductConfig(cache *PayoutProductSnapshot, productID int64, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(productID); ok {
			return v.ProductConfig
		}
	}
	return dbValue
}
