package base

import "context"

// BalanceChannel is optionally implemented for PSP balance query (e.g. POST .../query/balance).
type BalanceChannel interface {
	Key() string
	QueryBalance(ctx context.Context, cfg *ChannelConfig) (*BalanceSnapshot, error)
}

// BalanceSnapshot maps PSP balance fields (amounts in minor units).
type BalanceSnapshot struct {
	AvailableMinor int64 // available for payout
	UnsettledMinor int64 // collected but not yet available for payout
	FrozenMinor    int64 // in-flight payout
}
