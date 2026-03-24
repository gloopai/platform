package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReplayGuard interface {
	MarkSeen(ctx context.Context, merchantID, nonce string, ts int64) (bool, error)
}

type RedisReplayGuard struct {
	cli       *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedisReplayGuard(cli *redis.Client, keyPrefix string, ttl time.Duration) *RedisReplayGuard {
	return &RedisReplayGuard{
		cli:       cli,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (g *RedisReplayGuard) MarkSeen(ctx context.Context, merchantID, nonce string, ts int64) (bool, error) {
	key := g.keyPrefix + ":" + merchantID + ":" + nonce + ":" + time.Unix(ts, 0).UTC().Format(time.RFC3339)
	ok, err := g.cli.SetNX(ctx, key, "1", g.ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}
