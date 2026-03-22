package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PayoutProductsStore 代付对外产品与通道绑定（管理台与后续代付 API 共用库访问）。
type PayoutProductsStore struct {
	db *sql.DB
}

func NewPayoutProductsStore(db *sql.DB) *PayoutProductsStore {
	return &PayoutProductsStore{db: db}
}

type PayoutProductAdmin struct {
	ID        int64
	Code      string
	Name      string
	SortOrder int64
	Enabled   bool
}

type PayoutProductBindingAdmin struct {
	ID              int64
	PayoutProductID int64
	ChannelID       int64
	ChannelName     string
	Weight          int64
	Enabled         bool
	CostRateBps     sql.NullInt64
}

func (s *PayoutProductsStore) AdminListAllPayoutProducts(ctx context.Context) ([]PayoutProductAdmin, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, code, name, sort_order, enabled FROM payout_products ORDER BY sort_order ASC, id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayoutProductAdmin
	for rows.Next() {
		var p PayoutProductAdmin
		var en int
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en); err != nil {
			return nil, err
		}
		p.Enabled = en == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PayoutProductsStore) AdminGetPayoutProduct(ctx context.Context, id int64) (*PayoutProductAdmin, error) {
	var p PayoutProductAdmin
	var en int
	err := s.db.QueryRowContext(ctx, `
SELECT id, code, name, sort_order, enabled FROM payout_products WHERE id = ? LIMIT 1
`, id).Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en)
	if err != nil {
		return nil, err
	}
	p.Enabled = en == 1
	return &p, nil
}

func (s *PayoutProductsStore) AdminCreatePayoutProduct(ctx context.Context, code, name string, sortOrder int64, enabled bool) (int64, error) {
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
INSERT INTO payout_products (code, name, sort_order, enabled) VALUES (?, ?, ?, ?)
`, code, name, sortOrder, en)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *PayoutProductsStore) AdminUpdatePayoutProduct(ctx context.Context, id int64, code, name string, sortOrder int64, enabled bool) error {
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE payout_products SET code = ?, name = ?, sort_order = ?, enabled = ?, updated_at = NOW() WHERE id = ?
`, code, name, sortOrder, en, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PayoutProductsStore) AdminListPayoutBindings(ctx context.Context, payoutProductID int64) ([]PayoutProductBindingAdmin, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ppc.id, ppc.payout_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled, ppc.cost_rate_bps
FROM payout_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.payout_product_id = ?
ORDER BY ppc.id ASC
`, payoutProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayoutProductBindingAdmin
	for rows.Next() {
		var b PayoutProductBindingAdmin
		var en int
		if err := rows.Scan(&b.ID, &b.PayoutProductID, &b.ChannelID, &b.ChannelName, &b.Weight, &en, &b.CostRateBps); err != nil {
			return nil, err
		}
		b.Enabled = en == 1
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *PayoutProductsStore) AdminUpsertPayoutBinding(ctx context.Context, payoutProductID, channelID int64, weight int64, enabled bool, costRateBps *int64) (int64, error) {
	if weight <= 0 {
		return 0, errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO payout_product_channels (payout_product_id, channel_id, weight, cost_rate_bps, enabled)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE weight = VALUES(weight), cost_rate_bps = VALUES(cost_rate_bps), enabled = VALUES(enabled), updated_at = NOW()
`, payoutProductID, channelID, weight, costRateBps, en)
	if err != nil {
		return 0, err
	}
	var bid int64
	err = s.db.QueryRowContext(ctx, `
SELECT id FROM payout_product_channels WHERE payout_product_id = ? AND channel_id = ? LIMIT 1
`, payoutProductID, channelID).Scan(&bid)
	if err != nil {
		return 0, fmt.Errorf("load binding id: %w", err)
	}
	return bid, nil
}

func (s *PayoutProductsStore) AdminUpdatePayoutBinding(ctx context.Context, bindingID int64, weight int64, enabled bool, costRateBps *int64) error {
	if weight <= 0 {
		return errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE payout_product_channels SET weight = ?, enabled = ?, cost_rate_bps = ?, updated_at = NOW() WHERE id = ?
`, weight, en, costRateBps, bindingID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PayoutProductsStore) AdminGetPayoutBindingByID(ctx context.Context, bindingID int64) (*PayoutProductBindingAdmin, error) {
	var b PayoutProductBindingAdmin
	var en int
	err := s.db.QueryRowContext(ctx, `
SELECT ppc.id, ppc.payout_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled, ppc.cost_rate_bps
FROM payout_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.id = ? LIMIT 1
`, bindingID).Scan(&b.ID, &b.PayoutProductID, &b.ChannelID, &b.ChannelName, &b.Weight, &en, &b.CostRateBps)
	if err != nil {
		return nil, err
	}
	b.Enabled = en == 1
	return &b, nil
}

func (s *PayoutProductsStore) AdminDeletePayoutBinding(ctx context.Context, bindingID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM payout_product_channels WHERE id = ?`, bindingID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PayoutProductsStore) AdminChannelSupportsPayout(ctx context.Context, channelID int64) (bool, error) {
	var sp int
	err := s.db.QueryRowContext(ctx, `SELECT supports_payout FROM channels WHERE id = ? LIMIT 1`, channelID).Scan(&sp)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return sp == 1, nil
}
