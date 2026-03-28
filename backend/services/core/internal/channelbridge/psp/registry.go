// Package psp: driver registry and in-process ChannelDriver cache. Built-in drivers are registered in NewRegistry.
package psp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"gorm.io/gorm"
)

// DriverFactory builds a [contracts.ChannelDriver] for one channels.id (config loaded inside the driver).
type DriverFactory func(channelID int64) (contracts.ChannelDriver, error)

// Registry maps channels.driver_key to constructors and caches instances per channel_id.
type Registry struct {
	mu          sync.RWMutex
	factories   map[string]DriverFactory
	cache       map[int64]cachedDriver
	channels    *store.ChannelsStore
	channelSnap *kvcache.ChannelSnapshot
}

type cachedDriver struct {
	driverKey string
	cfgHash   string
	drv       contracts.ChannelDriver
}

// NewRegistry builds a registry with built-in drivers (hexmeta) wired to the same DB + channel KV snapshot as GetChannelDriver.
func NewRegistry(ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) *Registry {
	r := &Registry{
		factories:   make(map[string]DriverFactory),
		cache:       make(map[int64]cachedDriver),
		channels:    ch,
		channelSnap: snap,
	}
	if ch != nil {
		r.Register(hexmeta.DriverKey, func(channelID int64) (contracts.ChannelDriver, error) {
			return hexmeta.NewDriver(channelID, ch, snap)
		})
	}
	return r
}

func (r *Registry) Register(key string, f DriverFactory) {
	if r == nil {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" || f == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.factories == nil {
		r.factories = make(map[string]DriverFactory)
	}
	r.factories[key] = f
}

// configCacheHash detects config changes for the in-process driver cache (hexmeta uses canonical merged JSON).
func (r *Registry) configCacheHash(driverKey string, channelID int64, mergedConfigJSON string) string {
	hash := func(s string) string {
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])
	}
	if strings.TrimSpace(driverKey) == hexmeta.DriverKey && r != nil && r.channels != nil {
		s, err := hexmeta.CanonicalBindJSONFromKV(r.channels, r.channelSnap, channelID)
		if err == nil {
			return hash(s)
		}
	}
	return hash(mergedConfigJSON)
}

func (r *Registry) GetChannelDriver(ctx context.Context, channelID int64) (contracts.ChannelDriver, error) {
	if r == nil {
		return nil, contracts.ErrNoDriver
	}
	if r.channels == nil {
		return nil, fmt.Errorf("psp: registry channels not configured")
	}
	if channelID <= 0 {
		return nil, fmt.Errorf("psp: invalid channel_id")
	}
	row, err := r.channels.AdminGetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("psp: channel %d not found", channelID)
		}
		return nil, err
	}
	dk := strings.TrimSpace(row.DriverKey)
	merged := strings.TrimSpace(kvcache.PickChannelConfig(r.channelSnap, channelID, row.ChannelConfig))
	h := r.configCacheHash(dk, channelID, merged)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cache == nil {
		r.cache = make(map[int64]cachedDriver)
	}
	if c, ok := r.cache[channelID]; ok && c.cfgHash == h && c.driverKey == dk {
		return c.drv, nil
	}
	f, ok := r.factories[dk]
	if !ok {
		return nil, contracts.ErrNoDriver
	}
	drv, err := f(channelID)
	if err != nil {
		return nil, err
	}
	r.cache[channelID] = cachedDriver{driverKey: dk, cfgHash: h, drv: drv}
	return drv, nil
}

func (r *Registry) InvalidateChannelDriver(channelID int64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cache == nil {
		return
	}
	delete(r.cache, channelID)
}
