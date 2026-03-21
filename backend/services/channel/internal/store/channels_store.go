package store

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
)

type channelWeight struct {
	ID     int64
	Weight int64
}

type ChannelsStore struct {
	db *sql.DB
}

func NewChannelsStore(db *sql.DB) *ChannelsStore {
	return &ChannelsStore{db: db}
}

func (s *ChannelsStore) Route(ctx context.Context, payType string, amount int64) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, weight
FROM channels
WHERE enabled = 1
  AND fuse_enabled = 0
  AND (pay_type = ? OR pay_type = '' OR pay_type IS NULL)
  AND weight > 0
  AND (min_amount = 0 OR min_amount <= ?)
  AND (max_amount = 0 OR max_amount >= ?)
`, payType, amount, amount)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var items []channelWeight
	var total int64
	for rows.Next() {
		var it channelWeight
		if err := rows.Scan(&it.ID, &it.Weight); err != nil {
			return 0, err
		}
		items = append(items, it)
		total += it.Weight
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(items) == 0 || total <= 0 {
		return 0, errors.New("no available channel")
	}

	r := rand.Int64N(total)
	var acc int64
	for _, it := range items {
		acc += it.Weight
		if r < acc {
			return it.ID, nil
		}
	}
	return items[len(items)-1].ID, nil
}
