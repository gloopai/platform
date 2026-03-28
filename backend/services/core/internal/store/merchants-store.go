package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/model"
	"gorm.io/gorm"
)

const globalKeyMerchantNumericIDStart = "merchant_numeric_id_start"

// DefaultMerchantNumericIDFloor 与 global_settings 默认 seed 一致：未配置或异常时的取号下限
const DefaultMerchantNumericIDFloor int64 = 5000000000

type MerchantsStore struct {
	db *gorm.DB
}

func NewMerchantsStore(db *gorm.DB) *MerchantsStore {
	return &MerchantsStore{db: db}
}

func (s *MerchantsStore) GetByMerchantId(ctx context.Context, merchantId string) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
app_secret,
password_hash,
status,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url,
COALESCE(merchant_config,'') AS merchant_config`).
		Where("merchant_id = ?", merchantId).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) GetByAppId(ctx context.Context, appId string) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
app_secret,
password_hash,
status,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url,
COALESCE(merchant_config,'') AS merchant_config`).
		Where("app_id = ?", appId).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) GetByEmail(ctx context.Context, email string) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
app_secret,
password_hash,
status,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url,
COALESCE(merchant_config,'') AS merchant_config`).
		Where("email = ?", email).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) List(ctx context.Context, limit int64) ([]model.Merchant, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	var out []model.Merchant
	if err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, app_id, email, app_secret, password_hash, status,
       COALESCE(ip_whitelist,'') AS ip_whitelist,
       COALESCE(payin_balance, 0) AS payin_balance,
       COALESCE(available_balance, 0) AS available_balance,
       COALESCE(frozen_balance, 0) AS frozen_balance,
       COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
       COALESCE(notify_url,'') AS notify_url,
       COALESCE(return_url,'') AS return_url,
       COALESCE(merchant_config,'') AS merchant_config
FROM merchants
ORDER BY id DESC
LIMIT ?
`, limit).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetMerchantNumericIDFloor 从 global_settings 读取新建商户数字 ID 下限（含）；缺省或非法为 1。
func (s *MerchantsStore) GetMerchantNumericIDFloor(ctx context.Context) (int64, error) {
	var v string
	if err := s.db.WithContext(ctx).Raw(`
SELECT setting_value FROM global_settings WHERE setting_key = ? LIMIT 1
`, globalKeyMerchantNumericIDStart).Scan(&v).Error; err != nil {
		return DefaultMerchantNumericIDFloor, err
	}
	if strings.TrimSpace(v) == "" {
		return DefaultMerchantNumericIDFloor, nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil || n < 1 {
		return DefaultMerchantNumericIDFloor, nil
	}
	if n > 9999999999 {
		return 9999999999, nil
	}
	return n, nil
}

// AllocNextMerchantNumericID 原子取下一个数字商户号；实际值为 max(上一值+1, floor)，floor 来自系统配置「起始号」。
func (s *MerchantsStore) AllocNextMerchantNumericID(ctx context.Context, floor int64) (int64, error) {
	if floor < 1 {
		floor = 1
	}
	if floor > 9999999999 {
		floor = 9999999999
	}
	var n int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(`
UPDATE merchant_numeric_seq
SET next_id = LAST_INSERT_ID(GREATEST(next_id + 1, ?))
WHERE slot = 1
`, floor)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("merchant_numeric_seq missing or not updated (need slot=1 row)")
		}
		return tx.Raw(`SELECT LAST_INSERT_ID()`).Scan(&n).Error
	})
	return n, err
}

func (s *MerchantsStore) Create(ctx context.Context, m *model.Merchant) error {
	return s.db.WithContext(ctx).Exec(`
INSERT INTO merchants (merchant_id, app_id, email, app_secret, password_hash, status, ip_whitelist, payin_balance, available_balance, frozen_balance, withdrawn_amount, notify_url, return_url, merchant_config, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, m.MerchantId, m.AppId, m.Email, m.AppSecret, m.PasswordHash, m.Status, m.IpWhitelist, m.PayinBalance, m.AvailableBalance, m.FrozenBalance, m.WithdrawnAmount, m.NotifyUrl, m.ReturnUrl, m.MerchantConfig).Error
}

func (s *MerchantsStore) UpdateByMerchantId(ctx context.Context, merchantId string, m *model.Merchant) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE merchants
SET app_secret = ?, password_hash = ?, status = ?, ip_whitelist = ?, notify_url = ?, return_url = ?, merchant_config = ?, updated_at = NOW()
WHERE merchant_id = ?
`, m.AppSecret, m.PasswordHash, m.Status, m.IpWhitelist, m.NotifyUrl, m.ReturnUrl, m.MerchantConfig, merchantId).Error
}

func (s *MerchantsStore) UpdatePasswordByMerchantId(ctx context.Context, merchantId string, passwordHash string) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE merchants
SET password_hash = ?, updated_at = NOW()
WHERE merchant_id = ?
`, passwordHash, merchantId).Error
}

// GetPasswordHash returns only password_hash for merchant_id (merchant console login when row is otherwise served from Consul).
func (s *MerchantsStore) GetPasswordHash(ctx context.Context, merchantID string) (string, error) {
	var r struct {
		PasswordHash string `gorm:"column:password_hash"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT COALESCE(password_hash,'') AS password_hash FROM merchants WHERE merchant_id = ? LIMIT 1
`, merchantID).Scan(&r)
	if tx.Error != nil {
		return "", tx.Error
	}
	if tx.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return r.PasswordHash, nil
}
