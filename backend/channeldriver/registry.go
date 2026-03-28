package channeldriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
)

// ChannelResolver loads [BindInput] for a channel row (DB + merged channel_config).
type ChannelResolver interface {
	ResolveBindInput(ctx context.Context, channelID int64) (BindInput, error)
}

// DriverFactory builds a bound [ChannelDriver] from merged channel_config JSON.
type DriverFactory func(channelID int64, channelConfigJSON string) (ChannelDriver, error)

// Registry maps driver_key (e.g. channels.payin_type) to constructors and caches per-channel instances.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]DriverFactory
	cache     map[int64]cachedDriver
}

type cachedDriver struct {
	driverKey string
	cfgHash   string
	drv       ChannelDriver
}

// NewRegistry returns an empty registry; register factories via [Registry.Register], then
// [setup.RegisterDefaultMockPSPs] or equivalent.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]DriverFactory),
		cache:     make(map[int64]cachedDriver),
	}
}

// Register binds a driver_key to a factory. Silently ignores empty key or nil factory.
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

// OpenPayin returns a driver for payin operations (same instance as payout when the PSP is unified).
func (r *Registry) OpenPayin(in BindInput) (ChannelDriver, error) {
	return r.open(in)
}

// OpenPayout returns a driver for payout operations.
func (r *Registry) OpenPayout(in BindInput) (ChannelDriver, error) {
	return r.open(in)
}

func (r *Registry) open(in BindInput) (ChannelDriver, error) {
	if r == nil {
		return nil, ErrNoDriver
	}
	key := strings.TrimSpace(in.DriverKey)
	r.mu.RLock()
	f, ok := r.factories[key]
	r.mu.RUnlock()
	if !ok {
		return nil, ErrNoDriver
	}
	return f(in.ChannelID, in.ChannelConfigJSON)
}

// GetChannelDriver resolves the channel row and returns a cached driver.
func (r *Registry) GetChannelDriver(ctx context.Context, channelID int64, res ChannelResolver) (ChannelDriver, error) {
	if r == nil || res == nil {
		return nil, ErrNoDriver
	}
	in, err := res.ResolveBindInput(ctx, channelID)
	if err != nil {
		return nil, err
	}
	h := hashChannelConfig(in.ChannelConfigJSON)
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
		return nil, ErrNoDriver
	}
	drv, err := f(in.ChannelID, in.ChannelConfigJSON)
	if err != nil {
		return nil, err
	}
	r.cache[channelID] = cachedDriver{driverKey: dk, cfgHash: h, drv: drv}
	return drv, nil
}

// InvalidateChannelDriver drops the cached driver after admin updates channel_config.
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
