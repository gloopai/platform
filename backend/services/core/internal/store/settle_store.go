package store

import (
	"context"
	"database/sql"
	"time"
)

type SettleStore struct {
	db *sql.DB
}

func NewSettleStore(db *sql.DB) *SettleStore {
	return &SettleStore{db: db}
}

func (s *SettleStore) Credit(ctx context.Context, merchantId, orderNo string, amount int64, reason string) (bool, int64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var balanceBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Scan(&balanceBefore); err != nil {
		return false, 0, err
	}

	var exists int
	err = tx.QueryRowContext(ctx, `
SELECT 1
FROM fund_logs
WHERE order_no = ? AND change_type = 'ORDER_PAID'
LIMIT 1
`, orderNo).Scan(&exists)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return false, 0, err
		}
		return false, balanceBefore, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, 0, err
	}

	balanceAfter := balanceBefore + amount
	if _, err := tx.ExecContext(ctx, `UPDATE merchants SET balance = ?, updated_at = NOW() WHERE merchant_id = ?`, balanceAfter, merchantId); err != nil {
		return false, 0, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'ORDER_PAID', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, amount, balanceBefore, balanceAfter, reason); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return true, balanceAfter, nil
}

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

func (s *SettleStore) ListByMerchant(ctx context.Context, merchantId string, limit int64) ([]FundLogRow, error) {
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
