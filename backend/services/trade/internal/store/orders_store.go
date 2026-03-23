package store

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

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
	PayProductId    int64
	PayProductCode  string
	ChannelLocked   int32
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
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, pay_product_id, COALESCE(pay_product_code,''), channel_locked, paid_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo)
	return scanOrder(row)
}

func (s *OrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, pay_product_id, COALESCE(pay_product_code,''), channel_locked, paid_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM orders
WHERE order_no = ?
LIMIT 1
`, orderNo)
	return scanOrder(row)
}

func (s *OrdersStore) Insert(ctx context.Context, rec *OrderRecord) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, pay_product_id, pay_product_code, channel_locked, paid_amount, return_url, notify_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayProductId, nullIfEmpty(rec.PayProductCode), rec.ChannelLocked, rec.PaidAmount, rec.ReturnUrl, rec.NotifyUrl)
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

func (s *OrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, pay_product_id, COALESCE(pay_product_code,''), channel_locked, paid_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM orders
WHERE merchant_id = ?
`
	args := []any{merchantId}
	if keyword != "" {
		query += " AND (order_no = ? OR merchant_order_no = ?)"
		args = append(args, keyword, keyword)
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

	var out []OrderRecord
	for rows.Next() {
		rec, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// AdminList 管理台跨商户列表。merchantId 为空则不限商户；keyword 匹配 order_no、merchant_order_no 或 merchant_id（精确）。
func (s *OrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, pay_product_id, COALESCE(pay_product_code,''), channel_locked, paid_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM orders
WHERE 1=1`
	args := []any{}
	if merchantId != "" {
		query += " AND merchant_id = ?"
		args = append(args, merchantId)
	}
	if keyword != "" {
		query += " AND (order_no = ? OR merchant_order_no = ? OR merchant_id = ?)"
		args = append(args, keyword, keyword, keyword)
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

	var out []OrderRecord
	for rows.Next() {
		rec, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdatePendingPayRoute 待支付订单更新路由结果（收银台选定支付产品后调用）。
func (s *OrdersStore) UpdatePendingPayRoute(ctx context.Context, orderNo string, channelID, payProductID int64, payProductCode string) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE orders
SET channel_id = ?, pay_product_id = ?, pay_product_code = ?, updated_at = NOW()
WHERE order_no = ? AND status = ?
`, channelID, payProductID, nullIfEmpty(payProductCode), orderNo, OrderStatusPending)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
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
WHERE merchant_id = ? AND created_at >= CURDATE() AND status = ?
`, merchantId, OrderStatusPaid).Scan(&successCount); err != nil {
		return 0, 0, 0, err
	}

	return totalAmount, totalCount, successCount, nil
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
		&rec.PayProductId,
		&rec.PayProductCode,
		&rec.ChannelLocked,
		&rec.PaidAmount,
		&rec.ReturnUrl,
		&rec.NotifyUrl,
		&rec.UpstreamTradeNo,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
