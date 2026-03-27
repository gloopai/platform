package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrInvalidWithdrawalStatus = errors.New("invalid withdrawal status")
var ErrWithdrawalNotFound = errors.New("withdrawal not found")

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
		availableAfter = m.AvailableBefore - amount
		if err := tx.Exec(`UPDATE merchants SET available_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, availableAfter, merchantId).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'PAYOUT_DEBIT', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, -amount, m.AvailableBefore, availableAfter, reason).Error; err != nil {
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

// DepositAvailable 向商户可用余额增加法币分（amountCents），并写入 fund_logs（AVAILABLE_DEPOSIT）。
func (s *SettleStore) DepositAvailable(ctx context.Context, merchantId string, amountCents int64, reason string) (orderNo string, availableAfter int64, err error) {
	if amountCents <= 0 {
		return "", 0, fmt.Errorf("invalid amount")
	}
	orderNo = fmt.Sprintf("DEP-%s-%d", merchantId, time.Now().UnixNano())
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m struct {
			AvailableBefore int64 `gorm:"column:available_balance"`
		}
		if err := tx.Table("merchants").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("available_balance").
			Where("merchant_id = ?", merchantId).
			Limit(1).
			Take(&m).Error; err != nil {
			return err
		}
		availableAfter = m.AvailableBefore + amountCents
		if err := tx.Exec(`UPDATE merchants SET available_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, availableAfter, merchantId).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'AVAILABLE_DEPOSIT', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, amountCents, m.AvailableBefore, availableAfter, reason).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return orderNo, availableAfter, nil
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

type WithdrawalRow struct {
	Id             int64
	WithdrawNo     string
	MerchantId     string
	ApplyAmount    int64
	FeeAmount      int64
	NetAmount      int64
	FiatDebitAmount int64
	Status         int32
	ReceiveAccount string
	ReceiveName    string
	BankName       string
	ApplyNote      string
	ReviewNote     string
	PayoutNote     string
	ReviewedBy     string
	ReviewedAt     *time.Time
	PayoutedBy     string
	PayoutedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
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

func (s *SettleStore) CreateWithdrawal(ctx context.Context, in WithdrawalRow) (WithdrawalRow, error) {
	out := in
	out.Status = 0
	out.NetAmount = in.ApplyAmount - in.FeeAmount
	if out.NetAmount < 0 || in.FiatDebitAmount <= 0 {
		return WithdrawalRow{}, fmt.Errorf("invalid amount")
	}
	if err := s.db.WithContext(ctx).Exec(`
INSERT INTO merchant_withdrawals (
  withdraw_no, merchant_id, apply_amount, fee_amount, net_amount, status,
  fiat_debit_amount, receive_account, receive_name, bank_name, apply_note
) VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?)
`, in.WithdrawNo, in.MerchantId, in.ApplyAmount, in.FeeAmount, out.NetAmount, in.FiatDebitAmount,
		in.ReceiveAccount, in.ReceiveName, in.BankName, in.ApplyNote).Error; err != nil {
		return WithdrawalRow{}, err
	}
	return s.GetWithdrawalByNo(ctx, in.WithdrawNo)
}

func sanitizeWithdrawNoLikeKeyword(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// ListWithdrawals 按条件分页查询提现申请；total 为符合筛选的总条数。
func (s *SettleStore) ListWithdrawals(ctx context.Context, merchantId string, limit, offset int64, status *int32, withdrawNoContains string) ([]WithdrawalRow, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	scope := func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table("merchant_withdrawals")
		if mid := strings.TrimSpace(merchantId); mid != "" {
			tx = tx.Where("merchant_id = ?", mid)
		}
		if status != nil {
			tx = tx.Where("status = ?", *status)
		}
		if kw := sanitizeWithdrawNoLikeKeyword(withdrawNoContains); kw != "" {
			tx = tx.Where("withdraw_no LIKE ?", "%"+kw+"%")
		}
		return tx
	}
	var total int64
	if err := s.db.WithContext(ctx).Scopes(scope).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []WithdrawalRow
	if err := s.db.WithContext(ctx).Scopes(scope).Order("created_at DESC").Offset(int(offset)).Limit(int(limit)).Scan(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *SettleStore) GetWithdrawalByNo(ctx context.Context, withdrawNo string) (WithdrawalRow, error) {
	var out WithdrawalRow
	if err := s.db.WithContext(ctx).
		Table("merchant_withdrawals").
		Where("withdraw_no = ?", withdrawNo).
		Limit(1).
		Take(&out).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return WithdrawalRow{}, ErrWithdrawalNotFound
		}
		return WithdrawalRow{}, err
	}
	return out, nil
}

func (s *SettleStore) ReviewWithdrawal(ctx context.Context, withdrawNo string, approved bool, reviewNote, operator string) (WithdrawalRow, bool, int64, error) {
	var out WithdrawalRow
	var changed bool
	var availableAfter int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("merchant_withdrawals").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("withdraw_no = ?", withdrawNo).
			Limit(1).
			Take(&out).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		if out.Status != 0 {
			changed = false
			return nil
		}
		if !approved {
			if err := tx.Exec(`
UPDATE merchant_withdrawals
SET status = 1, review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE withdraw_no = ?
`, reviewNote, operator, withdrawNo).Error; err != nil {
				return err
			}
			changed = true
			return nil
		}

		ok, after, err := s.debitPayoutInTx(ctx, tx, out.MerchantId, out.WithdrawNo, out.FiatDebitAmount, "WITHDRAW_APPROVE")
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvalidWithdrawalStatus
		}
		availableAfter = after
		if err := tx.Exec(`
UPDATE merchant_withdrawals
SET status = 2, review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE withdraw_no = ?
`, reviewNote, operator, withdrawNo).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		return WithdrawalRow{}, false, 0, err
	}
	latest, err := s.GetWithdrawalByNo(ctx, withdrawNo)
	if err != nil {
		return WithdrawalRow{}, changed, availableAfter, err
	}
	return latest, changed, availableAfter, nil
}

func (s *SettleStore) MarkWithdrawalPayoutSuccess(ctx context.Context, withdrawNo, payoutNote, operator string) (WithdrawalRow, bool, error) {
	var out WithdrawalRow
	changed := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("merchant_withdrawals").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("withdraw_no = ?", withdrawNo).
			Limit(1).
			Take(&out).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		if out.Status == 4 {
			changed = false
			return nil
		}
		if out.Status != 2 && out.Status != 3 {
			return ErrInvalidWithdrawalStatus
		}
		if err := tx.Exec(`
UPDATE merchant_withdrawals
SET status = 4, payout_note = ?, payouted_by = ?, payouted_at = NOW(), updated_at = NOW()
WHERE withdraw_no = ?
`, payoutNote, operator, withdrawNo).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		return WithdrawalRow{}, false, err
	}
	latest, err := s.GetWithdrawalByNo(ctx, withdrawNo)
	if err != nil {
		return WithdrawalRow{}, changed, err
	}
	return latest, changed, nil
}

func (s *SettleStore) debitPayoutInTx(ctx context.Context, tx *gorm.DB, merchantId, orderNo string, amount int64, reason string) (bool, int64, error) {
	var m struct {
		AvailableBefore int64 `gorm:"column:available_balance"`
	}
	if err := tx.WithContext(ctx).Table("merchants").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("available_balance").
		Where("merchant_id = ?", merchantId).
		Limit(1).
		Take(&m).Error; err != nil {
		return false, 0, err
	}
	if m.AvailableBefore < amount {
		return false, m.AvailableBefore, ErrInsufficientBalance
	}
	availableAfter := m.AvailableBefore - amount
	if err := tx.Exec(`UPDATE merchants SET available_balance = ?, updated_at = NOW() WHERE merchant_id = ?`, availableAfter, merchantId).Error; err != nil {
		return false, 0, err
	}
	if err := tx.Exec(`
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES (?, ?, 'PAYOUT_DEBIT', ?, ?, ?, ?, NOW())
`, merchantId, orderNo, -amount, m.AvailableBefore, availableAfter, reason).Error; err != nil {
		return false, 0, err
	}
	return true, availableAfter, nil
}
