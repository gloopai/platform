package mockpsp2

import (
	"sync"

	"github.com/gloopai/pay/channeldriver"
)

type payinRec struct {
	sysOrderNo      string
	merchantOrderNo string
	amountMinor     int64
	status          channeldriver.PayinOrderStatus
	referenceNo     string
	failReason      string
}

type payoutRec struct {
	sysOrderNo      string
	merchantOrderNo string
	amountMinor     int64
	status          channeldriver.PayoutOrderStatus
	referenceNo     string
}

// Store holds in-memory PSP state for tests.
type Store struct {
	mu sync.RWMutex

	payin  map[string]*payinRec
	payout map[string]*payoutRec

	availMinor     int64
	unsettledMinor int64
	frozenMinor    int64
}

// NewStore returns an empty store with default balance (1_000_000 minor units available).
func NewStore() *Store {
	return &Store{
		payin:          make(map[string]*payinRec),
		payout:         make(map[string]*payoutRec),
		availMinor:     1_000_000,
		unsettledMinor: 0,
		frozenMinor:    0,
	}
}

// SetBalances overrides mock balance snapshot (minor units).
func (s *Store) SetBalances(available, unsettled, frozen int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.availMinor = available
	s.unsettledMinor = unsettled
	s.frozenMinor = frozen
}
