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

// PayinProductBindingsSnapshot holds payin product → channel bindings from Consul KV.
type PayinProductBindingsSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]*configkv.PayinProductBindingsKV
}

func NewPayinProductBindingsSnapshot(store *consulx.ConfigStore) *PayinProductBindingsSnapshot {
	return &PayinProductBindingsSnapshot{
		store:  store,
		prefix: configkv.PayinProductChannelBindingsKVPrefix(),
		byID:   make(map[int64]*configkv.PayinProductBindingsKV),
	}
}

func (c *PayinProductBindingsSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *PayinProductBindingsSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache payin product bindings SyncPrefixOnce: %v", err)
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

func (c *PayinProductBindingsSnapshot) applyKV(key string, data []byte) {
	id, ok := parseNumericSuffixKey(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byID, id)
		return
	}
	var kv configkv.PayinProductBindingsKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache payin product bindings bad json key=%s: %v", key, err)
		return
	}
	c.byID[id] = &kv
}

func parseNumericSuffixKey(fullKey, prefix string) (int64, bool) {
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

// Get returns bindings for payin product id.
func (c *PayinProductBindingsSnapshot) Get(payinProductID int64) (*configkv.PayinProductBindingsKV, bool) {
	if c == nil || payinProductID <= 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[payinProductID]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}
