package store

import (
	"context"
	"database/sql"
	"strings"
)

type PayinOrdersStore struct {
	db *sql.DB
}

func NewPayinOrdersStore(db *sql.DB) *PayinOrdersStore {
	return &PayinOrdersStore{db: db}
}

func (s *PayinOrdersStore) FindByMerchantOrderNo(ctx context.Context, merchantId, merchantOrderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, COALESCE(payin_product_code,''), channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payin_orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo)
	return scanOrder(row)
}

func (s *PayinOrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, COALESCE(payin_product_code,''), channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payin_orders
WHERE order_no = ?
LIMIT 1
`, orderNo)
	return scanOrder(row)
}

func (s *PayinOrdersStore) Insert(ctx context.Context, rec *OrderRecord) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO payin_orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, payin_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayinProductId, nullIfEmpty(rec.PayinProductCode), rec.ChannelLocked, rec.PaidAmount, rec.FeeMode, rec.FeeRateBps, rec.FeeFixedAmount, rec.FeeAmount, rec.NetAmount, rec.ReturnUrl, rec.NotifyUrl)
	return err
}

func (s *PayinOrdersStore) MarkPaid(ctx context.Context, orderNo string, paidAmount int64, upstreamTradeNo string, channelId int64) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
UPDATE payin_orders
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

func (s *PayinOrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, COALESCE(payin_product_code,''), channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payin_orders
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

func (s *PayinOrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, COALESCE(payin_product_code,''), channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payin_orders
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

func (s *PayinOrdersStore) UpdatePendingPayRoute(ctx context.Context, orderNo string, channelID, payProductID int64, payinProductCode string) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE payin_orders
SET channel_id = ?, payin_product_id = ?, payin_product_code = ?, updated_at = NOW()
WHERE order_no = ? AND status = ?
`, channelID, payProductID, nullIfEmpty(payinProductCode), orderNo, OrderStatusPending)
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

func (s *PayinOrdersStore) TodaySummary(ctx context.Context, merchantId string) (int64, int64, int64, error) {
	var (
		totalAmount  int64
		totalCount   int64
		successCount int64
	)

	if err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(amount), 0), COUNT(*)
FROM payin_orders
WHERE merchant_id = ? AND created_at >= CURDATE()
`, merchantId).Scan(&totalAmount, &totalCount); err != nil {
		return 0, 0, 0, err
	}

	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM payin_orders
WHERE merchant_id = ? AND created_at >= CURDATE() AND status = ?
`, merchantId, OrderStatusPaid).Scan(&successCount); err != nil {
		return 0, 0, 0, err
	}

	return totalAmount, totalCount, successCount, nil
}
