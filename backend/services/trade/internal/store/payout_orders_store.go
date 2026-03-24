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
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo)
	return scanOrder(row)
}

func (s *PayoutOrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
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
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayinProductId, nullIfEmpty(rec.PayinProductCode), rec.PaidAmount, rec.FeeMode, rec.FeeRateBps, rec.FeeFixedAmount, rec.FeeAmount, rec.NetAmount, rec.NotifyUrl, nullIfEmpty(rec.UpstreamTradeNo))
	return err
}

func (s *PayoutOrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)

	where := "WHERE merchant_id = ?"
	args := []any{merchantId}
	if keyword != "" {
		where += " AND (order_no = ? OR merchant_order_no = ?)"
		args = append(args, keyword, keyword)
	}
	if status >= 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM payout_orders " + where
	if err := s.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
` + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []OrderRecord
	for rows.Next() {
		rec, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *PayoutOrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)

	where := "WHERE 1=1"
	args := []any{}
	if merchantId != "" {
		where += " AND merchant_id = ?"
		args = append(args, merchantId)
	}
	if keyword != "" {
		where += " AND (order_no = ? OR merchant_order_no = ? OR merchant_id = ?)"
		args = append(args, keyword, keyword, keyword)
	}
	if status >= 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM payout_orders " + where
	if err := s.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(upstream_trade_no,''), created_at, updated_at
FROM payout_orders
` + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []OrderRecord
	for rows.Next() {
		rec, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
