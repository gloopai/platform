package store

import (
	"context"
	"database/sql"
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
