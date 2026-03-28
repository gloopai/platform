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

// PayinProductSnapshot holds full payin_products row JSON from Consul KV.
type PayinProductSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu     sync.RWMutex
	byID   map[int64]*configkv.PayinProductKV
	byCode map[string]int64 // code(lower) -> id
}

func NewPayinProductSnapshot(store *consulx.ConfigStore) *PayinProductSnapshot {
	return &PayinProductSnapshot{
		store:  store,
		prefix: configkv.PayinProductSnapshotKVPrefix(),
		byID:   make(map[int64]*configkv.PayinProductKV),
		byCode: make(map[string]int64),
	}
}

func (c *PayinProductSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *PayinProductSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache payin product snapshot SyncPrefixOnce: %v", err)
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

func (c *PayinProductSnapshot) applyKV(key string, data []byte) {
	id, ok := parsePayinProductSnapshotID(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		if old := c.byID[id]; old != nil {
			if code := strings.TrimSpace(strings.ToLower(old.Code)); code != "" {
				if c.byCode[code] == id {
					delete(c.byCode, code)
				}
			}
		}
		delete(c.byID, id)
		return
	}
	var kv configkv.PayinProductKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache payin product snapshot bad json key=%s: %v", key, err)
		return
	}
	if kv.ID <= 0 {
		return
	}
	if old := c.byID[id]; old != nil {
		if code := strings.TrimSpace(strings.ToLower(old.Code)); code != "" {
			if c.byCode[code] == id {
				delete(c.byCode, code)
			}
		}
	}
	c.byID[id] = &kv
	if code := strings.TrimSpace(strings.ToLower(kv.Code)); code != "" {
		c.byCode[code] = kv.ID
	}
}

func parsePayinProductSnapshotID(fullKey, prefix string) (int64, bool) {
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
func (c *PayinProductSnapshot) Get(productID int64) (*configkv.PayinProductKV, bool) {
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

// GetByCode returns payin product snapshot by code (case-insensitive).
func (c *PayinProductSnapshot) GetByCode(code string) (*configkv.PayinProductKV, bool) {
	if c == nil {
		return nil, false
	}
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	id, ok := c.byCode[code]
	if !ok || id <= 0 {
		return nil, false
	}
	s, ok := c.byID[id]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

// ForEach invokes fn for every cached payin product (read lock).
func (c *PayinProductSnapshot) ForEach(fn func(id int64, p *configkv.PayinProductKV)) {
	if c == nil || fn == nil {
		return
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for id, p := range c.byID {
		if p != nil {
			fn(id, p)
		}
	}
}

// PickPayinProductConfig returns product_config from the snapshot when present, otherwise dbValue.
func PickPayinProductConfig(cache *PayinProductSnapshot, productID int64, dbValue string) string {
	if cache != nil {
		if v, ok := cache.Get(productID); ok {
			return v.ProductConfig
		}
	}
	return dbValue
}
