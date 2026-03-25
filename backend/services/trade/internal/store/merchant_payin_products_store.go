package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// MerchantPayinProductsStore 与 gateway 侧一致：用于校验商户是否可使用某支付产品（白名单 / 开放模式）。
type MerchantPayinProductsStore struct {
	db *gorm.DB
}

func NewMerchantPayinProductsStore(db *gorm.DB) *MerchantPayinProductsStore {
	return &MerchantPayinProductsStore{db: db}
}

// MerchantPayWhitelistStrict 见 gateway PayinProductsStore.MerchantPayWhitelistStrict。
func (s *MerchantPayinProductsStore) MerchantPayWhitelistStrict(ctx context.Context, merchantID string) (bool, error) {
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

// MerchantHasPayinProductCode 与 gateway 侧逻辑对齐：未配置白名单则开放；已配置则必须在白名单内。
func (s *MerchantPayinProductsStore) MerchantHasPayinProductCode(ctx context.Context, merchantID, payinProductCode string) (bool, error) {
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
