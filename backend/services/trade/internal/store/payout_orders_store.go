package store

import (
	"context"
	"database/sql"
	"strings"
)

type PayoutOrdersStore struct {
	db *sql.DB
}

func NewPayoutOrdersStore(db *sql.DB) *PayoutOrdersStore {
	return &PayoutOrdersStore{db: db}
}

func (s *PayoutOrdersStore) FindByMerchantOrderNo(ctx context.Context, merchantId, merchantOrderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS pay_product_id, COALESCE(payout_product_code,'') AS pay_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo)
	return scanOrder(row)
}

func (s *PayoutOrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS pay_product_id, COALESCE(payout_product_code,'') AS pay_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE order_no = ?
LIMIT 1
`, orderNo)
	return scanOrder(row)
}

func (s *PayoutOrdersStore) Insert(ctx context.Context, rec *OrderRecord) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO payout_orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id, payout_product_code, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, notify_url, upstream_trade_no, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayProductId, nullIfEmpty(rec.PayProductCode), rec.PaidAmount, rec.FeeMode, rec.FeeRateBps, rec.FeeFixedAmount, rec.FeeAmount, rec.NetAmount, rec.NotifyUrl, nullIfEmpty(rec.UpstreamTradeNo))
	return err
}

func (s *PayoutOrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)
	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS pay_product_id, COALESCE(payout_product_code,'') AS pay_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE merchant_id = ?`
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

func (s *PayoutOrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, limit int64) ([]OrderRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)
	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS pay_product_id, COALESCE(payout_product_code,'') AS pay_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
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
