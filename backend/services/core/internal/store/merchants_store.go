package store

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type Merchant struct {
	ID                   int64
	MerchantId           string
	ApiSecret            string
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
	err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, api_secret, status, default_payin_rate_bps, default_payout_rate_bps, COALESCE(ip_whitelist,''),
       COALESCE(payin_balance, 0), COALESCE(available_balance, 0), COALESCE(frozen_balance, 0), COALESCE(withdrawn_amount, 0),
       COALESCE(notify_url,''), COALESCE(return_url,'')
FROM merchants
WHERE merchant_id = ?
LIMIT 1
`, merchantId).Row().Scan(
		&m.ID,
		&m.MerchantId,
		&m.ApiSecret,
		&m.Status,
		&m.DefaultPayinRateBps,
		&m.DefaultPayoutRateBps,
		&m.IpWhitelist,
		&m.PayinBalance,
		&m.AvailableBalance,
		&m.FrozenBalance,
		&m.WithdrawnAmount,
		&m.NotifyUrl,
		&m.ReturnUrl,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MerchantsStore) List(ctx context.Context, limit int64) ([]Merchant, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	var out []Merchant
	if err := s.db.WithContext(ctx).Raw(`
SELECT id, merchant_id, api_secret, status, default_payin_rate_bps, default_payout_rate_bps, COALESCE(ip_whitelist,''),
       COALESCE(payin_balance, 0), COALESCE(available_balance, 0), COALESCE(frozen_balance, 0), COALESCE(withdrawn_amount, 0),
       COALESCE(notify_url,''), COALESCE(return_url,'')
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
INSERT INTO merchants (merchant_id, api_secret, status, default_payin_rate_bps, default_payout_rate_bps, ip_whitelist, payin_balance, available_balance, frozen_balance, withdrawn_amount, notify_url, return_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, m.MerchantId, m.ApiSecret, m.Status, m.DefaultPayinRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.PayinBalance, m.AvailableBalance, m.FrozenBalance, m.WithdrawnAmount, m.NotifyUrl, m.ReturnUrl).Error
}

func (s *MerchantsStore) UpdateByMerchantId(ctx context.Context, merchantId string, m *Merchant) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE merchants
SET api_secret = ?, status = ?, default_payin_rate_bps = ?, default_payout_rate_bps = ?, ip_whitelist = ?, notify_url = ?, return_url = ?, updated_at = NOW()
WHERE merchant_id = ?
`, m.ApiSecret, m.Status, m.DefaultPayinRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.NotifyUrl, m.ReturnUrl, merchantId).Error
}
