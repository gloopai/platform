package store

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type NotifyLogRow struct {
	Id           int64
	MerchantId   string
	OrderNo      string
	NotifyUrl    string
	Attempt      int64
	HttpStatus   int64
	ResponseBody string
	ErrorMsg     string
	CreatedAt    time.Time
}

type NotifyLogsStore struct {
	db *gorm.DB
}

func NewNotifyLogsStore(db *gorm.DB) *NotifyLogsStore {
	return &NotifyLogsStore{db: db}
}

func (s *NotifyLogsStore) ListByOrder(ctx context.Context, merchantId, orderNo string, limit int64) ([]NotifyLogRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, order_no, notify_url, attempt, http_status, COALESCE(response_body,''), COALESCE(error_msg,''), created_at
FROM merchant_notify_logs
WHERE merchant_id = ? AND order_no = ?
ORDER BY created_at DESC
LIMIT ?
`, merchantId, orderNo, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []NotifyLogRow
	for rows.Next() {
		var r NotifyLogRow
		if err := rows.Scan(&r.Id, &r.MerchantId, &r.OrderNo, &r.NotifyUrl, &r.Attempt, &r.HttpStatus, &r.ResponseBody, &r.ErrorMsg, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
