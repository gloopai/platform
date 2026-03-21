package store

import (
	"context"
	"database/sql"
	"time"
)

type FundLogRow struct {
	Id            int64
	MerchantId    string
	OrderNo       string
	ChangeType    string
	Amount        int64
	BalanceBefore int64
	BalanceAfter  int64
	Reason        string
	CreatedAt     time.Time
}

type FundLogsStore struct {
	db *sql.DB
}

func NewFundLogsStore(db *sql.DB) *FundLogsStore {
	return &FundLogsStore{db: db}
}

func (s *FundLogsStore) ListByMerchant(ctx context.Context, merchantId string, limit int64) ([]FundLogRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, merchant_id, order_no, change_type, amount, balance_before, balance_after, COALESCE(reason,''), created_at
FROM fund_logs
WHERE merchant_id = ?
ORDER BY created_at DESC
LIMIT ?
`, merchantId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FundLogRow
	for rows.Next() {
		var f FundLogRow
		if err := rows.Scan(&f.Id, &f.MerchantId, &f.OrderNo, &f.ChangeType, &f.Amount, &f.BalanceBefore, &f.BalanceAfter, &f.Reason, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

