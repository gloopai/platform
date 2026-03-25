package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// PayoutGrant 商户代付产品行。
type PayoutGrant struct {
	PayoutProductID int64
	FeeMode         int64
	RateBps         *int64
	FixedFeeAmount  int64
}

type MerchantPayoutProductsStore struct {
	db *gorm.DB
}

func NewMerchantPayoutProductsStore(db *gorm.DB) *MerchantPayoutProductsStore {
	return &MerchantPayoutProductsStore{db: db}
}

func (s *MerchantPayoutProductsStore) Replace(ctx context.Context, merchantID string, grants []PayoutGrant) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return errors.New("merchant_id required")
	}
	seen := make(map[int64]struct{})
	var uniq []PayoutGrant
	for _, g := range grants {
		if g.PayoutProductID <= 0 {
			continue
		}
		if _, ok := seen[g.PayoutProductID]; ok {
			continue
		}
		seen[g.PayoutProductID] = struct{}{}
		uniq = append(uniq, g)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i].PayoutProductID < uniq[j].PayoutProductID })

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM merchant_payout_products WHERE merchant_id = ?`, merchantID).Error; err != nil {
			return err
		}
		for i, g := range uniq {
			feeMode := g.FeeMode
			if feeMode < 1 || feeMode > 3 {
				feeMode = 1
			}
			if g.RateBps == nil {
				if err := tx.Exec(`
INSERT INTO merchant_payout_products (merchant_id, payout_product_id, enabled, sort_order, fee_mode, merchant_rate_bps, fee_fixed_amount)
VALUES (?, ?, 1, ?, ?, NULL, ?)
`, merchantID, g.PayoutProductID, i, feeMode, g.FixedFeeAmount).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Exec(`
INSERT INTO merchant_payout_products (merchant_id, payout_product_id, enabled, sort_order, fee_mode, merchant_rate_bps, fee_fixed_amount)
VALUES (?, ?, 1, ?, ?, ?, ?)
`, merchantID, g.PayoutProductID, i, feeMode, *g.RateBps, g.FixedFeeAmount).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *MerchantPayoutProductsStore) ListPayoutGrants(ctx context.Context, merchantID string) ([]PayoutGrant, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	rows, err := s.db.WithContext(ctx).Raw(`
SELECT payout_product_id, fee_mode, merchant_rate_bps, fee_fixed_amount
FROM merchant_payout_products
WHERE merchant_id = ? AND enabled = 1
ORDER BY sort_order ASC, payout_product_id ASC
`, merchantID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayoutGrant
	for rows.Next() {
		var g PayoutGrant
		var rate sql.NullInt64
		if err := rows.Scan(&g.PayoutProductID, &g.FeeMode, &rate, &g.FixedFeeAmount); err != nil {
			return nil, err
		}
		if g.FeeMode < 1 || g.FeeMode > 3 {
			g.FeeMode = 1
		}
		if rate.Valid {
			v := rate.Int64
			g.RateBps = &v
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *MerchantPayoutProductsStore) ListPayoutProductIDs(ctx context.Context, merchantID string) ([]int64, error) {
	grants, err := s.ListPayoutGrants(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(grants))
	for _, g := range grants {
		out = append(out, g.PayoutProductID)
	}
	return out, nil
}
