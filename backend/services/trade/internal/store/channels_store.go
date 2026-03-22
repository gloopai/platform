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
	ChannelID    int64
	Weight       int64
	PayProductID int64
}

type ChannelsStore struct {
	db *sql.DB
}

func NewChannelsStore(db *sql.DB) *ChannelsStore {
	return &ChannelsStore{db: db}
}

// Route 按支付产品编码选一条上游通道：优先 pay_products + pay_product_channels；否则回退到 channels.pay_type 旧逻辑。
func (s *ChannelsStore) Route(ctx context.Context, payProductCode string, amount int64) (channelID, payProductID int64, err error) {
	code := strings.TrimSpace(payProductCode)
	if code == "" {
		return 0, 0, errors.New("pay_type (product code) required")
	}

	if ch, pid, e := s.routeByPayProduct(ctx, code, amount); e == nil && ch > 0 {
		return ch, pid, nil
	}

	ch, e := s.routeLegacy(ctx, code, amount)
	if e != nil {
		return 0, 0, e
	}
	return ch, 0, nil
}

func (s *ChannelsStore) routeByPayProduct(ctx context.Context, payProductCode string, amount int64) (channelID, payProductID int64, err error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT c.id, ppc.weight, pp.id
FROM pay_products pp
INNER JOIN pay_product_channels ppc ON pp.id = ppc.pay_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.code = ? AND pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0
  AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
`, payProductCode, amount, amount)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var picks []routePick
	var total int64
	for rows.Next() {
		var p routePick
		if err := rows.Scan(&p.ChannelID, &p.Weight, &p.PayProductID); err != nil {
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
			return p.ChannelID, p.PayProductID, nil
		}
	}
	last := picks[len(picks)-1]
	return last.ChannelID, last.PayProductID, nil
}

func (s *ChannelsStore) routeLegacy(ctx context.Context, payType string, amount int64) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, weight
FROM channels
WHERE enabled = 1
  AND fuse_enabled = 0
  AND (pay_type = ? OR pay_type = '' OR pay_type IS NULL)
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
