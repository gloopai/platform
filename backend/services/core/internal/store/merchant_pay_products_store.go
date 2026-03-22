package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
)

type MerchantPayProductsStore struct {
	db *sql.DB
}

func NewMerchantPayProductsStore(db *sql.DB) *MerchantPayProductsStore {
	return &MerchantPayProductsStore{db: db}
}

func (s *MerchantPayProductsStore) Replace(ctx context.Context, merchantID string, productIDs []int64) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return errors.New("merchant_id required")
	}
	seen := make(map[int64]struct{})
	var uniq []int64
	for _, id := range productIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i] < uniq[j] })

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM merchant_pay_products WHERE merchant_id = ?`, merchantID); err != nil {
		_ = tx.Rollback()
		return err
	}
	for i, pid := range uniq {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO merchant_pay_products (merchant_id, pay_product_id, enabled, sort_order)
VALUES (?, ?, 1, ?)
`, merchantID, pid, i); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *MerchantPayProductsStore) ListProductIDs(ctx context.Context, merchantID string) ([]int64, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT pay_product_id
FROM merchant_pay_products
WHERE merchant_id = ? AND enabled = 1
ORDER BY sort_order ASC, pay_product_id ASC
`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
