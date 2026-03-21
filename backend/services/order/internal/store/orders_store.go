package store

import (
	"context"
	"database/sql"
	"time"
)

const (
	OrderStatusPending int32 = 0
	OrderStatusPaid    int32 = 1
	OrderStatusFailed  int32 = 2
	OrderStatusClosed  int32 = 3
)

type OrderRecord struct {
	OrderNo         string
	MerchantId      string
	MerchantOrderNo string
	Amount          int64
	Currency        string
	Status          int32
	ChannelId       int64
	ReturnUrl       string
	NotifyUrl       string
	UpstreamTradeNo string
	PaidAmount      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrdersStore struct {
	db *sql.DB
}

func NewOrdersStore(db *sql.DB) *OrdersStore {
	return &OrdersStore{db: db}
}

func (s *OrdersStore) FindByMerchantOrderNo(ctx context.Context, merchantId, merchantOrderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, return_url, notify_url, upstream_trade_no, paid_amount, created_at, updated_at
FROM orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo)
	return scanOrder(row)
}

func (s *OrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, return_url, notify_url, upstream_trade_no, paid_amount, created_at, updated_at
FROM orders
WHERE order_no = ?
LIMIT 1
`, orderNo)
	return scanOrder(row)
}

func (s *OrdersStore) Insert(ctx context.Context, rec *OrderRecord) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, return_url, notify_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.ReturnUrl, rec.NotifyUrl)
	return err
}

func (s *OrdersStore) MarkPaid(ctx context.Context, orderNo string, paidAmount int64, upstreamTradeNo string, channelId int64) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
UPDATE orders
SET status = ?, paid_amount = ?, upstream_trade_no = ?, channel_id = ?, updated_at = NOW()
WHERE order_no = ? AND status = ?
`, OrderStatusPaid, paidAmount, upstreamTradeNo, channelId, orderNo, OrderStatusPending)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (*OrderRecord, error) {
	var rec OrderRecord
	err := row.Scan(
		&rec.OrderNo,
		&rec.MerchantId,
		&rec.MerchantOrderNo,
		&rec.Amount,
		&rec.Currency,
		&rec.Status,
		&rec.ChannelId,
		&rec.ReturnUrl,
		&rec.NotifyUrl,
		&rec.UpstreamTradeNo,
		&rec.PaidAmount,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
