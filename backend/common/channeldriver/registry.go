package channeldriver

import (
	"fmt"
	"sync"
)

// Registry holds globally registered upstream protocol implementations keyed by driver_key.
// Register all drivers at process startup; lookups are safe for concurrent use.
type Registry struct {
	mu sync.RWMutex

	payin   map[string]PayinUpstream
	payout  map[string]PayoutUpstream
	balance map[string]BalanceUpstream
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		payin:   make(map[string]PayinUpstream),
		payout:  make(map[string]PayoutUpstream),
		balance: make(map[string]BalanceUpstream),
	}
}

// RegisterPayin registers a payin driver; key must match cfg.DriverKey at runtime.
func (r *Registry) RegisterPayin(drv PayinUpstream) error {
	if drv == nil {
		return fmt.Errorf("channeldriver: RegisterPayin: nil driver")
	}
	k := drv.Key()
	if k == "" {
		return fmt.Errorf("channeldriver: RegisterPayin: empty key")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.payin[k]; dup {
		return fmt.Errorf("channeldriver: RegisterPayin: duplicate key %q", k)
	}
	r.payin[k] = drv
	return nil
}

// RegisterPayout registers a payout driver.
func (r *Registry) RegisterPayout(drv PayoutUpstream) error {
	if drv == nil {
		return fmt.Errorf("channeldriver: RegisterPayout: nil driver")
	}
	k := drv.Key()
	if k == "" {
		return fmt.Errorf("channeldriver: RegisterPayout: empty key")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.payout[k]; dup {
		return fmt.Errorf("channeldriver: RegisterPayout: duplicate key %q", k)
	}
	r.payout[k] = drv
	return nil
}

// RegisterBalance registers a balance query driver (optional per PSP).
func (r *Registry) RegisterBalance(drv BalanceUpstream) error {
	if drv == nil {
		return fmt.Errorf("channeldriver: RegisterBalance: nil driver")
	}
	k := drv.Key()
	if k == "" {
		return fmt.Errorf("channeldriver: RegisterBalance: empty key")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.balance[k]; dup {
		return fmt.Errorf("channeldriver: RegisterBalance: duplicate key %q", k)
	}
	r.balance[k] = drv
	return nil
}

// Payin returns the payin implementation for driver_key, or ErrNoDriver.
func (r *Registry) Payin(driverKey string) (PayinUpstream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.payin[driverKey]
	if !ok {
		return nil, ErrNoDriver
	}
	return d, nil
}

// Payout returns the payout implementation for driver_key, or ErrNoDriver.
func (r *Registry) Payout(driverKey string) (PayoutUpstream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.payout[driverKey]
	if !ok {
		return nil, ErrNoDriver
	}
	return d, nil
}

// Balance returns the balance implementation for driver_key, or ErrNoDriver.
func (r *Registry) Balance(driverKey string) (BalanceUpstream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.balance[driverKey]
	if !ok {
		return nil, ErrNoDriver
	}
	return d, nil
}
