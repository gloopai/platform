package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// PayinGrant 商户代收产品行（含可选覆盖费率）。
type PayinGrant struct {
	PayinProductID int64
	RateBps        *int64
}

type MerchantPayinProductsStore struct {
	db *gorm.DB
}

func NewMerchantPayinProductsStore(db *gorm.DB) *MerchantPayinProductsStore {
	return &MerchantPayinProductsStore{db: db}
}

func (s *MerchantPayinProductsStore) Replace(ctx context.Context, merchantID string, grants []PayinGrant) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return errors.New("merchant_id required")
	}
	seen := make(map[int64]struct{})
	var uniq []PayinGrant
	for _, g := range grants {
		if g.PayinProductID <= 0 {
			continue
		}
		if _, ok := seen[g.PayinProductID]; ok {
			continue
		}
		seen[g.PayinProductID] = struct{}{}
		uniq = append(uniq, g)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i].PayinProductID < uniq[j].PayinProductID })

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM merchant_payin_products WHERE merchant_id = ?`, merchantID).Error; err != nil {
			return err
		}
		for i, g := range uniq {
			if g.RateBps == nil {
				if err := tx.Exec(`
INSERT INTO merchant_payin_products (merchant_id, payin_product_id, enabled, sort_order, merchant_rate_bps)
VALUES (?, ?, 1, ?, NULL)
`, merchantID, g.PayinProductID, i).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Exec(`
INSERT INTO merchant_payin_products (merchant_id, payin_product_id, enabled, sort_order, merchant_rate_bps)
VALUES (?, ?, 1, ?, ?)
`, merchantID, g.PayinProductID, i, *g.RateBps).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *MerchantPayinProductsStore) ListProductIDs(ctx context.Context, merchantID string) ([]int64, error) {
	grants, err := s.ListPayinGrants(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(grants))
	for _, g := range grants {
		out = append(out, g.PayinProductID)
	}
	return out, nil
}

// ListPayinGrants 含费率覆盖（NULL 在 DB 中表示用商户默认代收费率）。
func (s *MerchantPayinProductsStore) ListPayinGrants(ctx context.Context, merchantID string) ([]PayinGrant, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	var out []PayinGrant
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT payin_product_id, merchant_rate_bps
FROM merchant_payin_products
WHERE merchant_id = ? AND enabled = 1
ORDER BY sort_order ASC, payin_product_id ASC
`, merchantID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g PayinGrant
		var rate sql.NullInt64
		if err := rows.Scan(&g.PayinProductID, &rate); err != nil {
			return nil, err
		}
		if rate.Valid {
			v := rate.Int64
			g.RateBps = &v
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
