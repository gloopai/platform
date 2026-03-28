package store

import (
	"context"

	"gorm.io/gorm"
)

type PayoutOrdersStore struct {
	db *gorm.DB
}

func NewPayoutOrdersStore(db *gorm.DB) *PayoutOrdersStore {
	return &PayoutOrdersStore{db: db}
}

func (s *PayoutOrdersStore) MarkSuccess(ctx context.Context, orderNo, channelTradeNo string) (bool, error) {
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payout_orders
SET status = 1,
    paid_amount = CASE WHEN paid_amount > 0 THEN paid_amount ELSE amount END,
    channel_trade_no = CASE WHEN channel_trade_no IS NULL OR channel_trade_no = '' THEN ? ELSE channel_trade_no END,
    updated_at = NOW()
WHERE order_no = ? AND status = 0
`, channelTradeNo, orderNo)
	return tx.RowsAffected > 0, tx.Error
}

func (s *PayoutOrdersStore) MarkFailed(ctx context.Context, orderNo string) (bool, error) {
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payout_orders
SET status = 2,
    updated_at = NOW()
WHERE order_no = ? AND status = 0
`, orderNo)
	return tx.RowsAffected > 0, tx.Error
}

