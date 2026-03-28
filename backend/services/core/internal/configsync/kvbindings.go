package configsync

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/core/internal/store"
)

// SyncMerchantPayinGrants writes merchant payin grants to Consul KV (empty list deletes the key).
func SyncMerchantPayinGrants(ctx context.Context, cfg *consulx.ConfigStore, st *store.MerchantPayinProductsStore, merchantID string) error {
	if cfg == nil || st == nil {
		return nil
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil
	}
	key := configkv.MerchantPayinGrantsKVKey(merchantID)
	if key == "" {
		return nil
	}
	grants, err := st.ListPayinGrants(ctx, merchantID)
	if err != nil {
		return err
	}
	if len(grants) == 0 {
		return cfg.Delete(ctx, key)
	}
	rows := make([]configkv.PayinGrantKV, 0, len(grants))
	for _, g := range grants {
		rows = append(rows, configkv.PayinGrantKV{
			PayinProductID:  g.PayinProductID,
			MerchantRateBps: g.RateBps,
		})
	}
	return cfg.PutJSON(ctx, key, &configkv.MerchantPayinGrantsKV{MerchantID: merchantID, Grants: rows})
}

// SyncMerchantPayoutGrants writes merchant payout grants to Consul KV (empty list deletes the key).
func SyncMerchantPayoutGrants(ctx context.Context, cfg *consulx.ConfigStore, st *store.MerchantPayoutProductsStore, merchantID string) error {
	if cfg == nil || st == nil {
		return nil
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil
	}
	key := configkv.MerchantPayoutGrantsKVKey(merchantID)
	if key == "" {
		return nil
	}
	grants, err := st.ListPayoutGrants(ctx, merchantID)
	if err != nil {
		return err
	}
	if len(grants) == 0 {
		return cfg.Delete(ctx, key)
	}
	rows := make([]configkv.PayoutGrantKV, 0, len(grants))
	for _, g := range grants {
		feeMode := g.FeeMode
		if feeMode < 1 || feeMode > 3 {
			feeMode = 1
		}
		rows = append(rows, configkv.PayoutGrantKV{
			PayoutProductID: g.PayoutProductID,
			FeeMode:         feeMode,
			MerchantRateBps: g.RateBps,
			FeeFixedAmount:  g.FixedFeeAmount,
		})
	}
	return cfg.PutJSON(ctx, key, &configkv.MerchantPayoutGrantsKV{MerchantID: merchantID, Grants: rows})
}

// SyncPayinProductChannelBindings writes all channel bindings for a payin product to Consul KV.
func SyncPayinProductChannelBindings(ctx context.Context, cfg *consulx.ConfigStore, st *store.PayinProductsStore, payinProductID int64) error {
	if cfg == nil || st == nil || payinProductID <= 0 {
		return nil
	}
	key := configkv.PayinProductChannelBindingsKVKey(payinProductID)
	if key == "" {
		return nil
	}
	rows, err := st.AdminListBindings(ctx, payinProductID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return cfg.Delete(ctx, key)
	}
	binds := make([]configkv.PayinProductChannelKV, 0, len(rows))
	for i := range rows {
		r := rows[i]
		binds = append(binds, configkv.PayinProductChannelKV{
			ID:        r.ID,
			ChannelID: r.ChannelID,
			Weight:    r.Weight,
			Enabled:   r.Enabled,
		})
	}
	return cfg.PutJSON(ctx, key, &configkv.PayinProductBindingsKV{PayinProductID: payinProductID, Bindings: binds})
}

// SyncPayoutProductChannelBindings writes all channel bindings for a payout product to Consul KV.
func SyncPayoutProductChannelBindings(ctx context.Context, cfg *consulx.ConfigStore, st *store.PayoutProductsStore, payoutProductID int64) error {
	if cfg == nil || st == nil || payoutProductID <= 0 {
		return nil
	}
	key := configkv.PayoutProductChannelBindingsKVKey(payoutProductID)
	if key == "" {
		return nil
	}
	rows, err := st.AdminListPayoutBindings(ctx, payoutProductID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return cfg.Delete(ctx, key)
	}
	binds := make([]configkv.PayoutProductChannelKV, 0, len(rows))
	for i := range rows {
		r := rows[i]
		binds = append(binds, configkv.PayoutProductChannelKV{
			ID:        r.ID,
			ChannelID: r.ChannelID,
			Weight:    r.Weight,
			Enabled:   r.Enabled,
		})
	}
	return cfg.PutJSON(ctx, key, &configkv.PayoutProductBindingsKV{PayoutProductID: payoutProductID, Bindings: binds})
}
