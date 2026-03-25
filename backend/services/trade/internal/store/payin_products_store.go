package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"
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
	var n int
	err := s.db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM merchant_payin_products WHERE merchant_id = ? AND enabled = 1
`, merchantID).Row().Scan(&n)
	if err != nil {
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
	var one int
	err = s.db.WithContext(ctx).Raw(`
SELECT 1
FROM merchant_payin_products mpp
INNER JOIN payin_products pp ON pp.id = mpp.payin_product_id AND pp.enabled = 1
WHERE mpp.merchant_id = ? AND mpp.enabled = 1 AND pp.code = ?
LIMIT 1
`, merchantID, code).Row().Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
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
	if strict {
		err = s.db.WithContext(ctx).Raw(`
SELECT pp.id, pp.code
FROM payin_product_channels ppc
INNER JOIN payin_products pp ON pp.id = ppc.payin_product_id AND pp.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
INNER JOIN merchant_payin_products mpp ON mpp.payin_product_id = pp.id AND mpp.merchant_id = ? AND mpp.enabled = 1
WHERE ppc.channel_id = ? AND ppc.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND c.supports_payin = 1 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY ppc.weight DESC, pp.id ASC
LIMIT 1
`, merchantID, channelID, amount, amount).Row().Scan(&payProductID, &payinProductCode)
	} else {
		err = s.db.WithContext(ctx).Raw(`
SELECT pp.id, pp.code
FROM payin_product_channels ppc
INNER JOIN payin_products pp ON pp.id = ppc.payin_product_id AND pp.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE ppc.channel_id = ? AND ppc.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND c.supports_payin = 1 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY ppc.weight DESC, pp.id ASC
LIMIT 1
`, channelID, amount, amount).Row().Scan(&payProductID, &payinProductCode)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", errors.New("channel not allowed for merchant or amount out of range")
		}
		return 0, "", err
	}
	return payProductID, payinProductCode, nil
}

// GetPayinProductDisplayName 按 code 取展示名，不存在则返回 code。
func (s *PayinProductsStore) GetPayinProductDisplayName(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", nil
	}
	var name string
	err := s.db.WithContext(ctx).Raw(`
SELECT COALESCE(NULLIF(TRIM(name), ''), code) FROM payin_products WHERE code = ? AND enabled = 1 LIMIT 1
`, code).Row().Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return code, nil
		}
		return "", err
	}
	return name, nil
}
