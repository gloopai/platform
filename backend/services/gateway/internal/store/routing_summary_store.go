package store

import (
	"context"
	"database/sql"
)

// RoutingSummary 路由相关表的可观测计数（供管理台「路由策略」页展示）。
type RoutingSummary struct {
	EnabledPayProducts     int64
	EnabledChannels        int64
	ActiveBindings         int64
	MerchantsWithWhitelist int64
	FusedChannels          int64
}

type RoutingSummaryStore struct {
	db *sql.DB
}

func NewRoutingSummaryStore(db *sql.DB) *RoutingSummaryStore {
	return &RoutingSummaryStore{db: db}
}

func (s *RoutingSummaryStore) Get(ctx context.Context) (RoutingSummary, error) {
	var out RoutingSummary
	err := s.db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM pay_products WHERE enabled = 1),
  (SELECT COUNT(*) FROM channels WHERE enabled = 1),
  (SELECT COUNT(*) FROM pay_product_channels WHERE enabled = 1),
  (SELECT COUNT(DISTINCT merchant_id) FROM merchant_pay_products WHERE enabled = 1),
  (SELECT COUNT(*) FROM channels WHERE enabled = 1 AND fuse_enabled = 1)
`).Scan(
		&out.EnabledPayProducts,
		&out.EnabledChannels,
		&out.ActiveBindings,
		&out.MerchantsWithWhitelist,
		&out.FusedChannels,
	)
	if err != nil {
		return RoutingSummary{}, err
	}
	return out, nil
}
