// Package psp: driver registry. Add new PSPs under drivers/<name>/ and register in register.go.
package psp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
)

// ChannelResolver loads [contracts.BindInput] for a channel row.
type ChannelResolver interface {
	ResolveBindInput(ctx context.Context, channelID int64) (contracts.BindInput, error)
}

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

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]DriverFactory),
		cache:     make(map[int64]cachedDriver),
	}
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

func hashChannelConfig(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (r *Registry) bindConfigHash(in contracts.BindInput) string {
	if strings.TrimSpace(in.DriverKey) == hexmeta.DriverKey && r != nil && r.channels != nil {
		s, err := hexmeta.CanonicalBindJSONFromKV(r.channels, r.channelSnap, in.ChannelID)
		if err == nil {
			return hashChannelConfig(s)
		}
	}
	return hashChannelConfig(in.ChannelConfigJSON)
}

func (r *Registry) OpenPayin(in contracts.BindInput) (contracts.ChannelDriver, error) {
	return r.open(in)
}

func (r *Registry) OpenPayout(in contracts.BindInput) (contracts.ChannelDriver, error) {
	return r.open(in)
}

func (r *Registry) open(in contracts.BindInput) (contracts.ChannelDriver, error) {
	if r == nil {
		return nil, contracts.ErrNoDriver
	}
	key := strings.TrimSpace(in.DriverKey)
	r.mu.RLock()
	f, ok := r.factories[key]
	r.mu.RUnlock()
	if !ok {
		return nil, contracts.ErrNoDriver
	}
	return f(in.ChannelID)
}

func (r *Registry) GetChannelDriver(ctx context.Context, channelID int64, res ChannelResolver) (contracts.ChannelDriver, error) {
	if r == nil || res == nil {
		return nil, contracts.ErrNoDriver
	}
	in, err := res.ResolveBindInput(ctx, channelID)
	if err != nil {
		return nil, err
	}
	h := r.bindConfigHash(in)
	dk := strings.TrimSpace(in.DriverKey)
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
	drv, err := f(in.ChannelID)
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
