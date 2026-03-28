package store

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/gloopai/pay/trade/internal/kvcache"
)

// PayinProductOption 收银台展示的支付产品（与 payin_products 一致，按订单金额过滤可用通道）。
type PayinProductOption struct {
	Code string
	Name string
}

type PayinProductsStore struct {
	db *gorm.DB
}

func NewPayinProductsStore(db *gorm.DB) *PayinProductsStore {
	return &PayinProductsStore{db: db}
}

// MerchantPayWhitelistStrict 为 true 表示该商户在 merchant_payin_products 中至少有一条启用记录，需按白名单约束产品与收银台展示。
// 为 false 表示未配置白名单：对外视为「开放模式」，收银台与下单校验使用全平台可用支付产品（仍受 payin_products / 通道限额约束）。
func (s *PayinProductsStore) MerchantPayWhitelistStrict(ctx context.Context, merchantID string) (bool, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return false, nil
	}
	var n int64
	if err := s.db.WithContext(ctx).
		Table("merchant_payin_products").
		Where("merchant_id = ? AND enabled = 1", merchantID).
		Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListTerminalPayinProducts 收银台用：已配置白名单则只展示白名单内可用产品；未配置则与 ListAvailableForAmount 一致。
func (s *PayinProductsStore) ListTerminalPayinProducts(ctx context.Context, merchantID string, amount int64) ([]PayinProductOption, error) {
	strict, err := s.MerchantPayWhitelistStrict(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if strict {
		return s.ListAvailableForMerchantAndAmount(ctx, merchantID, amount)
	}
	return s.ListAvailableForAmount(ctx, amount)
}

// ListAvailableForAmount 返回：至少有一条可用上游通道、且金额在通道限额内的支付产品。
func (s *PayinProductsStore) ListAvailableForAmount(ctx context.Context, amount int64) ([]PayinProductOption, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT DISTINCT pp.code, pp.name
FROM payin_products pp
INNER JOIN payin_product_channels ppc ON pp.id = ppc.payin_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND c.supports_payin = 1 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY pp.sort_order ASC, pp.id ASC
`, amount, amount).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayinProductOption
	for rows.Next() {
		var o PayinProductOption
		if err := rows.Scan(&o.Code, &o.Name); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) > 0 {
		return out, nil
	}
	// 未迁移 payin_products 时的回退：按 channels.payin_type 去重
	return s.listLegacyChannelPayinTypes(ctx, amount)
}

func (s *PayinProductsStore) listLegacyChannelPayinTypes(ctx context.Context, amount int64) ([]PayinProductOption, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT DISTINCT COALESCE(NULLIF(TRIM(payin_type), ''), 'mock')
FROM channels
WHERE enabled = 1 AND fuse_enabled = 0 AND supports_payin = 1 AND weight > 0
  AND (min_amount = 0 OR min_amount <= ?)
  AND (max_amount = 0 OR max_amount >= ?)
ORDER BY 1
`, amount, amount).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayinProductOption
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, PayinProductOption{Code: code, Name: code})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAvailableForMerchantAndAmount 商户白名单内、且金额满足通道限额的支付产品（未配置白名单时请用 ListTerminalPayinProducts）。
func (s *PayinProductsStore) ListAvailableForMerchantAndAmount(ctx context.Context, merchantID string, amount int64) ([]PayinProductOption, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT DISTINCT pp.code, pp.name
FROM payin_products pp
INNER JOIN merchant_payin_products mpp ON mpp.payin_product_id = pp.id AND mpp.merchant_id = ? AND mpp.enabled = 1
INNER JOIN payin_product_channels ppc ON pp.id = ppc.payin_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND c.supports_payin = 1 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY mpp.sort_order ASC, pp.sort_order ASC, pp.id ASC
`, merchantID, amount, amount).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayinProductOption
	for rows.Next() {
		var o PayinProductOption
		if err := rows.Scan(&o.Code, &o.Name); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// MerchantHasPayinProductCode 判断商户是否可使用该支付产品编码；未配置白名单时视为开放模式（任意有效产品编码在后续路由中再校验）。
func (s *PayinProductsStore) MerchantHasPayinProductCode(ctx context.Context, merchantID, payinProductCode string) (bool, error) {
	merchantID = strings.TrimSpace(merchantID)
	code := strings.TrimSpace(payinProductCode)
	if merchantID == "" || code == "" {
		return false, nil
	}
	strict, err := s.MerchantPayWhitelistStrict(ctx, merchantID)
	if err != nil {
		return false, err
	}
	if !strict {
		return true, nil
	}
	var one struct {
		One int `gorm:"column:one"`
	}
	tx := s.db.WithContext(ctx).
		Table("merchant_payin_products mpp").
		Select("1 AS one").
		Joins("INNER JOIN payin_products pp ON pp.id = mpp.payin_product_id AND pp.enabled = 1").
		Where("mpp.merchant_id = ? AND mpp.enabled = 1 AND pp.code = ?", merchantID, code).
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

// ResolveLockedChannelForMerchant 商户 API 指定 channel_id 时，解析对应的 payin_product（白名单开启时通道须落在白名单关联的产品上；开放模式仅校验通道与金额）。
func (s *PayinProductsStore) ResolveLockedChannelForMerchant(ctx context.Context, merchantID string, channelID int64, amount int64) (payProductID int64, payinProductCode string, err error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" || channelID <= 0 {
		return 0, "", errors.New("merchant_id and channel_id required")
	}
	strict, err := s.MerchantPayWhitelistStrict(ctx, merchantID)
	if err != nil {
		return 0, "", err
	}
	type row struct {
		PayProductID int64  `gorm:"column:pay_product_id"`
		Code         string `gorm:"column:code"`
	}
	var r row
	q := s.db.WithContext(ctx).
		Table("payin_product_channels ppc").
		Select("pp.id AS pay_product_id, pp.code AS code").
		Joins("INNER JOIN payin_products pp ON pp.id = ppc.payin_product_id AND pp.enabled = 1").
		Joins("INNER JOIN channels c ON c.id = ppc.channel_id").
		Where("ppc.channel_id = ? AND ppc.enabled = 1", channelID).
		Where("c.enabled = 1 AND c.fuse_enabled = 0 AND c.supports_payin = 1 AND ppc.weight > 0").
		Where("(c.min_amount = 0 OR c.min_amount <= ?)", amount).
		Where("(c.max_amount = 0 OR c.max_amount >= ?)", amount)
	if strict {
		q = q.Joins("INNER JOIN merchant_payin_products mpp ON mpp.payin_product_id = pp.id AND mpp.merchant_id = ? AND mpp.enabled = 1", merchantID)
	}
	tx := q.Order("ppc.weight DESC, pp.id ASC").Limit(1).Take(&r)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return 0, "", errors.New("channel not allowed for merchant or amount out of range")
		}
		return 0, "", tx.Error
	}
	return r.PayProductID, r.Code, nil
}

// GetPayinProductDisplayName 按 code 取展示名；优先 Consul 内存中的 product_config（display_name），否则库表 name，不存在则返回 code。
func (s *PayinProductsStore) GetPayinProductDisplayName(ctx context.Context, code string, cfgCache *kvcache.PayinProductConfig) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", nil
	}
	var r struct {
		ID            int64  `gorm:"column:id"`
		Name          string `gorm:"column:name"`
		ProductConfig string `gorm:"column:product_config"`
	}
	tx := s.db.WithContext(ctx).
		Table("payin_products").
		Select("id, COALESCE(NULLIF(TRIM(name), ''), code) AS name, COALESCE(product_config,'') AS product_config").
		Where("code = ? AND enabled = 1", code).
		Limit(1).
		Take(&r)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return code, nil
		}
		return "", tx.Error
	}
	merged := kvcache.PickPayinProductConfig(cfgCache, r.ID, r.ProductConfig)
	if dn := kvcache.DisplayNameFromProductJSON(merged); dn != "" {
		return dn, nil
	}
	return r.Name, nil
}
