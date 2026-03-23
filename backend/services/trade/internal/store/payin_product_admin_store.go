package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PayinProductAdmin 管理台支付产品行。
type PayinProductAdmin struct {
	ID        int64
	Code      string
	Name      string
	SortOrder int64
	Enabled   bool
}

// PayinProductBindingAdmin 产品与上游通道绑定（费率在通道与商户侧配置，此处仅路由权重）。
type PayinProductBindingAdmin struct {
	ID             int64
	PayinProductID int64
	ChannelID      int64
	ChannelName    string
	Weight         int64
	Enabled        bool
}

// AdminListAllPayinProducts 全部支付产品（含停用）。
func (s *PayinProductsStore) AdminListAllPayinProducts(ctx context.Context) ([]PayinProductAdmin, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, code, name, sort_order, enabled
FROM payin_products
ORDER BY sort_order ASC, id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayinProductAdmin
	for rows.Next() {
		var p PayinProductAdmin
		var en int
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en); err != nil {
			return nil, err
		}
		p.Enabled = en == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// AdminGetPayinProduct 按 ID。
func (s *PayinProductsStore) AdminGetPayinProduct(ctx context.Context, id int64) (*PayinProductAdmin, error) {
	var p PayinProductAdmin
	var en int
	err := s.db.QueryRowContext(ctx, `
SELECT id, code, name, sort_order, enabled FROM payin_products WHERE id = ? LIMIT 1
`, id).Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en)
	if err != nil {
		return nil, err
	}
	p.Enabled = en == 1
	return &p, nil
}

// AdminCreatePayinProduct 新建；code 唯一。
func (s *PayinProductsStore) AdminCreatePayinProduct(ctx context.Context, code, name string, sortOrder int64, enabled bool) (int64, error) {
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
INSERT INTO payin_products (code, name, sort_order, enabled) VALUES (?, ?, ?, ?)
`, code, name, sortOrder, en)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// AdminUpdatePayinProduct 更新。
func (s *PayinProductsStore) AdminUpdatePayinProduct(ctx context.Context, id int64, code, name string, sortOrder int64, enabled bool) error {
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE payin_products SET code = ?, name = ?, sort_order = ?, enabled = ?, updated_at = NOW() WHERE id = ?
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

// AdminListBindings 某产品下的通道绑定。
func (s *PayinProductsStore) AdminListBindings(ctx context.Context, payProductID int64) ([]PayinProductBindingAdmin, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ppc.id, ppc.payin_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled
FROM payin_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.payin_product_id = ?
ORDER BY ppc.id ASC
`, payProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayinProductBindingAdmin
	for rows.Next() {
		var b PayinProductBindingAdmin
		var en int
		if err := rows.Scan(&b.ID, &b.PayinProductID, &b.ChannelID, &b.ChannelName, &b.Weight, &en); err != nil {
			return nil, err
		}
		b.Enabled = en == 1
		out = append(out, b)
	}
	return out, rows.Err()
}

// AdminUpsertBinding 插入或更新 (product_id, channel_id) 唯一键。
func (s *PayinProductsStore) AdminUpsertBinding(ctx context.Context, payProductID, channelID int64, weight int64, enabled bool) (int64, error) {
	if weight <= 0 {
		return 0, errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO payin_product_channels (payin_product_id, channel_id, weight, enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled), updated_at = NOW()
`, payProductID, channelID, weight, en)
	if err != nil {
		return 0, err
	}
	var bid int64
	err = s.db.QueryRowContext(ctx, `
SELECT id FROM payin_product_channels WHERE payin_product_id = ? AND channel_id = ? LIMIT 1
`, payProductID, channelID).Scan(&bid)
	if err != nil {
		return 0, fmt.Errorf("load binding id: %w", err)
	}
	return bid, nil
}

// AdminUpdateBinding 按绑定行 ID 更新权重与启用。
func (s *PayinProductsStore) AdminUpdateBinding(ctx context.Context, bindingID int64, weight int64, enabled bool) error {
	if weight <= 0 {
		return errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE payin_product_channels SET weight = ?, enabled = ?, updated_at = NOW() WHERE id = ?
`, weight, en, bindingID)
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

// AdminGetBindingByID 单条绑定（含通道名）。
func (s *PayinProductsStore) AdminGetBindingByID(ctx context.Context, bindingID int64) (*PayinProductBindingAdmin, error) {
	var b PayinProductBindingAdmin
	var en int
	err := s.db.QueryRowContext(ctx, `
SELECT ppc.id, ppc.payin_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled
FROM payin_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.id = ? LIMIT 1
`, bindingID).Scan(&b.ID, &b.PayinProductID, &b.ChannelID, &b.ChannelName, &b.Weight, &en)
	if err != nil {
		return nil, err
	}
	b.Enabled = en == 1
	return &b, nil
}

// AdminDeleteBinding 删除一条绑定。
func (s *PayinProductsStore) AdminDeleteBinding(ctx context.Context, bindingID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM payin_product_channels WHERE id = ?`, bindingID)
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

// AdminChannelSupportsPayin 通道是否存在且支持代收。
func (s *PayinProductsStore) AdminChannelSupportsPayin(ctx context.Context, channelID int64) (bool, error) {
	var sc int
	err := s.db.QueryRowContext(ctx, `SELECT supports_payin FROM channels WHERE id = ? LIMIT 1`, channelID).Scan(&sc)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return sc == 1, nil
}

// AdminChannelExists 通道是否存在。
func (s *PayinProductsStore) AdminChannelExists(ctx context.Context, channelID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM channels WHERE id = ? LIMIT 1`, channelID).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
