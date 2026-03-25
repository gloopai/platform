package store

import (
	"context"
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
	var n int64
	if err := s.db.WithContext(ctx).
		Table("merchant_payin_products").
		Where("merchant_id = ? AND enabled = 1", merchantID).
		Count(&n).Error; err != nil {
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
