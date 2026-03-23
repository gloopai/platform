package store

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"strings"
)

type channelWeight struct {
	ID     int64
	Weight int64
}

type routePick struct {
	ChannelID      int64
	Weight         int64
	PayinProductID int64
}

type ChannelsStore struct {
	db *sql.DB
}

func NewChannelsStore(db *sql.DB) *ChannelsStore {
	return &ChannelsStore{db: db}
}

// Route 按支付产品编码选一条上游通道：优先 payin_products + payin_product_channels；否则回退到 channels.payin_type 旧逻辑。仅 supports_payin=1 的通道参与代收路由。
func (s *ChannelsStore) Route(ctx context.Context, payinProductCode string, amount int64) (channelID, payProductID int64, err error) {
	code := strings.TrimSpace(payinProductCode)
	if code == "" {
		return 0, 0, errors.New("payin_type (product code) required")
	}

	if ch, pid, e := s.routeByPayinProduct(ctx, code, amount); e == nil && ch > 0 {
		return ch, pid, nil
	}

	ch, e := s.routeLegacy(ctx, code, amount)
	if e != nil {
		return 0, 0, e
	}
	return ch, 0, nil
}

func (s *ChannelsStore) routeByPayinProduct(ctx context.Context, payinProductCode string, amount int64) (channelID, payProductID int64, err error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT c.id, ppc.weight, pp.id
FROM payin_products pp
INNER JOIN payin_product_channels ppc ON pp.id = ppc.payin_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.code = ? AND pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0
  AND c.supports_payin = 1
  AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
`, payinProductCode, amount, amount)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var picks []routePick
	var total int64
	for rows.Next() {
		var p routePick
		if err := rows.Scan(&p.ChannelID, &p.Weight, &p.PayinProductID); err != nil {
			return 0, 0, err
		}
		picks = append(picks, p)
		total += p.Weight
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	if len(picks) == 0 || total <= 0 {
		return 0, 0, nil
	}

	r := rand.Int64N(total)
	var acc int64
	for _, p := range picks {
		acc += p.Weight
		if r < acc {
			return p.ChannelID, p.PayinProductID, nil
		}
	}
	last := picks[len(picks)-1]
	return last.ChannelID, last.PayinProductID, nil
}

func (s *ChannelsStore) routeLegacy(ctx context.Context, payType string, amount int64) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, weight
FROM channels
WHERE enabled = 1
  AND fuse_enabled = 0
  AND supports_payin = 1
  AND (payin_type = ? OR payin_type = '' OR payin_type IS NULL)
  AND weight > 0
  AND (min_amount = 0 OR min_amount <= ?)
  AND (max_amount = 0 OR max_amount >= ?)
`, payType, amount, amount)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var items []channelWeight
	var total int64
	for rows.Next() {
		var it channelWeight
		if err := rows.Scan(&it.ID, &it.Weight); err != nil {
			return 0, err
		}
		items = append(items, it)
		total += it.Weight
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(items) == 0 || total <= 0 {
		return 0, errors.New("no available channel")
	}

	r := rand.Int64N(total)
	var acc int64
	for _, it := range items {
		acc += it.Weight
		if r < acc {
			return it.ID, nil
		}
	}
	return items[len(items)-1].ID, nil
}

// GetGatewayURLAndPayinType 用于收银台组装跳转/二维码载体。
func (s *ChannelsStore) GetGatewayURLAndPayinType(ctx context.Context, channelID int64) (gatewayURL, payinType string, err error) {
	err = s.db.QueryRowContext(ctx, `
SELECT COALESCE(gateway_url,''), COALESCE(payin_type,'')
FROM channels WHERE id = ? LIMIT 1
`, channelID).Scan(&gatewayURL, &payinType)
	return
}

func (s *ChannelsStore) GetSignSecret(ctx context.Context, channelId int64) (string, error) {
	var secret string
	if err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(sign_secret,'')
FROM channels
WHERE id = ?
LIMIT 1
`, channelId).Scan(&secret); err != nil {
		return "", err
	}
	return secret, nil
}

// Channel 管理台 CRUD（与 gateway 原 channels 表结构一致）。
type Channel struct {
	ID                     int64
	Name                   string
	PayinType              string
	GatewayUrl             string
	UpstreamMerchantNo     string
	RsaPrivateKey          string
	SignSecret             string
	Weight                 int64
	MinAmount              int64
	MaxAmount              int64
	SupportsPayin          bool
	SupportsPayout         bool
	UpstreamPayinRateBps   int64
	UpstreamPayoutRateBps  int64
	UpstreamPayoutFeeMode  int64
	UpstreamPayoutFixedFee int64
	Enabled                bool
	FuseEnabled            bool
}

func (s *ChannelsStore) AdminGetByID(ctx context.Context, id int64) (*Channel, error) {
	var c Channel
	var sc, sp int
	err := s.db.QueryRowContext(ctx, `
SELECT id, COALESCE(name,''), COALESCE(payin_type,''), COALESCE(gateway_url,''),
       COALESCE(upstream_merchant_no,''), COALESCE(rsa_private_key,''), COALESCE(sign_secret,''),
       weight, min_amount, max_amount,
       supports_payin, supports_payout, upstream_payin_rate_bps, upstream_payout_rate_bps, upstream_payout_fee_mode, upstream_payout_fixed_fee,
       enabled, fuse_enabled
FROM channels
WHERE id = ?
LIMIT 1
`, id).Scan(
		&c.ID,
		&c.Name,
		&c.PayinType,
		&c.GatewayUrl,
		&c.UpstreamMerchantNo,
		&c.RsaPrivateKey,
		&c.SignSecret,
		&c.Weight,
		&c.MinAmount,
		&c.MaxAmount,
		&sc,
		&sp,
		&c.UpstreamPayinRateBps,
		&c.UpstreamPayoutRateBps,
		&c.UpstreamPayoutFeeMode,
		&c.UpstreamPayoutFixedFee,
		&c.Enabled,
		&c.FuseEnabled,
	)
	if err != nil {
		return nil, err
	}
	c.SupportsPayin = sc == 1
	c.SupportsPayout = sp == 1
	return &c, nil
}

func (s *ChannelsStore) AdminList(ctx context.Context) ([]Channel, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, COALESCE(name,''), COALESCE(payin_type,''), COALESCE(gateway_url,''),
       COALESCE(upstream_merchant_no,''), COALESCE(rsa_private_key,''), COALESCE(sign_secret,''),
       weight, min_amount, max_amount,
       supports_payin, supports_payout, upstream_payin_rate_bps, upstream_payout_rate_bps, upstream_payout_fee_mode, upstream_payout_fixed_fee,
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
		var sc, sp int
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.PayinType,
			&c.GatewayUrl,
			&c.UpstreamMerchantNo,
			&c.RsaPrivateKey,
			&c.SignSecret,
			&c.Weight,
			&c.MinAmount,
			&c.MaxAmount,
			&sc,
			&sp,
			&c.UpstreamPayinRateBps,
			&c.UpstreamPayoutRateBps,
			&c.UpstreamPayoutFeeMode,
			&c.UpstreamPayoutFixedFee,
			&c.Enabled,
			&c.FuseEnabled,
		); err != nil {
			return nil, err
		}
		c.SupportsPayin = sc == 1
		c.SupportsPayout = sp == 1
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *ChannelsStore) AdminCreate(ctx context.Context, c *Channel) (int64, error) {
	sc, sp := 0, 0
	if c.SupportsPayin {
		sc = 1
	}
	if c.SupportsPayout {
		sp = 1
	}
	res, err := s.db.ExecContext(ctx, `
INSERT INTO channels (name, payin_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount,
  supports_payin, supports_payout, upstream_payin_rate_bps, upstream_payout_rate_bps, upstream_payout_fee_mode, upstream_payout_fixed_fee,
  enabled, fuse_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
`, c.Name, c.PayinType, c.GatewayUrl, c.UpstreamMerchantNo, c.RsaPrivateKey, c.SignSecret, c.Weight, c.MinAmount, c.MaxAmount,
		sc, sp, c.UpstreamPayinRateBps, c.UpstreamPayoutRateBps, c.UpstreamPayoutFeeMode, c.UpstreamPayoutFixedFee, c.Enabled, c.FuseEnabled)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *ChannelsStore) AdminUpdate(ctx context.Context, id int64, c *Channel) error {
	sc, sp := 0, 0
	if c.SupportsPayin {
		sc = 1
	}
	if c.SupportsPayout {
		sp = 1
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE channels
SET name = ?, payin_type = ?, gateway_url = ?, upstream_merchant_no = ?, rsa_private_key = ?, sign_secret = ?,
    weight = ?, min_amount = ?, max_amount = ?,
    supports_payin = ?, supports_payout = ?, upstream_payin_rate_bps = ?, upstream_payout_rate_bps = ?, upstream_payout_fee_mode = ?, upstream_payout_fixed_fee = ?,
    enabled = ?, fuse_enabled = ?, updated_at = NOW()
WHERE id = ?
`, c.Name, c.PayinType, c.GatewayUrl, c.UpstreamMerchantNo, c.RsaPrivateKey, c.SignSecret,
		c.Weight, c.MinAmount, c.MaxAmount, sc, sp, c.UpstreamPayinRateBps, c.UpstreamPayoutRateBps, c.UpstreamPayoutFeeMode, c.UpstreamPayoutFixedFee,
		c.Enabled, c.FuseEnabled, id)
	return err
}
