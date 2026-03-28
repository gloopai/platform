package store

import (
	"context"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OrderStatsStore 管理台统计：读库聚合 payin_orders（与 trade 共用库表）。
type OrderStatsStore struct {
	db *gorm.DB
}

func NewOrderStatsStore(db *gorm.DB) *OrderStatsStore {
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
	ProductCode string
	ProductName string
	OrderCount  int64
	PaidAmount  int64
	PaidCount   int64
	FailedCount int64
}

type ChannelAggRow struct {
	ChannelID   int64
	ChannelName string
	OrderCount  int64
	PaidAmount  int64
	PaidCount   int64
	FailedCount int64
}

// TodayOverview 统计「今日」创建订单（按服务器本地自然日）。
func (s *OrderStatsStore) TodayOverview(ctx context.Context) (TodayTotals, []ProductAggRow, []ChannelAggRow, error) {
	now := time.Now().In(time.Local)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return s.DayOverview(ctx, day, "")
}

// DayOverview 按本地自然日 [day 00:00, day+1 00:00) 聚合订单（与 TodayOverview 同一套指标）。
func (s *OrderStatsStore) DayOverview(ctx context.Context, day time.Time, merchantId string) (TodayTotals, []ProductAggRow, []ChannelAggRow, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.AddDate(0, 0, 1)
	startArg := start.Format("2006-01-02 15:04:05")
	endArg := end.Format("2006-01-02 15:04:05")
	merchantId = strings.TrimSpace(merchantId)

	where := "created_at >= ? AND created_at < ?"
	whereArgs := []any{startArg, endArg}
	merchantCondAlias := ""
	if merchantId != "" {
		where += " AND merchant_id = ?"
		merchantCondAlias = "  AND o.merchant_id = ?\n"
		whereArgs = append(whereArgs, merchantId)
	}

	var t TodayTotals
	var tRow struct {
		OrderCount   int64 `gorm:"column:order_count"`
		PaidAmount   int64 `gorm:"column:paid_amount"`
		PaidCount    int64 `gorm:"column:paid_count"`
		FailedCount  int64 `gorm:"column:failed_count"`
		PendingCount int64 `gorm:"column:pending_count"`
		ClosedCount  int64 `gorm:"column:closed_count"`
	}
	err := s.db.WithContext(ctx).Raw(`
SELECT
  COUNT(*) AS order_count,
  COALESCE(SUM(CASE WHEN status = 1 THEN COALESCE(paid_amount, amount) ELSE 0 END), 0) AS paid_amount,
  COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS paid_count,
  COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS failed_count,
  COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) AS pending_count,
  COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) AS closed_count
FROM payin_orders
WHERE `+where+`
`, whereArgs...).Scan(&tRow).Error
	if err != nil {
		return TodayTotals{}, nil, nil, err
	}
	t = TodayTotals{
		OrderCount:   tRow.OrderCount,
		PaidAmount:   tRow.PaidAmount,
		PaidCount:    tRow.PaidCount,
		FailedCount:  tRow.FailedCount,
		PendingCount: tRow.PendingCount,
		ClosedCount:  tRow.ClosedCount,
	}

	// 产品展示名由 core Channel 服务维护；此处仅按订单上的 product_code 聚合（避免 trade 直连 catalog 表）。
	prodRows, err := s.db.WithContext(ctx).Raw(`
SELECT
  COALESCE(NULLIF(TRIM(o.payin_product_code), ''), '(未指定产品)'),
  MAX(COALESCE(NULLIF(TRIM(o.payin_product_code), ''), '(未指定产品)')),
  COUNT(*),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN COALESCE(o.paid_amount, o.amount) ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 2 THEN 1 ELSE 0 END), 0)
FROM payin_orders o
WHERE o.created_at >= ? AND o.created_at < ?
`+merchantCondAlias+`
GROUP BY COALESCE(NULLIF(TRIM(o.payin_product_code), ''), '(未指定产品)')
ORDER BY 4 DESC, 3 DESC
`, append([]any{}, whereArgs...)...).Rows()
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

	chRows, err := s.db.WithContext(ctx).Raw(`
SELECT
  o.channel_id,
  IF(o.channel_id = 0, '未路由', CONCAT('通道#', o.channel_id)),
  COUNT(*),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN COALESCE(o.paid_amount, o.amount) ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 1 THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN o.status = 2 THEN 1 ELSE 0 END), 0)
FROM payin_orders o
WHERE o.created_at >= ? AND o.created_at < ?
`+merchantCondAlias+`
GROUP BY o.channel_id
ORDER BY 4 DESC, 3 DESC
`, append([]any{}, whereArgs...)...).Rows()
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
