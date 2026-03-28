package kvcache

import (
	"errors"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/model"
)

type routePick struct {
	ChannelID      int64
	Weight         int64
	PayinProductID int64
}

// ChannelEligiblePayin returns true when channel snapshot supports payin routing for amount.
func ChannelEligiblePayin(ch *configkv.ChannelKV, amount int64) bool {
	if ch == nil || !ch.Enabled || ch.FuseEnabled || !ch.SupportsPayin || ch.Weight <= 0 {
		return false
	}
	if ch.MinAmount > 0 && ch.MinAmount > amount {
		return false
	}
	if ch.MaxAmount > 0 && ch.MaxAmount < amount {
		return false
	}
	return true
}

// MerchantPayWhitelistStrictMemory true when merchant has at least one payin grant in Consul snapshot.
func MerchantPayWhitelistStrictMemory(grantsSnap *MerchantPayinGrantsSnapshot, merchantID string) bool {
	if grantsSnap == nil {
		return false
	}
	g, ok := grantsSnap.Get(merchantID)
	return ok && g != nil && len(g.Grants) > 0
}

// ListAvailableForAmountMemory lists payin product code/name using only Consul snapshots.
func ListAvailableForAmountMemory(
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
	amount int64,
) []model.PayinProductOption {
	if payinSnap == nil || bindSnap == nil || chSnap == nil {
		return nil
	}
	type pair struct {
		sortOrder int64
		id        int64
		code      string
		name      string
	}
	var rows []pair
	payinSnap.ForEach(func(id int64, p *configkv.PayinProductKV) {
		if p == nil || !p.Enabled {
			return
		}
		if !productHasRoutableBinding(id, bindSnap, chSnap, amount) {
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = strings.TrimSpace(p.Code)
		}
		rows = append(rows, pair{sortOrder: p.SortOrder, id: id, code: strings.TrimSpace(p.Code), name: name})
	})
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].sortOrder != rows[j].sortOrder {
			return rows[i].sortOrder < rows[j].sortOrder
		}
		return rows[i].id < rows[j].id
	})
	out := make([]model.PayinProductOption, 0, len(rows))
	for _, r := range rows {
		if r.code != "" {
			out = append(out, model.PayinProductOption{Code: r.code, Name: r.name})
		}
	}
	if len(out) > 0 {
		return out
	}
	return listLegacyChannelPayinTypesMemory(chSnap, amount)
}

func productHasRoutableBinding(payinProductID int64, bindSnap *PayinProductBindingsSnapshot, chSnap *ChannelSnapshot, amount int64) bool {
	blob, ok := bindSnap.Get(payinProductID)
	if !ok || blob == nil {
		return false
	}
	for i := range blob.Bindings {
		b := blob.Bindings[i]
		if !b.Enabled || b.Weight <= 0 {
			continue
		}
		ch, ok := chSnap.Get(b.ChannelID)
		if !ok || !ChannelEligiblePayin(ch, amount) {
			continue
		}
		return true
	}
	return false
}

// ListAvailableForMerchantAndAmountMemory lists products for merchant whitelist using snapshots only.
func ListAvailableForMerchantAndAmountMemory(
	merchantID string,
	grantsSnap *MerchantPayinGrantsSnapshot,
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
	amount int64,
) []model.PayinProductOption {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil
	}
	gkv, ok := grantsSnap.Get(merchantID)
	if !ok || gkv == nil || len(gkv.Grants) == 0 {
		return nil
	}
	type pair struct {
		sortOrder int64
		id        int64
		code      string
		name      string
		ord       int
	}
	var rows []pair
	for ord, g := range gkv.Grants {
		id := g.PayinProductID
		if id <= 0 {
			continue
		}
		p, ok := payinSnap.Get(id)
		if !ok || p == nil || !p.Enabled {
			continue
		}
		if !productHasRoutableBinding(id, bindSnap, chSnap, amount) {
			continue
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = strings.TrimSpace(p.Code)
		}
		rows = append(rows, pair{sortOrder: p.SortOrder, id: id, code: strings.TrimSpace(p.Code), name: name, ord: ord})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ord != rows[j].ord {
			return rows[i].ord < rows[j].ord
		}
		if rows[i].sortOrder != rows[j].sortOrder {
			return rows[i].sortOrder < rows[j].sortOrder
		}
		return rows[i].id < rows[j].id
	})
	out := make([]model.PayinProductOption, 0, len(rows))
	for _, r := range rows {
		if r.code != "" {
			out = append(out, model.PayinProductOption{Code: r.code, Name: r.name})
		}
	}
	return out
}

func listLegacyChannelPayinTypesMemory(chSnap *ChannelSnapshot, amount int64) []model.PayinProductOption {
	if chSnap == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var codes []string
	chSnap.ForEach(func(id int64, ch *configkv.ChannelKV) {
		if ch == nil || !ch.Enabled || ch.FuseEnabled || !ch.SupportsPayin || ch.Weight <= 0 {
			return
		}
		if ch.MinAmount > 0 && ch.MinAmount > amount {
			return
		}
		if ch.MaxAmount > 0 && ch.MaxAmount < amount {
			return
		}
		pt := strings.TrimSpace(ch.DriverKey)
		if pt == "" {
			pt = "mock"
		}
		if _, ok := seen[pt]; ok {
			return
		}
		seen[pt] = struct{}{}
		codes = append(codes, pt)
	})
	sort.Strings(codes)
	out := make([]model.PayinProductOption, 0, len(codes))
	for _, c := range codes {
		out = append(out, model.PayinProductOption{Code: c, Name: c})
	}
	return out
}

// MerchantHasPayinProductCodeMemory checks merchant payin code allowance using snapshots only.
func MerchantHasPayinProductCodeMemory(
	merchantID, payinProductCode string,
	grantsSnap *MerchantPayinGrantsSnapshot,
	payinSnap *PayinProductSnapshot,
) bool {
	merchantID = strings.TrimSpace(merchantID)
	code := strings.TrimSpace(payinProductCode)
	if merchantID == "" || code == "" {
		return false
	}
	if !MerchantPayWhitelistStrictMemory(grantsSnap, merchantID) {
		return true
	}
	p, ok := payinSnap.GetByCode(code)
	if !ok || p == nil || !p.Enabled {
		return false
	}
	gkv, ok := grantsSnap.Get(merchantID)
	if !ok || gkv == nil {
		return false
	}
	for _, g := range gkv.Grants {
		if g.PayinProductID == p.ID {
			return true
		}
	}
	return false
}

// RoutePayinFromMemory selects channel by payin product code using snapshots only.
func RoutePayinFromMemory(
	payinProductCode string,
	amount int64,
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
) (channelID, payProductID int64, err error) {
	code := strings.TrimSpace(payinProductCode)
	if code == "" {
		return 0, 0, errors.New("payin_type (product code) required")
	}
	if ch, pid, e := routeByPayinProductMemory(code, amount, payinSnap, bindSnap, chSnap); e == nil && ch > 0 {
		return ch, pid, nil
	}
	ch, e := routeLegacyMemory(code, amount, chSnap)
	if e != nil {
		return 0, 0, e
	}
	return ch, 0, nil
}

func routeByPayinProductMemory(
	payinProductCode string,
	amount int64,
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
) (channelID, payProductID int64, err error) {
	p, ok := payinSnap.GetByCode(payinProductCode)
	if !ok || p == nil || !p.Enabled {
		return 0, 0, errors.New("no route")
	}
	blob, ok := bindSnap.Get(p.ID)
	if !ok || blob == nil {
		return 0, 0, errors.New("no route")
	}
	var picks []routePick
	var total int64
	for i := range blob.Bindings {
		b := blob.Bindings[i]
		if !b.Enabled || b.Weight <= 0 {
			continue
		}
		ch, ok := chSnap.Get(b.ChannelID)
		if !ok || !ChannelEligiblePayin(ch, amount) {
			continue
		}
		picks = append(picks, routePick{ChannelID: b.ChannelID, Weight: b.Weight, PayinProductID: p.ID})
		total += b.Weight
	}
	if len(picks) == 0 || total <= 0 {
		return 0, 0, errors.New("no route")
	}
	r := rand.Int64N(total)
	var acc int64
	for _, p := range picks {
		acc += p.Weight
		if r < acc {
			return p.ChannelID, p.PayinProductID, nil
		}
	}
	last := picks[len(picks)-1]
	return last.ChannelID, last.PayinProductID, nil
}

func routeLegacyMemory(payType string, amount int64, chSnap *ChannelSnapshot) (int64, error) {
	if chSnap == nil {
		return 0, errors.New("no available channel")
	}
	type cw struct {
		ID     int64
		Weight int64
	}
	var items []cw
	var total int64
	chSnap.ForEach(func(id int64, ch *configkv.ChannelKV) {
		if ch == nil || !ch.Enabled || ch.FuseEnabled || !ch.SupportsPayin || ch.Weight <= 0 {
			return
		}
		if ch.MinAmount > 0 && ch.MinAmount > amount {
			return
		}
		if ch.MaxAmount > 0 && ch.MaxAmount < amount {
			return
		}
		pt := strings.TrimSpace(ch.DriverKey)
		if pt != "" && pt != payType {
			return
		}
		items = append(items, cw{ID: ch.ID, Weight: ch.Weight})
		total += ch.Weight
	})
	if len(items) == 0 || total <= 0 {
		return 0, errors.New("no available channel")
	}
	r := rand.Int64N(total)
	var acc int64
	for _, it := range items {
		acc += it.Weight
		if r < acc {
			return it.ID, nil
		}
	}
	return items[len(items)-1].ID, nil
}

// ResolveLockedChannelForMerchantMemory resolves payin product when API locks channel_id.
func ResolveLockedChannelForMerchantMemory(
	merchantID string,
	channelID int64,
	amount int64,
	grantsSnap *MerchantPayinGrantsSnapshot,
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
) (payProductID int64, payinProductCode string, err error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" || channelID <= 0 {
		return 0, "", errors.New("merchant_id and channel_id required")
	}
	strict := MerchantPayWhitelistStrictMemory(grantsSnap, merchantID)
	ch, ok := chSnap.Get(channelID)
	if !ok || !ChannelEligiblePayin(ch, amount) {
		return 0, "", errors.New("channel not allowed for merchant or amount out of range")
	}

	allowProduct := func(pid int64) bool {
		if !strict {
			return true
		}
		gkv, ok := grantsSnap.Get(merchantID)
		if !ok || gkv == nil {
			return false
		}
		for _, g := range gkv.Grants {
			if g.PayinProductID == pid {
				return true
			}
		}
		return false
	}

	var bestW int64 = -1
	var bestPID int64
	var bestCode string
	payinSnap.ForEach(func(pid int64, p *configkv.PayinProductKV) {
		if p == nil || !p.Enabled {
			return
		}
		if !allowProduct(pid) {
			return
		}
		blob, ok := bindSnap.Get(pid)
		if !ok || blob == nil {
			return
		}
		for i := range blob.Bindings {
			b := blob.Bindings[i]
			if b.ChannelID != channelID || !b.Enabled || b.Weight <= 0 {
				continue
			}
			if b.Weight > bestW {
				bestW = b.Weight
				bestPID = pid
				bestCode = strings.TrimSpace(p.Code)
			}
		}
	})
	if bestPID <= 0 || bestCode == "" {
		return 0, "", errors.New("channel not allowed for merchant or amount out of range")
	}
	return bestPID, bestCode, nil
}

// ListTerminalPayinProductsMemory mirrors PayinProductsStore.ListTerminalPayinProducts using snapshots only.
func ListTerminalPayinProductsMemory(
	merchantID string,
	amount int64,
	grantsSnap *MerchantPayinGrantsSnapshot,
	payinSnap *PayinProductSnapshot,
	bindSnap *PayinProductBindingsSnapshot,
	chSnap *ChannelSnapshot,
) []model.PayinProductOption {
	if MerchantPayWhitelistStrictMemory(grantsSnap, merchantID) {
		return ListAvailableForMerchantAndAmountMemory(merchantID, grantsSnap, payinSnap, bindSnap, chSnap, amount)
	}
	return ListAvailableForAmountMemory(payinSnap, bindSnap, chSnap, amount)
}

// GetPayinProductDisplayNameMemory resolves display name from payin product snapshot only (no DB).
func GetPayinProductDisplayNameMemory(code string, payinSnap *PayinProductSnapshot) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	if payinSnap == nil {
		return code
	}
	p, ok := payinSnap.GetByCode(code)
	if !ok || p == nil || !p.Enabled {
		return code
	}
	merged := strings.TrimSpace(p.ProductConfig)
	if dn := DisplayNameFromProductJSON(merged); dn != "" {
		return dn
	}
	if n := strings.TrimSpace(p.Name); n != "" {
		return n
	}
	return code
}
