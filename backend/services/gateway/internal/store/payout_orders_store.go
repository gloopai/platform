package store

import (
	"context"
	"database/sql"
)

type PayoutOrderSnapshot struct {
	OrderNo      string
	MerchantId   string
	Status       int32
	Amount       int64
	PaidAmount   int64
	ChannelId    int64
	UpstreamNo   string
	PayoutBalance int64
}

type PayoutOrdersStore struct {
	db *sql.DB
}

func NewPayoutOrdersStore(db *sql.DB) *PayoutOrdersStore {
	return &PayoutOrdersStore{db: db}
}

func (s *PayoutOrdersStore) MarkSuccess(ctx context.Context, orderNo, upstreamTradeNo string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
UPDATE payout_orders
SET status = 1,
    paid_amount = CASE WHEN paid_amount > 0 THEN paid_amount ELSE amount END,
    upstream_trade_no = CASE WHEN upstream_trade_no IS NULL OR upstream_trade_no = '' THEN ? ELSE upstream_trade_no END,
    updated_at = NOW()
WHERE order_no = ? AND status = 0
`, upstreamTradeNo, orderNo)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
