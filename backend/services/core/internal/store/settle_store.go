package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type SettleStore struct {
	db *gorm.DB
}

func NewSettleStore(db *gorm.DB) *SettleStore {
	return &SettleStore{db: db}
}

func (s *SettleStore) Credit(ctx context.Context, merchantId, orderNo string, amount int64, reason string) (bool, int64, error) {
	var changed bool
	var payinAfter int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payinBefore int64
		if err := tx.Raw(`SELECT payin_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Row().Scan(&payinBefore); err != nil {
			return err
		}

		var exists int
		err := tx.Raw(`
SELECT 1
FROM fund_logs
WHERE order_no = ? AND change_type = 'ORDER_PAID'
LIMIT 1
`, orderNo).Row().Scan(&exists)
		if err == nil {
			changed = false
			payinAfter = payinBefore
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}

		payinAfter = payinBefore + amount
		if err := tx.Exec(`UPDATE merchants SET payin_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, payinAfter, merchantId).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'ORDER_PAID', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, amount, payinBefore, payinAfter, reason).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	return changed, payinAfter, nil
}

func (s *SettleStore) DebitPayout(ctx context.Context, merchantId, orderNo string, amount int64, reason string) (bool, int64, error) {
	var changed bool
	var availableAfter int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var availableBefore int64
		if err := tx.Raw(`SELECT available_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Row().Scan(&availableBefore); err != nil {
			return err
		}
		if availableBefore < amount {
			changed = false
			availableAfter = availableBefore
			return ErrInsufficientBalance
		}
		_ = orderNo
		_ = reason

		availableAfter = availableBefore - amount
		if err := tx.Exec(`UPDATE merchants SET available_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, availableAfter, merchantId).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			return false, availableAfter, ErrInsufficientBalance
		}
		return false, 0, err
	}
	return changed, availableAfter, nil
}

func (s *SettleStore) TransferPayinToPayout(ctx context.Context, merchantId string, amount int64, reason string) (bool, int64, int64, error) {
	var changed bool
	var payinAfter, availableAfter int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payinBefore, availableBefore int64
		if err := tx.Raw(`SELECT payin_balance, available_balance FROM merchants WHERE merchant_id = ? FOR UPDATE`, merchantId).Row().Scan(&payinBefore, &availableBefore); err != nil {
			return err
		}
		if payinBefore < amount {
			changed = false
			payinAfter = payinBefore
			availableAfter = availableBefore
			return ErrInsufficientBalance
		}
		payinAfter = payinBefore - amount
		availableAfter = availableBefore + amount
		if err := tx.Exec(`
UPDATE merchants
SET payin_balance = ?, available_balance = ?, updated_at = NOW()
WHERE merchant_id = ?
`, payinAfter, availableAfter, merchantId).Error; err != nil {
			return err
		}
		transferNo := "TRANSFER-" + merchantId + "-" + time.Now().Format("20060102150405")
		if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'PAYIN_TO_PAYOUT', ?, ?, ?, ?, NOW())
`, merchantId, transferNo, amount, payinBefore, payinAfter, reason).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			return false, payinAfter, availableAfter, ErrInsufficientBalance
		}
		return false, 0, 0, err
	}
	return changed, payinAfter, availableAfter, nil
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
		var out []FundLogRow
		if err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, order_no, change_type, amount, balance_before, balance_after, COALESCE(reason,''), created_at
FROM fund_logs
ORDER BY created_at DESC
LIMIT ?
`, limit).Scan(&out).Error; err != nil {
			return nil, err
		}
		return out, nil
	}

	var out []FundLogRow
	if err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, order_no, change_type, amount, balance_before, balance_after, COALESCE(reason,''), created_at
FROM fund_logs
WHERE merchant_id = ?
ORDER BY created_at DESC
LIMIT ?
`, merchantId, limit).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
