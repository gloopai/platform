package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// PayProductOption 收银台展示的支付产品（与 pay_products 一致，按订单金额过滤可用通道）。
type PayProductOption struct {
	Code string
	Name string
}

type PayProductsStore struct {
	db *sql.DB
}

func NewPayProductsStore(db *sql.DB) *PayProductsStore {
	return &PayProductsStore{db: db}
}

// ListAvailableForAmount 返回：至少有一条可用上游通道、且金额在通道限额内的支付产品。
func (s *PayProductsStore) ListAvailableForAmount(ctx context.Context, amount int64) ([]PayProductOption, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT pp.code, pp.name
FROM pay_products pp
INNER JOIN pay_product_channels ppc ON pp.id = ppc.pay_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY pp.sort_order ASC, pp.id ASC
`, amount, amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayProductOption
	for rows.Next() {
		var o PayProductOption
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
	// 未迁移 pay_products 时的回退：按 channels.pay_type 去重
	return s.listLegacyChannelPayTypes(ctx, amount)
}

func (s *PayProductsStore) listLegacyChannelPayTypes(ctx context.Context, amount int64) ([]PayProductOption, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT COALESCE(NULLIF(TRIM(pay_type), ''), 'mock')
FROM channels
WHERE enabled = 1 AND fuse_enabled = 0 AND weight > 0
  AND (min_amount = 0 OR min_amount <= ?)
  AND (max_amount = 0 OR max_amount >= ?)
ORDER BY 1
`, amount, amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayProductOption
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, PayProductOption{Code: code, Name: code})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAvailableForMerchantAndAmount 商户白名单内、且金额满足通道限额的支付产品（无白名单配置时返回空）。
func (s *PayProductsStore) ListAvailableForMerchantAndAmount(ctx context.Context, merchantID string, amount int64) ([]PayProductOption, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT pp.code, pp.name
FROM pay_products pp
INNER JOIN merchant_pay_products mpp ON mpp.pay_product_id = pp.id AND mpp.merchant_id = ? AND mpp.enabled = 1
INNER JOIN pay_product_channels ppc ON pp.id = ppc.pay_product_id AND ppc.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
WHERE pp.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY mpp.sort_order ASC, pp.sort_order ASC, pp.id ASC
`, merchantID, amount, amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayProductOption
	for rows.Next() {
		var o PayProductOption
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

// MerchantHasPayProductCode 判断商户是否被分配了该支付产品编码。
func (s *PayProductsStore) MerchantHasPayProductCode(ctx context.Context, merchantID, payProductCode string) (bool, error) {
	merchantID = strings.TrimSpace(merchantID)
	code := strings.TrimSpace(payProductCode)
	if merchantID == "" || code == "" {
		return false, nil
	}
	var one int
	err := s.db.QueryRowContext(ctx, `
SELECT 1
FROM merchant_pay_products mpp
INNER JOIN pay_products pp ON pp.id = mpp.pay_product_id AND pp.enabled = 1
WHERE mpp.merchant_id = ? AND mpp.enabled = 1 AND pp.code = ?
LIMIT 1
`, merchantID, code).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ResolveLockedChannelForMerchant 商户 API 指定 channel_id 时，解析对应的 pay_product（须在白名单内且金额可用）。
func (s *PayProductsStore) ResolveLockedChannelForMerchant(ctx context.Context, merchantID string, channelID int64, amount int64) (payProductID int64, payProductCode string, err error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" || channelID <= 0 {
		return 0, "", errors.New("merchant_id and channel_id required")
	}
	err = s.db.QueryRowContext(ctx, `
SELECT pp.id, pp.code
FROM pay_product_channels ppc
INNER JOIN pay_products pp ON pp.id = ppc.pay_product_id AND pp.enabled = 1
INNER JOIN channels c ON c.id = ppc.channel_id
INNER JOIN merchant_pay_products mpp ON mpp.pay_product_id = pp.id AND mpp.merchant_id = ? AND mpp.enabled = 1
WHERE ppc.channel_id = ? AND ppc.enabled = 1
  AND c.enabled = 1 AND c.fuse_enabled = 0 AND ppc.weight > 0
  AND (c.min_amount = 0 OR c.min_amount <= ?)
  AND (c.max_amount = 0 OR c.max_amount >= ?)
ORDER BY ppc.weight DESC, pp.id ASC
LIMIT 1
`, merchantID, channelID, amount, amount).Scan(&payProductID, &payProductCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", errors.New("channel not allowed for merchant or amount out of range")
		}
		return 0, "", err
	}
	return payProductID, payProductCode, nil
}

// GetPayProductDisplayName 按 code 取展示名，不存在则返回 code。
func (s *PayProductsStore) GetPayProductDisplayName(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", nil
	}
	var name string
	err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(NULLIF(TRIM(name), ''), code) FROM pay_products WHERE code = ? AND enabled = 1 LIMIT 1
`, code).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return code, nil
		}
		return "", err
	}
	return name, nil
}
