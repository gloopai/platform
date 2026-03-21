package store

import (
	"context"
	"database/sql"
	"time"
)

type OrderRow struct {
	OrderNo         string
	MerchantId      string
	MerchantOrderNo string
	Amount          int64
	Currency        string
	Status          int32
	ChannelId       int64
	PaidAmount      int64
	UpstreamTradeNo string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrdersStore struct {
	db *sql.DB
}

func NewOrdersStore(db *sql.DB) *OrdersStore {
	return &OrdersStore{db: db}
}

func (s *OrdersStore) ListByMerchant(ctx context.Context, merchantId, orderNo string, status int32, limit int64) ([]OrderRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, paid_amount, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM orders
WHERE merchant_id = ?
`
	args := []any{merchantId}
	if orderNo != "" {
		query += " AND (order_no = ? OR merchant_order_no = ?)"
		args = append(args, orderNo, orderNo)
	}
	if status >= 0 {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OrderRow
	for rows.Next() {
		var o OrderRow
		if err := rows.Scan(
			&o.OrderNo,
			&o.MerchantId,
			&o.MerchantOrderNo,
			&o.Amount,
			&o.Currency,
			&o.Status,
			&o.ChannelId,
			&o.PaidAmount,
			&o.UpstreamTradeNo,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OrdersStore) TodaySummary(ctx context.Context, merchantId string) (int64, int64, int64, error) {
	var (
		totalAmount  int64
		totalCount   int64
		successCount int64
	)

	if err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(amount), 0), COUNT(*)
FROM orders
WHERE merchant_id = ? AND created_at >= CURDATE()
`, merchantId).Scan(&totalAmount, &totalCount); err != nil {
		return 0, 0, 0, err
	}

	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM orders
WHERE merchant_id = ? AND created_at >= CURDATE() AND status = 1
`, merchantId).Scan(&successCount); err != nil {
		return 0, 0, 0, err
	}

	return totalAmount, totalCount, successCount, nil
}

