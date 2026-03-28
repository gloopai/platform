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

// PayoutGrantsModelFromKV converts Consul snapshot grants to model.PayoutGrant slice.
func PayoutGrantsModelFromKV(g *configkv.MerchantPayoutGrantsKV) []model.PayoutGrant {
	if g == nil {
		return nil
	}
	var out []model.PayoutGrant
	for _, row := range g.Grants {
		if row.PayoutProductID <= 0 {
			continue
		}
		feeMode := row.FeeMode
		if feeMode < 1 || feeMode > 3 {
			feeMode = 1
		}
		out = append(out, model.PayoutGrant{
			PayoutProductID: row.PayoutProductID,
			FeeMode:         feeMode,
			RateBps:         row.MerchantRateBps,
			FixedFeeAmount:  row.FeeFixedAmount,
		})
	}
	return out
}

// MerchantPayoutGrantsSnapshot holds merchant → payout product grants from Consul KV.
type MerchantPayoutGrantsSnapshot struct {
	store  *consulx.ConfigStore
	prefix string

	mu      sync.RWMutex
	byMerID map[string]*configkv.MerchantPayoutGrantsKV
}

func NewMerchantPayoutGrantsSnapshot(store *consulx.ConfigStore) *MerchantPayoutGrantsSnapshot {
	return &MerchantPayoutGrantsSnapshot{
		store:   store,
		prefix:  configkv.MerchantPayoutGrantsKVPrefix(),
		byMerID: make(map[string]*configkv.MerchantPayoutGrantsKV),
	}
}

func (c *MerchantPayoutGrantsSnapshot) Start(ctx context.Context) {
	if c.store == nil {
		return
	}
	go c.run(ctx)
}

func (c *MerchantPayoutGrantsSnapshot) run(ctx context.Context) {
	p := c.prefix
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.store.SyncPrefixOnce(syncCtx, p); err != nil {
		logx.Errorf("kvcache merchant payout grants SyncPrefixOnce: %v", err)
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

func (c *MerchantPayoutGrantsSnapshot) applyKV(key string, data []byte) {
	id, ok := parseMerchantPayoutGrantsKey(key, c.prefix)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(data) == 0 {
		delete(c.byMerID, id)
		return
	}
	var kv configkv.MerchantPayoutGrantsKV
	if err := json.Unmarshal(data, &kv); err != nil {
		logx.Errorf("kvcache merchant payout grants bad json key=%s: %v", key, err)
		return
	}
	c.byMerID[id] = &kv
}

func parseMerchantPayoutGrantsKey(fullKey, prefix string) (string, bool) {
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

// Get returns grants snapshot when present.
func (c *MerchantPayoutGrantsSnapshot) Get(merchantID string) (*configkv.MerchantPayoutGrantsKV, bool) {
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
