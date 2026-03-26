package store

import (
	"context"

	"gorm.io/gorm"
)

type Merchant struct {
	ID                   int64
	MerchantId           string
	AppId                string
	Email                string
	AppSecret            string
	PasswordHash         string
	Status               int64
	DefaultPayinRateBps  int64
	DefaultPayoutRateBps int64
	IpWhitelist          string
	PayinBalance         int64
	AvailableBalance     int64
	FrozenBalance        int64
	WithdrawnAmount      int64
	NotifyUrl            string
	ReturnUrl            string
}

type MerchantsStore struct {
	db *gorm.DB
}

func NewMerchantsStore(db *gorm.DB) *MerchantsStore {
	return &MerchantsStore{db: db}
}

func (s *MerchantsStore) GetByMerchantId(ctx context.Context, merchantId string) (*Merchant, error) {
	var m Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
api_secret AS app_secret,
password_hash,
status,
default_payin_rate_bps,
default_payout_rate_bps,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url`).
		Where("merchant_id = ?", merchantId).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) GetByAppId(ctx context.Context, appId string) (*Merchant, error) {
	var m Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
api_secret AS app_secret,
password_hash,
status,
default_payin_rate_bps,
default_payout_rate_bps,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url`).
		Where("app_id = ?", appId).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) GetByEmail(ctx context.Context, email string) (*Merchant, error) {
	var m Merchant
	tx := s.db.WithContext(ctx).
		Table("merchants").
		Select(`
id,
merchant_id,
app_id,
email,
api_secret AS app_secret,
password_hash,
status,
default_payin_rate_bps,
default_payout_rate_bps,
COALESCE(ip_whitelist,'') AS ip_whitelist,
COALESCE(payin_balance, 0) AS payin_balance,
COALESCE(available_balance, 0) AS available_balance,
COALESCE(frozen_balance, 0) AS frozen_balance,
COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
COALESCE(notify_url,'') AS notify_url,
COALESCE(return_url,'') AS return_url`).
		Where("email = ?", email).
		Limit(1).
		Take(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}

func (s *MerchantsStore) List(ctx context.Context, limit int64) ([]Merchant, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	var out []Merchant
	if err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, app_id, email, api_secret AS app_secret, password_hash, status, default_payin_rate_bps, default_payout_rate_bps,
       COALESCE(ip_whitelist,'') AS ip_whitelist,
       COALESCE(payin_balance, 0) AS payin_balance,
       COALESCE(available_balance, 0) AS available_balance,
       COALESCE(frozen_balance, 0) AS frozen_balance,
       COALESCE(withdrawn_amount, 0) AS withdrawn_amount,
       COALESCE(notify_url,'') AS notify_url,
       COALESCE(return_url,'') AS return_url
FROM merchants
ORDER BY id DESC
LIMIT ?
`, limit).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *MerchantsStore) Create(ctx context.Context, m *Merchant) error {
	return s.db.WithContext(ctx).Exec(`
INSERT INTO merchants (merchant_id, app_id, email, api_secret, password_hash, status, default_payin_rate_bps, default_payout_rate_bps, ip_whitelist, payin_balance, available_balance, frozen_balance, withdrawn_amount, notify_url, return_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, m.MerchantId, m.AppId, m.Email, m.AppSecret, m.PasswordHash, m.Status, m.DefaultPayinRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.PayinBalance, m.AvailableBalance, m.FrozenBalance, m.WithdrawnAmount, m.NotifyUrl, m.ReturnUrl).Error
}

func (s *MerchantsStore) UpdateByMerchantId(ctx context.Context, merchantId string, m *Merchant) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE merchants
SET api_secret = ?, password_hash = ?, status = ?, default_payin_rate_bps = ?, default_payout_rate_bps = ?, ip_whitelist = ?, notify_url = ?, return_url = ?, updated_at = NOW()
WHERE merchant_id = ?
`, m.AppSecret, m.PasswordHash, m.Status, m.DefaultPayinRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.NotifyUrl, m.ReturnUrl, merchantId).Error
}

func (s *MerchantsStore) UpdatePasswordByMerchantId(ctx context.Context, merchantId string, passwordHash string) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE merchants
SET password_hash = ?, updated_at = NOW()
WHERE merchant_id = ?
`, passwordHash, merchantId).Error
}
