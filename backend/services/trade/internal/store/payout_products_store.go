package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gloopai/pay/common/model"
	"gorm.io/gorm"
)

// PayoutProductsStore 代付对外产品与通道绑定（管理台与后续代付 API 共用库访问）。
type PayoutProductsStore struct {
	db *gorm.DB
}

func NewPayoutProductsStore(db *gorm.DB) *PayoutProductsStore {
	return &PayoutProductsStore{db: db}
}

func (s *PayoutProductsStore) AdminListAllPayoutProducts(ctx context.Context) ([]model.PayoutProductAdmin, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT id, code, name, sort_order, enabled, COALESCE(product_config,'') AS product_config FROM payout_products ORDER BY sort_order ASC, id ASC
`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PayoutProductAdmin
	for rows.Next() {
		var p model.PayoutProductAdmin
		var en int
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &en, &p.ProductConfig); err != nil {
			return nil, err
		}
		p.Enabled = en == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PayoutProductsStore) AdminGetPayoutProduct(ctx context.Context, id int64) (*model.PayoutProductAdmin, error) {
	var r struct {
		ID        int64  `gorm:"column:id"`
		Code      string `gorm:"column:code"`
		Name      string `gorm:"column:name"`
		SortOrder int64  `gorm:"column:sort_order"`
		Enabled         int    `gorm:"column:enabled"`
		ProductConfig   string `gorm:"column:product_config"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT id, code, name, sort_order, enabled, COALESCE(product_config,'') AS product_config FROM payout_products WHERE id = ? LIMIT 1
`, id).Scan(&r)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &model.PayoutProductAdmin{ID: r.ID, Code: r.Code, Name: r.Name, SortOrder: r.SortOrder, Enabled: r.Enabled == 1, ProductConfig: r.ProductConfig}, nil
}

func (s *PayoutProductsStore) AdminCreatePayoutProduct(ctx context.Context, code, name string, sortOrder int64, enabled bool, productConfig string) (int64, error) {
	en := 0
	if enabled {
		en = 1
	}
	tx := s.db.WithContext(ctx).Exec(`
INSERT INTO payout_products (code, name, sort_order, enabled, product_config) VALUES (?, ?, ?, ?, ?)
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

func (s *PayoutProductsStore) AdminUpdatePayoutProduct(ctx context.Context, id int64, code, name string, sortOrder int64, enabled bool, productConfig string) error {
	en := 0
	if enabled {
		en = 1
	}
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payout_products SET code = ?, name = ?, sort_order = ?, enabled = ?, product_config = ?, updated_at = NOW() WHERE id = ?
`, code, name, sortOrder, en, productConfig, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *PayoutProductsStore) AdminListPayoutBindings(ctx context.Context, payoutProductID int64) ([]model.PayoutProductBindingAdmin, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT ppc.id, ppc.payout_product_id, ppc.channel_id, COALESCE(c.name,''), ppc.weight, ppc.enabled
FROM payout_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.payout_product_id = ?
ORDER BY ppc.id ASC
`, payoutProductID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PayoutProductBindingAdmin
	for rows.Next() {
		var b model.PayoutProductBindingAdmin
		var en int
		if err := rows.Scan(&b.ID, &b.PayoutProductID, &b.ChannelID, &b.ChannelName, &b.Weight, &en); err != nil {
			return nil, err
		}
		b.Enabled = en == 1
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *PayoutProductsStore) AdminUpsertPayoutBinding(ctx context.Context, payoutProductID, channelID int64, weight int64, enabled bool) (int64, error) {
	if weight <= 0 {
		return 0, errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	if err := s.db.WithContext(ctx).Exec(`
INSERT INTO payout_product_channels (payout_product_id, channel_id, weight, enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled), updated_at = NOW()
`, payoutProductID, channelID, weight, en).Error; err != nil {
		return 0, err
	}
	var bid int64
	var r struct {
		ID int64 `gorm:"column:id"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT id FROM payout_product_channels WHERE payout_product_id = ? AND channel_id = ? LIMIT 1
`, payoutProductID, channelID).Scan(&r)
	if tx.Error != nil {
		return 0, fmt.Errorf("load binding id: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return 0, fmt.Errorf("load binding id: %w", gorm.ErrRecordNotFound)
	}
	bid = r.ID
	return bid, nil
}

func (s *PayoutProductsStore) AdminUpdatePayoutBinding(ctx context.Context, bindingID int64, weight int64, enabled bool) error {
	if weight <= 0 {
		return errors.New("weight must be positive")
	}
	en := 0
	if enabled {
		en = 1
	}
	tx := s.db.WithContext(ctx).Exec(`
UPDATE payout_product_channels SET weight = ?, enabled = ?, updated_at = NOW() WHERE id = ?
`, weight, en, bindingID)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *PayoutProductsStore) AdminGetPayoutBindingByID(ctx context.Context, bindingID int64) (*model.PayoutProductBindingAdmin, error) {
	var r struct {
		ID              int64  `gorm:"column:id"`
		PayoutProductID int64  `gorm:"column:payout_product_id"`
		ChannelID       int64  `gorm:"column:channel_id"`
		ChannelName     string `gorm:"column:channel_name"`
		Weight          int64  `gorm:"column:weight"`
		Enabled         int    `gorm:"column:enabled"`
	}
	tx := s.db.WithContext(ctx).Raw(`
SELECT ppc.id,
       ppc.payout_product_id,
       ppc.channel_id,
       COALESCE(c.name,'') AS channel_name,
       ppc.weight,
       ppc.enabled
FROM payout_product_channels ppc
LEFT JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.id = ? LIMIT 1
`, bindingID).Scan(&r)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &model.PayoutProductBindingAdmin{
		ID:              r.ID,
		PayoutProductID: r.PayoutProductID,
		ChannelID:       r.ChannelID,
		ChannelName:     r.ChannelName,
		Weight:          r.Weight,
		Enabled:         r.Enabled == 1,
	}, nil
}

func (s *PayoutProductsStore) AdminDeletePayoutBinding(ctx context.Context, bindingID int64) error {
	tx := s.db.WithContext(ctx).Exec(`DELETE FROM payout_product_channels WHERE id = ?`, bindingID)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *PayoutProductsStore) AdminChannelSupportsPayout(ctx context.Context, channelID int64) (bool, error) {
	var r struct {
		SupportsPayout int `gorm:"column:supports_payout"`
	}
	tx := s.db.WithContext(ctx).
		Table("channels").
		Select("supports_payout").
		Where("id = ?", channelID).
		Limit(1).
		Take(&r)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, tx.Error
	}
	return r.SupportsPayout == 1, nil
}
