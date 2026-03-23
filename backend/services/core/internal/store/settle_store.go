package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

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

	var collectBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT collect_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Scan(&collectBefore); err != nil {
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
		return false, collectBefore, nil
	}
	if err != sql.ErrNoRows {
		return false, 0, err
	}

	collectAfter := collectBefore + amount
	if _, err := tx.ExecContext(ctx, `UPDATE merchants SET collect_balance = ?, balance = ?, updated_at = NOW() WHERE merchant_id = ?`, collectAfter, collectAfter, merchantId); err != nil {
		return false, 0, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'ORDER_PAID', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, amount, collectBefore, collectAfter, reason); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return true, collectAfter, nil
}

func (s *SettleStore) DebitPayout(ctx context.Context, merchantId, orderNo string, amount int64, reason string) (bool, int64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var payoutBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT payout_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Scan(&payoutBefore); err != nil {
		return false, 0, err
	}
	if payoutBefore < amount {
		return false, payoutBefore, ErrInsufficientBalance
	}
	_ = orderNo
	_ = reason

	payoutAfter := payoutBefore - amount
	if _, err := tx.ExecContext(ctx, `UPDATE merchants SET payout_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, payoutAfter, merchantId); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return true, payoutAfter, nil
}

func (s *SettleStore) TransferCollectToPayout(ctx context.Context, merchantId string, amount int64, reason string) (bool, int64, int64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var collectBefore, payoutBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT collect_balance, payout_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Scan(&collectBefore, &payoutBefore); err != nil {
		return false, 0, 0, err
	}
	if collectBefore < amount {
		return false, collectBefore, payoutBefore, ErrInsufficientBalance
	}
	collectAfter := collectBefore - amount
	payoutAfter := payoutBefore + amount
	if _, err := tx.ExecContext(ctx, `
UPDATE merchants
SET collect_balance = ?, payout_balance = ?, balance = ?, updated_at = NOW()
WHERE merchant_id = ?
`, collectAfter, payoutAfter, collectAfter, merchantId); err != nil {
		return false, 0, 0, err
	}
	transferNo := "TRANSFER-" + merchantId + "-" + time.Now().Format("20060102150405")
	if _, err := tx.ExecContext(ctx, `
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'COLLECT_TO_PAYOUT', ?, ?, ?, ?, NOW())
`, merchantId, transferNo, amount, collectBefore, collectAfter, reason); err != nil {
		return false, 0, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, 0, err
	}
	return true, collectAfter, payoutAfter, nil
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
	if merchantId == "" {
		rows, err := s.db.QueryContext(ctx, `
SELECT id, merchant_id, order_no, change_type, amount, balance_before, balance_after, COALESCE(reason,''), created_at
FROM fund_logs
ORDER BY created_at DESC
LIMIT ?
`, limit)
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
