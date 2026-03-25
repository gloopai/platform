package store

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		var m struct {
			PayinBefore int64 `gorm:"column:payin_balance"`
		}
		if err := tx.
			Table("merchants").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("payin_balance").
			Where("merchant_id = ?", merchantId).
			Limit(1).
			Take(&m).Error; err != nil {
			return err
		}

		var one struct {
			One int `gorm:"column:one"`
		}
		existsTx := tx.
			Table("fund_logs").
			Select("1 AS one").
			Where("order_no = ? AND change_type = 'ORDER_PAID'", orderNo).
			Limit(1).
			Take(&one)
		if existsTx.Error == nil {
			changed = false
			payinAfter = m.PayinBefore
			return nil
		}
		if existsTx.Error != nil && existsTx.Error != gorm.ErrRecordNotFound {
			return existsTx.Error
		}

		payinAfter = m.PayinBefore + amount
		if err := tx.Exec(`UPDATE merchants SET payin_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, payinAfter, merchantId).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'ORDER_PAID', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, amount, m.PayinBefore, payinAfter, reason).Error; err != nil {
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
		var m struct {
			AvailableBefore int64 `gorm:"column:available_balance"`
		}
		if err := tx.
			Table("merchants").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("available_balance").
			Where("merchant_id = ?", merchantId).
			Limit(1).
			Take(&m).Error; err != nil {
			return err
		}
		if m.AvailableBefore < amount {
			changed = false
			availableAfter = m.AvailableBefore
			return ErrInsufficientBalance
		}
		_ = orderNo
		_ = reason

		availableAfter = m.AvailableBefore - amount
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
		var m struct {
			PayinBefore     int64 `gorm:"column:payin_balance"`
			AvailableBefore int64 `gorm:"column:available_balance"`
		}
		if err := tx.
			Table("merchants").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("payin_balance, available_balance").
			Where("merchant_id = ?", merchantId).
			Limit(1).
			Take(&m).Error; err != nil {
			return err
		}
		if m.PayinBefore < amount {
			changed = false
			payinAfter = m.PayinBefore
			availableAfter = m.AvailableBefore
			return ErrInsufficientBalance
		}
		payinAfter = m.PayinBefore - amount
		availableAfter = m.AvailableBefore + amount
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
`, merchantId, transferNo, amount, m.PayinBefore, payinAfter, reason).Error; err != nil {
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
