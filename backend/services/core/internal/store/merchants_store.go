package store

import (
	"context"
	"database/sql"
)

type Merchant struct {
	ID                     int64
	MerchantId             string
	ApiSecret              string
	Status                 int64
	DefaultCollectRateBps  int64
	DefaultPayoutRateBps   int64
	IpWhitelist            string
	Balance                int64
	FrozenBalance          int64
	WithdrawnAmount        int64
	NotifyUrl              string
	ReturnUrl              string
}

type MerchantsStore struct {
	db *sql.DB
}

func NewMerchantsStore(db *sql.DB) *MerchantsStore {
	return &MerchantsStore{db: db}
}

func (s *MerchantsStore) GetByMerchantId(ctx context.Context, merchantId string) (*Merchant, error) {
	var m Merchant
	err := s.db.QueryRowContext(ctx, `
SELECT id, merchant_id, api_secret, status, default_collect_rate_bps, default_payout_rate_bps, COALESCE(ip_whitelist,''),
       balance, COALESCE(frozen_balance, 0), COALESCE(withdrawn_amount, 0),
       COALESCE(notify_url,''), COALESCE(return_url,'')
FROM merchants
WHERE merchant_id = ?
LIMIT 1
`, merchantId).Scan(
		&m.ID,
		&m.MerchantId,
		&m.ApiSecret,
		&m.Status,
		&m.DefaultCollectRateBps,
		&m.DefaultPayoutRateBps,
		&m.IpWhitelist,
		&m.Balance,
		&m.FrozenBalance,
		&m.WithdrawnAmount,
		&m.NotifyUrl,
		&m.ReturnUrl,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MerchantsStore) List(ctx context.Context, limit int64) ([]Merchant, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, merchant_id, api_secret, status, default_collect_rate_bps, default_payout_rate_bps, COALESCE(ip_whitelist,''),
       balance, COALESCE(frozen_balance, 0), COALESCE(withdrawn_amount, 0),
       COALESCE(notify_url,''), COALESCE(return_url,'')
FROM merchants
ORDER BY id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Merchant
	for rows.Next() {
		var m Merchant
		if err := rows.Scan(
			&m.ID,
			&m.MerchantId,
			&m.ApiSecret,
			&m.Status,
			&m.DefaultCollectRateBps,
			&m.DefaultPayoutRateBps,
			&m.IpWhitelist,
			&m.Balance,
			&m.FrozenBalance,
			&m.WithdrawnAmount,
			&m.NotifyUrl,
			&m.ReturnUrl,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *MerchantsStore) Create(ctx context.Context, m *Merchant) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO merchants (merchant_id, api_secret, status, default_collect_rate_bps, default_payout_rate_bps, ip_whitelist, balance, frozen_balance, withdrawn_amount, notify_url, return_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, m.MerchantId, m.ApiSecret, m.Status, m.DefaultCollectRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.Balance, m.FrozenBalance, m.WithdrawnAmount, m.NotifyUrl, m.ReturnUrl)
	return err
}

func (s *MerchantsStore) UpdateByMerchantId(ctx context.Context, merchantId string, m *Merchant) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE merchants
SET api_secret = ?, status = ?, default_collect_rate_bps = ?, default_payout_rate_bps = ?, ip_whitelist = ?, notify_url = ?, return_url = ?, updated_at = NOW()
WHERE merchant_id = ?
`, m.ApiSecret, m.Status, m.DefaultCollectRateBps, m.DefaultPayoutRateBps, m.IpWhitelist, m.NotifyUrl, m.ReturnUrl, merchantId)
	return err
}
