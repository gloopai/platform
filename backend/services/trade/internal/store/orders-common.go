package store

import (
	"strings"

	"github.com/gloopai/pay/common/model"
)

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (*model.OrderRecord, error) {
	var rec model.OrderRecord
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

// Order status re-exports for callers that used store.OrderStatus*.
var (
	OrderStatusPending = model.OrderStatusPending
	OrderStatusPaid    = model.OrderStatusPaid
	OrderStatusFailed  = model.OrderStatusFailed
	OrderStatusClosed  = model.OrderStatusClosed
)
