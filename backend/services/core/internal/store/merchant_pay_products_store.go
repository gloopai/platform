package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
)

// CollectGrant 商户代收产品行（含可选覆盖费率）。
type CollectGrant struct {
	PayProductID int64
	RateBps      *int64
}

type MerchantPayProductsStore struct {
	db *sql.DB
}

func NewMerchantPayProductsStore(db *sql.DB) *MerchantPayProductsStore {
	return &MerchantPayProductsStore{db: db}
}

func (s *MerchantPayProductsStore) Replace(ctx context.Context, merchantID string, grants []CollectGrant) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return errors.New("merchant_id required")
	}
	seen := make(map[int64]struct{})
	var uniq []CollectGrant
	for _, g := range grants {
		if g.PayProductID <= 0 {
			continue
		}
		if _, ok := seen[g.PayProductID]; ok {
			continue
		}
		seen[g.PayProductID] = struct{}{}
		uniq = append(uniq, g)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i].PayProductID < uniq[j].PayProductID })

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM merchant_pay_products WHERE merchant_id = ?`, merchantID); err != nil {
		_ = tx.Rollback()
		return err
	}
	for i, g := range uniq {
		if g.RateBps == nil {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO merchant_pay_products (merchant_id, pay_product_id, enabled, sort_order, merchant_rate_bps)
VALUES (?, ?, 1, ?, NULL)
`, merchantID, g.PayProductID, i); err != nil {
				_ = tx.Rollback()
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO merchant_pay_products (merchant_id, pay_product_id, enabled, sort_order, merchant_rate_bps)
VALUES (?, ?, 1, ?, ?)
`, merchantID, g.PayProductID, i, *g.RateBps); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *MerchantPayProductsStore) ListProductIDs(ctx context.Context, merchantID string) ([]int64, error) {
	grants, err := s.ListCollectGrants(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(grants))
	for _, g := range grants {
		out = append(out, g.PayProductID)
	}
	return out, nil
}

// ListCollectGrants 含费率覆盖（NULL 在 DB 中表示用商户默认代收费率）。
func (s *MerchantPayProductsStore) ListCollectGrants(ctx context.Context, merchantID string) ([]CollectGrant, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT pay_product_id, merchant_rate_bps
FROM merchant_pay_products
WHERE merchant_id = ? AND enabled = 1
ORDER BY sort_order ASC, pay_product_id ASC
`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CollectGrant
	for rows.Next() {
		var g CollectGrant
		var rate sql.NullInt64
		if err := rows.Scan(&g.PayProductID, &rate); err != nil {
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
