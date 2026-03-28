package store

import (
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
	PayinProductId    int64
	PayinProductCode  string
	ChannelLocked   int32
	ReturnUrl       string
	NotifyUrl       string
	ChannelTradeNo string
	PaidAmount      int64
	FeeMode         int64
	FeeRateBps      int64
	FeeFixedAmount  int64
	FeeAmount       int64
	NetAmount       int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// 仅管理台列表等带 JOIN channels 的查询填充；其余路径为空
	ChannelName string `gorm:"column:channel_name"`
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
		&rec.PayinProductId,
		&rec.PayinProductCode,
		&rec.ChannelLocked,
		&rec.PaidAmount,
		&rec.FeeMode,
		&rec.FeeRateBps,
		&rec.FeeFixedAmount,
		&rec.FeeAmount,
		&rec.NetAmount,
		&rec.ReturnUrl,
		&rec.NotifyUrl,
		&rec.ChannelTradeNo,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
