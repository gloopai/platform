package store

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type PayoutOrdersStore struct {
	db *gorm.DB
}

func NewPayoutOrdersStore(db *gorm.DB) *PayoutOrdersStore {
	return &PayoutOrdersStore{db: db}
}

func (s *PayoutOrdersStore) FindByMerchantOrderNo(ctx context.Context, merchantId, merchantOrderNo string) (*OrderRecord, error) {
	var rec OrderRecord
	tx := s.db.WithContext(ctx).Raw(`
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(channel_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE merchant_id = ? AND merchant_order_no = ?
LIMIT 1
`, merchantId, merchantOrderNo).Scan(&rec)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) || tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &rec, nil
}

func (s *PayoutOrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*OrderRecord, error) {
	var rec OrderRecord
	tx := s.db.WithContext(ctx).Raw(`
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(channel_trade_no,''), created_at, updated_at
FROM payout_orders
WHERE order_no = ?
LIMIT 1
`, orderNo).Scan(&rec)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) || tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &rec, nil
}

func (s *PayoutOrdersStore) Insert(ctx context.Context, rec *OrderRecord) error {
	return s.db.WithContext(ctx).Exec(`
INSERT INTO payout_orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id, payout_product_code, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, notify_url, channel_trade_no, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayinProductId, nullIfEmpty(rec.PayinProductCode), rec.PaidAmount, rec.FeeMode, rec.FeeRateBps, rec.FeeFixedAmount, rec.FeeAmount, rec.NetAmount, rec.NotifyUrl, nullIfEmpty(rec.ChannelTradeNo)).Error
}

func (s *PayoutOrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)

	var total int64
	q := s.db.WithContext(ctx).Table("payout_orders").Where("merchant_id = ?", merchantId)
	if keyword != "" {
		q = q.Where("(order_no = ? OR merchant_order_no = ?)", keyword, keyword)
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var out []OrderRecord
	args := []any{merchantId}
	where := "WHERE merchant_id = ?"
	if keyword != "" {
		where += " AND (order_no = ? OR merchant_order_no = ?)"
		args = append(args, keyword, keyword)
	}
	if status >= 0 {
		where += " AND status = ?"
		args = append(args, status)
	}
	args = append(args, limit, offset)
	tx := s.db.WithContext(ctx).Raw(`
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payout_product_id AS payin_product_id, COALESCE(payout_product_code,'') AS payin_product_code, 0 AS channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, '' AS return_url, notify_url, COALESCE(channel_trade_no,''), created_at, updated_at
FROM payout_orders
`+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...).Scan(&out)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return out, total, nil
}

func (s *PayoutOrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)

	var total int64
	q := s.db.WithContext(ctx).Table("payout_orders")
	if merchantId != "" {
		q = q.Where("merchant_id = ?", merchantId)
	}
	if keyword != "" {
		q = q.Where("(order_no = ? OR merchant_order_no = ? OR merchant_id = ?)", keyword, keyword, keyword)
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var out []OrderRecord
	args := []any{}
	where := "WHERE 1=1"
	if merchantId != "" {
		where += " AND po.merchant_id = ?"
		args = append(args, merchantId)
	}
	if keyword != "" {
		where += " AND (po.order_no = ? OR po.merchant_order_no = ? OR po.merchant_id = ?)"
		args = append(args, keyword, keyword, keyword)
	}
	if status >= 0 {
		where += " AND po.status = ?"
		args = append(args, status)
	}
	args = append(args, limit, offset)
	tx := s.db.WithContext(ctx).Raw(`
SELECT po.order_no, po.merchant_id, po.merchant_order_no, po.amount, po.currency, po.status, po.channel_id, po.payout_product_id AS payin_product_id, COALESCE(po.payout_product_code,'') AS payin_product_code, 0 AS channel_locked, po.paid_amount, po.fee_mode, po.fee_rate_bps, po.fee_fixed_amount, po.fee_amount, po.net_amount, '' AS return_url, po.notify_url, COALESCE(po.channel_trade_no,'') AS channel_trade_no, po.created_at, po.updated_at,
  COALESCE(
    NULLIF(TRIM(COALESCE(c.name,'')),''),
    (SELECT GROUP_CONCAT(DISTINCT ch.name ORDER BY ch.name SEPARATOR ', ')
     FROM payout_product_channels ppc
     INNER JOIN channels ch ON ch.id = ppc.channel_id
     WHERE ppc.payout_product_id = po.payout_product_id AND ppc.enabled = 1)
  ) AS channel_name
FROM payout_orders po
LEFT JOIN channels c ON c.id = po.channel_id AND po.channel_id > 0
`+where+` ORDER BY po.created_at DESC LIMIT ? OFFSET ?`, args...).Scan(&out)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return out, total, nil
}
