package kvcache

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// PayinGrantsModelFromKV converts Consul snapshot grants to model.PayinGrant slice.
func PayinGrantsModelFromKV(g *configkv.MerchantPayinGrantsKV) []model.PayinGrant {
	if g == nil {
		return nil
	}
	var out []model.PayinGrant
	for _, row := range g.Grants {
		if row.PayinProductID <= 0 {
			continue
		}
		rp := row.MerchantRateBps
		out = append(out, model.PayinGrant{PayinProductID: row.PayinProductID, RateBps: rp})
	}
	return out
}

// MerchantPayinGrantsSnapshot holds merchant → payin product grants from Consul KV.
type MerchantPayinGrantsSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu      sync.RWMutex
	byMerID map[string]*configkv.MerchantPayinGrantsKV
}

func NewMerchantPayinGrantsSnapshot(store *consulx.ConfigStore) *MerchantPayinGrantsSnapshot {
	return &MerchantPayinGrantsSnapshot{
		store:   store,
		prefix:  configkv.MerchantPayinGrantsKVPrefix(),
		byMerID: make(map[string]*configkv.MerchantPayinGrantsKV),
	}
}

func (c *MerchantPayinGrantsSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *MerchantPayinGrantsSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache merchant payin grants SyncPrefixOnce: %v", err)
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

func (c *MerchantPayinGrantsSnapshot) applyKV(key string, data []byte) {
	id, ok := parseMerchantGrantsKey(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byMerID, id)
		return
	}
	var kv configkv.MerchantPayinGrantsKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache merchant payin grants bad json key=%s: %v", key, err)
		return
	}
	c.byMerID[id] = &kv
}

func parseMerchantGrantsKey(fullKey, prefix string) (string, bool) {
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

// Get returns grants snapshot when present (may be empty grants slice if key existed with empty — we delete empty keys on sync).
func (c *MerchantPayinGrantsSnapshot) Get(merchantID string) (*configkv.MerchantPayinGrantsKV, bool) {
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
