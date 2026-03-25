package store

import (
	"context"

	"gorm.io/gorm"
)

// RoutingSummary 路由相关表的可观测计数（供管理台「路由策略」页展示）。
type RoutingSummary struct {
	EnabledPayinProducts         int64
	EnabledPayoutProducts        int64
	EnabledChannels              int64
	ActiveBindings               int64
	ActivePayoutBindings         int64
	MerchantsWithPayinWhitelist  int64
	MerchantsWithPayoutWhitelist int64
	FusedChannels                int64
}

type RoutingSummaryStore struct {
	db *gorm.DB
}

func NewRoutingSummaryStore(db *gorm.DB) *RoutingSummaryStore {
	return &RoutingSummaryStore{db: db}
}

func (s *RoutingSummaryStore) Get(ctx context.Context) (RoutingSummary, error) {
	var out RoutingSummary
	err := s.db.WithContext(ctx).Raw(`
SELECT
  (SELECT COUNT(*) FROM payin_products WHERE enabled = 1),
  (SELECT COUNT(*) FROM payout_products WHERE enabled = 1),
  (SELECT COUNT(*) FROM channels WHERE enabled = 1),
  (SELECT COUNT(*) FROM payin_product_channels WHERE enabled = 1),
  (SELECT COUNT(*) FROM payout_product_channels WHERE enabled = 1),
  (SELECT COUNT(DISTINCT merchant_id) FROM merchant_payin_products WHERE enabled = 1),
  (SELECT COUNT(DISTINCT merchant_id) FROM merchant_payout_products WHERE enabled = 1),
  (SELECT COUNT(*) FROM channels WHERE enabled = 1 AND fuse_enabled = 1)
`).Row().Scan(
		&out.EnabledPayinProducts,
		&out.EnabledPayoutProducts,
		&out.EnabledChannels,
		&out.ActiveBindings,
		&out.ActivePayoutBindings,
		&out.MerchantsWithPayinWhitelist,
		&out.MerchantsWithPayoutWhitelist,
		&out.FusedChannels,
	)
	if err != nil {
		return RoutingSummary{}, err
	}
	return out, nil
}
