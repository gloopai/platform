package kvcache

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/zeromicro/go-zero/core/logx"
)

// PayoutProductBindingsSnapshot holds payout product → channel bindings from Consul KV.
type PayoutProductBindingsSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu   sync.RWMutex
	byID map[int64]*configkv.PayoutProductBindingsKV
}

func NewPayoutProductBindingsSnapshot(store *consulx.ConfigStore) *PayoutProductBindingsSnapshot {
	return &PayoutProductBindingsSnapshot{
		store:  store,
		prefix: configkv.PayoutProductChannelBindingsKVPrefix(),
		byID:   make(map[int64]*configkv.PayoutProductBindingsKV),
	}
}

func (c *PayoutProductBindingsSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *PayoutProductBindingsSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache payout product bindings SyncPrefixOnce: %v", err)
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

func (c *PayoutProductBindingsSnapshot) applyKV(key string, data []byte) {
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
	var kv configkv.PayoutProductBindingsKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache payout product bindings bad json key=%s: %v", key, err)
		return
	}
	c.byID[id] = &kv
}

// Get returns bindings for payout product id.
func (c *PayoutProductBindingsSnapshot) Get(payoutProductID int64) (*configkv.PayoutProductBindingsKV, bool) {
	if c == nil || payoutProductID <= 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byID[payoutProductID]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}
