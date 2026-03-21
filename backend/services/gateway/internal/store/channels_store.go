package store

import (
	"context"
	"database/sql"
)

type Channel struct {
	ID                 int64
	Name               string
	PayType             string
	GatewayUrl         string
	UpstreamMerchantNo string
	RsaPrivateKey      string
	SignSecret         string
	Weight             int64
	MinAmount          int64
	MaxAmount          int64
	Enabled            bool
	FuseEnabled        bool
}

type ChannelsStore struct {
	db *sql.DB
}

func NewChannelsStore(db *sql.DB) *ChannelsStore {
	return &ChannelsStore{db: db}
}

func (s *ChannelsStore) GetByID(ctx context.Context, id int64) (*Channel, error) {
	var c Channel
	err := s.db.QueryRowContext(ctx, `
SELECT id, COALESCE(name,''), COALESCE(pay_type,''), COALESCE(gateway_url,''),
       COALESCE(upstream_merchant_no,''), COALESCE(rsa_private_key,''), COALESCE(sign_secret,''),
       weight, min_amount, max_amount,
       enabled, fuse_enabled
FROM channels
WHERE id = ?
LIMIT 1
`, id).Scan(
		&c.ID,
		&c.Name,
		&c.PayType,
		&c.GatewayUrl,
		&c.UpstreamMerchantNo,
		&c.RsaPrivateKey,
		&c.SignSecret,
		&c.Weight,
		&c.MinAmount,
		&c.MaxAmount,
		&c.Enabled,
		&c.FuseEnabled,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *ChannelsStore) List(ctx context.Context) ([]Channel, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, COALESCE(name,''), COALESCE(pay_type,''), COALESCE(gateway_url,''),
       COALESCE(upstream_merchant_no,''), COALESCE(rsa_private_key,''), COALESCE(sign_secret,''),
       weight, min_amount, max_amount,
       enabled, fuse_enabled
FROM channels
ORDER BY id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.PayType,
			&c.GatewayUrl,
			&c.UpstreamMerchantNo,
			&c.RsaPrivateKey,
			&c.SignSecret,
			&c.Weight,
			&c.MinAmount,
			&c.MaxAmount,
			&c.Enabled,
			&c.FuseEnabled,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *ChannelsStore) Create(ctx context.Context, c *Channel) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO channels (name, pay_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount, enabled, fuse_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, c.Name, c.PayType, c.GatewayUrl, c.UpstreamMerchantNo, c.RsaPrivateKey, c.SignSecret, c.Weight, c.MinAmount, c.MaxAmount, c.Enabled, c.FuseEnabled)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *ChannelsStore) Update(ctx context.Context, id int64, c *Channel) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE channels
SET name = ?, pay_type = ?, gateway_url = ?, upstream_merchant_no = ?, rsa_private_key = ?, sign_secret = ?,
    weight = ?, min_amount = ?, max_amount = ?, enabled = ?, fuse_enabled = ?, updated_at = NOW()
WHERE id = ?
`, c.Name, c.PayType, c.GatewayUrl, c.UpstreamMerchantNo, c.RsaPrivateKey, c.SignSecret,
		c.Weight, c.MinAmount, c.MaxAmount, c.Enabled, c.FuseEnabled, id)
	return err
}
