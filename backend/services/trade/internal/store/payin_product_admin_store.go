package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gloopai/pay/common/model"
	"gorm.io/gorm"
)

// AdminListAllPayinProducts 全部支付产品（含停用）。
func (s *PayinProductsStore) AdminListAllPayinProducts(ctx context.Context) ([]model.PayinProductAdmin, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT id, code, name, sort_order, enabled, COALESCE(product_config,'') AS product_config
FROM payin_products
ORDER BY sort_order ASC, id ASC
`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PayinProductAdmin
	for rows.Next() {
		var p model.PayinProductAdmin
		var en int
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en, &p.ProductConfig); err != nil {
			return nil, err
		}
		p.Enabled = en == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// AdminGetPayinProduct 按 ID。
func (s *PayinProductsStore) AdminGetPayinProduct(ctx context.Context, id int64) (*model.PayinProductAdmin, error) {
	var r struct {
		ID        int64  `gorm:"column:id"`
		Code      string `gorm:"column:code"`
		Name      string `gorm:"column:name"`
		SortOrder int64  `gorm:"column:sort_order"`
		Enabled         int    `gorm:"column:enabled"`
		ProductConfig   string `gorm:"column:product_config"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT id, code, name, sort_order, enabled, COALESCE(product_config,'') AS product_config FROM payin_products WHERE id = ? LIMIT 1
`, id).Scan(&r)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &model.PayinProductAdmin{
		ID:            r.ID,
		Code:          r.Code,
		Name:          r.Name,
		SortOrder:     r.SortOrder,
		Enabled:       r.Enabled == 1,
		ProductConfig: r.ProductConfig,
	}, nil
}

// AdminCreatePayinProduct 新建；code 唯一。
func (s *PayinProductsStore) AdminCreatePayinProduct(ctx context.Context, code, name string, sortOrder int64, enabled bool, productConfig string) (int64, error) {
	en := 0
	if enabled {
		en = 1
	}
	tx := s.db.WithContext(ctx).Exec(`
INSERT INTO payin_products (code, name, sort_order, enabled, product_config) VALUES (?, ?, ?, ?, ?)
`, code, name, sortOrder, en, productConfig)
	if tx.Error != nil {
		return 0, tx.Error
	}
	var rid struct {
		ID int64 `gorm:"column:id"`
	}
	if err := s.db.WithContext(ctx).Raw(`SELECT LAST_INSERT_ID() AS id`).Scan(&rid).Error; err != nil {
		return 0, err
	}
	return rid.ID, nil
}

// AdminUpdatePayinProduct 更新。
func (s *PayinProductsStore) AdminUpdatePayinProduct(ctx context.Context, id int64, code, name string, sortOrder int64, enabled bool, productConfig string) error {
	en := 0
	if enabled {
		en = 1
	}
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payin_products SET code = ?, name = ?, sort_order = ?, enabled = ?, product_config = ?, updated_at = NOW() WHERE id = ?
`, code, name, sortOrder, en, productConfig, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// AdminListBindings 某产品下的通道绑定。
func (s *PayinProductsStore) AdminListBindings(ctx context.Context, payProductID int64) ([]model.PayinProductBindingAdmin, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT ppc.id, ppc.payin_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled
FROM payin_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.payin_product_id = ?
ORDER BY ppc.id ASC
`, payProductID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PayinProductBindingAdmin
	for rows.Next() {
		var b model.PayinProductBindingAdmin
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
	if err := s.db.WithContext(ctx).Exec(`
INSERT INTO payin_product_channels (payin_product_id, channel_id, weight, enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled), updated_at = NOW()
`, payProductID, channelID, weight, en).Error; err != nil {
		return 0, err
	}
	var bid int64
	var r struct {
		ID int64 `gorm:"column:id"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT id FROM payin_product_channels WHERE payin_product_id = ? AND channel_id = ? LIMIT 1
`, payProductID, channelID).Scan(&r)
	if tx.Error != nil {
		return 0, fmt.Errorf("load binding id: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return 0, fmt.Errorf("load binding id: %w", gorm.ErrRecordNotFound)
	}
	bid = r.ID
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
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payin_product_channels SET weight = ?, enabled = ?, updated_at = NOW() WHERE id = ?
`, weight, en, bindingID)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// AdminGetBindingByID 单条绑定（含通道名）。
func (s *PayinProductsStore) AdminGetBindingByID(ctx context.Context, bindingID int64) (*model.PayinProductBindingAdmin, error) {
	var b model.PayinProductBindingAdmin
	var r struct {
		ID             int64  `gorm:"column:id"`
		PayinProductID int64  `gorm:"column:payin_product_id"`
		ChannelID      int64  `gorm:"column:channel_id"`
		ChannelName    string `gorm:"column:channel_name"`
		Weight         int64  `gorm:"column:weight"`
		Enabled        int    `gorm:"column:enabled"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT ppc.id,
       ppc.payin_product_id,
       ppc.channel_id,
       COALESCE(c.name,'') AS channel_name,
       ppc.weight,
       ppc.enabled
FROM payin_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.id = ? LIMIT 1
`, bindingID).Scan(&r)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	b.ID = r.ID
	b.PayinProductID = r.PayinProductID
	b.ChannelID = r.ChannelID
	b.ChannelName = r.ChannelName
	b.Weight = r.Weight
	b.Enabled = r.Enabled == 1
	return &b, nil
}

// AdminDeleteBinding 删除一条绑定。
func (s *PayinProductsStore) AdminDeleteBinding(ctx context.Context, bindingID int64) error {
	tx := s.db.WithContext(ctx).Exec(`DELETE FROM payin_product_channels WHERE id = ?`, bindingID)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// AdminChannelSupportsPayin 通道是否存在且支持代收。
func (s *PayinProductsStore) AdminChannelSupportsPayin(ctx context.Context, channelID int64) (bool, error) {
	var r struct {
		SupportsPayin int `gorm:"column:supports_payin"`
	}
	tx := s.db.WithContext(ctx).
		Table("channels").
		Select("supports_payin").
		Where("id = ?", channelID).
		Limit(1).
		Take(&r)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, tx.Error
	}
	return r.SupportsPayin == 1, nil
}

// AdminChannelExists 通道是否存在。
func (s *PayinProductsStore) AdminChannelExists(ctx context.Context, channelID int64) (bool, error) {
	var one struct {
		One int `gorm:"column:one"`
	}
	tx := s.db.WithContext(ctx).
		Table("channels").
		Select("1 AS one").
		Where("id = ?", channelID).
		Limit(1).
		Take(&one)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, tx.Error
	}
	return true, nil
}
