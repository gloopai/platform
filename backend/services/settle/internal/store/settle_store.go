package store

import (
	"context"
	"database/sql"
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
