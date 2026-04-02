package gatewaymw

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ReplayGuard rejects duplicate (merchant, nonce, timestamp) tuples for OpenAPI replay protection.
type ReplayGuard interface {
	MarkSeen(ctx context.Context, merchantID, nonce string, ts int64) (ok bool, err error)
}

// RedisReplayGuard stores seen keys in Redis with SetNX + TTL.
type RedisReplayGuard struct {
	cli       *redis.Client
	keyPrefix string
	ttl       time.Duration
}

// NewRedisReplayGuard builds a Redis-backed replay guard.
func NewRedisReplayGuard(cli *redis.Client, keyPrefix string, ttl time.Duration) *RedisReplayGuard {
	return &RedisReplayGuard{
		cli:       cli,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

// MarkSeen returns true if this is the first time the tuple is seen within TTL.
func (g *RedisReplayGuard) MarkSeen(ctx context.Context, merchantID, nonce string, ts int64) (bool, error) {
	key := g.keyPrefix + ":" + merchantID + ":" + nonce + ":" + time.Unix(ts, 0).UTC().Format(time.RFC3339)
	ok, err := g.cli.SetNX(ctx, key, "1", g.ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}
