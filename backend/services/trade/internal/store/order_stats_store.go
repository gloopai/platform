package store

import (
	"context"
	"database/sql"
	"math"
)

// OrderStatsStore 管理台统计：读库聚合 orders（与 trade 共用库表）。
type OrderStatsStore struct {
	db *sql.DB
}

func NewOrderStatsStore(db *sql.DB) *OrderStatsStore {
	return &OrderStatsStore{db: db}
}

type TodayTotals struct {
	OrderCount   int64
	PaidAmount   int64
	PaidCount    int64
	FailedCount  int64
	PendingCount int64
	ClosedCount  int64
}

type ProductAggRow struct {
	ProductCode  string
	ProductName  string
	OrderCount   int64
	PaidAmount   int64
	PaidCount    int64
	FailedCount  int64
}

type ChannelAggRow struct {
	ChannelID   int64
	ChannelName string
	OrderCount  int64
	PaidAmount  int64
	PaidCount   int64
	FailedCount int64
}

// TodayOverview 统计「今日」创建订单（按服务器自然日 CURDATE）。
func (s *OrderStatsStore) TodayOverview(ctx context.Context) (TodayTotals, []ProductAggRow, []ChannelAggRow, error) {
	var t TodayTotals
	err := s.db.QueryRowContext(ctx, `
SELECT
  COUNT(*),
  COALESCE(SUM(CASE WHEN status = 1 THEN COALESCE(paid_amount, amount) ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0)
FROM orders
WHERE created_at >= CURDATE() AND created_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY)
`).Scan(&t.OrderCount, &t.PaidAmount, &t.PaidCount, &t.FailedCount, &t.PendingCount, &t.ClosedCount)
	if err != nil {
		return TodayTotals{}, nil, nil, err
	}

	prodRows, err := s.db.QueryContext(ctx, `
SELECT
  COALESCE(NULLIF(TRIM(o.pay_product_code), ''), '(未指定产品)'),
  MAX(COALESCE(pp.name, NULLIF(TRIM(o.pay_product_code), ''), '(未指定产品)')),
  COUNT(*),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN COALESCE(o.paid_amount, o.amount) ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 2 THEN 1 ELSE 0 END), 0)
FROM orders o
LEFT JOIN pay_products pp ON pp.code = o.pay_product_code
WHERE o.created_at >= CURDATE() AND o.created_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY)
GROUP BY COALESCE(NULLIF(TRIM(o.pay_product_code), ''), '(未指定产品)')
ORDER BY 4 DESC, 3 DESC
`)
	if err != nil {
		return TodayTotals{}, nil, nil, err
	}
	defer prodRows.Close()

	var products []ProductAggRow
	for prodRows.Next() {
		var r ProductAggRow
		if err := prodRows.Scan(&r.ProductCode, &r.ProductName, &r.OrderCount, &r.PaidAmount, &r.PaidCount, &r.FailedCount); err != nil {
			return TodayTotals{}, nil, nil, err
		}
		products = append(products, r)
	}
	if err := prodRows.Err(); err != nil {
		return TodayTotals{}, nil, nil, err
	}

	chRows, err := s.db.QueryContext(ctx, `
SELECT
  o.channel_id,
  COALESCE(MAX(c.name), IF(o.channel_id = 0, '未路由', CONCAT('通道#', o.channel_id))),
  COUNT(*),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN COALESCE(o.paid_amount, o.amount) ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 2 THEN 1 ELSE 0 END), 0)
FROM orders o
LEFT JOIN channels c ON c.id = o.channel_id
WHERE o.created_at >= CURDATE() AND o.created_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY)
GROUP BY o.channel_id
ORDER BY 4 DESC, 3 DESC
`)
	if err != nil {
		return TodayTotals{}, products, nil, err
	}
	defer chRows.Close()

	var channels []ChannelAggRow
	for chRows.Next() {
		var r ChannelAggRow
		if err := chRows.Scan(&r.ChannelID, &r.ChannelName, &r.OrderCount, &r.PaidAmount, &r.PaidCount, &r.FailedCount); err != nil {
			return TodayTotals{}, products, nil, err
		}
		channels = append(channels, r)
	}
	if err := chRows.Err(); err != nil {
		return TodayTotals{}, products, nil, err
	}

	return t, products, channels, nil
}

func pct(num, den int64) float64 {
	if den <= 0 {
		return 0
	}
	return math.Round(float64(num)*10000/float64(den)) / 100
}

// RateConversion 成功笔数 / 今日创建订单数。
func RateConversion(paid, orders int64) float64 {
	return pct(paid, orders)
}

// RateTerminalSuccess 成功 / (成功+失败)，反映支付结果维度（不含待支付、关单）。
func RateTerminalSuccess(paid, failed int64) float64 {
	return pct(paid, paid+failed)
}
