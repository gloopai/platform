package store

import (
	"context"
	"errors"
	"strings"

	"github.com/gloopai/pay/common/model"
	"gorm.io/gorm"
)

type PayinOrdersStore struct {
	db *gorm.DB
}

func NewPayinOrdersStore(db *gorm.DB) *PayinOrdersStore {
	return &PayinOrdersStore{db: db}
}

func (s *PayinOrdersStore) FindByMerchantOrderNo(ctx context.Context, merchantId, merchantOrderNo string) (*model.OrderRecord, error) {
	var rec model.OrderRecord
	tx := s.db.WithContext(ctx).
		Table("payin_orders").
		Where("merchant_id = ? AND merchant_order_no = ?", merchantId, merchantOrderNo).
		Limit(1).
		Take(&rec)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &rec, nil
}

func (s *PayinOrdersStore) FindByOrderNo(ctx context.Context, orderNo string) (*model.OrderRecord, error) {
	var rec model.OrderRecord
	tx := s.db.WithContext(ctx).
		Table("payin_orders").
		Where("order_no = ?", orderNo).
		Limit(1).
		Take(&rec)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &rec, nil
}

func (s *PayinOrdersStore) Insert(ctx context.Context, rec *model.OrderRecord) error {
	return s.db.WithContext(ctx).Exec(`
INSERT INTO payin_orders (order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, payin_product_id, payin_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, rec.OrderNo, rec.MerchantId, rec.MerchantOrderNo, rec.Amount, rec.Currency, rec.Status, rec.ChannelId, rec.PayinProductId, nullIfEmpty(rec.PayinProductCode), rec.ChannelLocked, rec.PaidAmount, rec.FeeMode, rec.FeeRateBps, rec.FeeFixedAmount, rec.FeeAmount, rec.NetAmount, rec.ReturnUrl, rec.NotifyUrl).Error
}

func (s *PayinOrdersStore) MarkPaid(ctx context.Context, orderNo string, paidAmount int64, channelTradeNo string) (bool, error) {
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payin_orders
SET status = ?, paid_amount = ?, channel_trade_no = ?, updated_at = NOW()
WHERE order_no = ? AND status = ?
`, OrderStatusPaid, paidAmount, channelTradeNo, orderNo, OrderStatusPending)
	return tx.RowsAffected > 0, tx.Error
}

func normalizeOrderPageLimit(limit int64) int64 {
	if limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func normalizeOrderOffset(off int64) int64 {
	if off < 0 {
		return 0
	}
	return off
}

// ListByMerchant returns payin orders for a merchant with pagination; total is the count matching filters.
func (s *PayinOrdersStore) ListByMerchant(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]model.OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)

	var total int64
	q := s.db.WithContext(ctx).
		Table("payin_orders").
		Where("merchant_id = ?", merchantId)
	if keyword != "" {
		q = q.Where("(order_no = ? OR merchant_order_no = ?)", keyword, keyword)
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.OrderRecord
	if err := q.Order("created_at DESC").Limit(int(limit)).Offset(int(offset)).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// AdminList is the admin cross-merchant payin list with pagination; total is the count matching filters.
func (s *PayinOrdersStore) AdminList(ctx context.Context, merchantId, keyword string, status int32, offset, limit int64) ([]model.OrderRecord, int64, error) {
	limit = normalizeOrderPageLimit(limit)
	offset = normalizeOrderOffset(offset)
	keyword = strings.TrimSpace(keyword)
	merchantId = strings.TrimSpace(merchantId)

	var total int64
	q := s.db.WithContext(ctx).Table("payin_orders")
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
	var out []model.OrderRecord
	args := []any{}
	where := "WHERE 1=1"
	if merchantId != "" {
		where += " AND o.merchant_id = ?"
		args = append(args, merchantId)
	}
	if keyword != "" {
		where += " AND (o.order_no = ? OR o.merchant_order_no = ? OR o.merchant_id = ?)"
		args = append(args, keyword, keyword, keyword)
	}
	if status >= 0 {
		where += " AND o.status = ?"
		args = append(args, status)
	}
	args = append(args, limit, offset)
	tx := s.db.WithContext(ctx).Raw(`
SELECT o.order_no, o.merchant_id, o.merchant_order_no, o.amount, o.currency, o.status, o.channel_id,
  o.payin_product_id, COALESCE(o.payin_product_code,'') AS payin_product_code, o.channel_locked,
  o.paid_amount, o.fee_mode, o.fee_rate_bps, o.fee_fixed_amount, o.fee_amount, o.net_amount,
  COALESCE(o.return_url,'') AS return_url, COALESCE(o.notify_url,'') AS notify_url,
  COALESCE(o.channel_trade_no,'') AS channel_trade_no,
  o.created_at, o.updated_at,
  COALESCE(c.name,'') AS channel_name
FROM payin_orders o
LEFT JOIN channels c ON c.id = o.channel_id
`+where+` ORDER BY o.created_at DESC LIMIT ? OFFSET ?`, args...).Scan(&out)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return out, total, nil
}

func (s *PayinOrdersStore) UpdatePendingPayRoute(ctx context.Context, orderNo string, channelID, payProductID int64, payinProductCode string) error {
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payin_orders
SET channel_id = ?, payin_product_id = ?, payin_product_code = ?, updated_at = NOW()
WHERE order_no = ? AND status = ?
`, channelID, payProductID, nullIfEmpty(payinProductCode), orderNo, OrderStatusPending)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *PayinOrdersStore) TodaySummary(ctx context.Context, merchantId string) (int64, int64, int64, error) {
	var a struct {
		TotalAmount int64 `gorm:"column:total_amount"`
		TotalCount  int64 `gorm:"column:total_count"`
	}
	if err := s.db.WithContext(ctx).Raw(`
SELECT COALESCE(SUM(amount), 0) AS total_amount, COUNT(*) AS total_count
FROM payin_orders
WHERE merchant_id = ? AND created_at >= CURDATE()
`, merchantId).Scan(&a).Error; err != nil {
		return 0, 0, 0, err
	}

	var b struct {
		SuccessCount int64 `gorm:"column:success_count"`
	}
	if err := s.db.WithContext(ctx).Raw(`
SELECT COUNT(*) AS success_count
FROM payin_orders
WHERE merchant_id = ? AND created_at >= CURDATE() AND status = ?
`, merchantId, OrderStatusPaid).Scan(&b).Error; err != nil {
		return 0, 0, 0, err
	}

	return a.TotalAmount, a.TotalCount, b.SuccessCount, nil
}
