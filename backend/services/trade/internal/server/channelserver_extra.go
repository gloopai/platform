package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gloopai/pay/common/consulx"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/trade/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// channelConfigJSONForAPI 优先返回 DB 中的 channel_config 列；若为空则把旧版分列字段合成 JSON，便于管理台展示与迁移。
func channelConfigJSONForAPI(c *store.Channel) string {
	if c == nil {
		return ""
	}
	if strings.TrimSpace(c.ChannelConfig) != "" {
		return c.ChannelConfig
	}
	leg := map[string]string{
		"gateway_url":         c.GatewayUrl,
		"channel_merchant_no": c.ChannelMerchantNo,
		"sign_secret":         c.SignSecret,
		"rsa_private_key":     c.RsaPrivateKey,
	}
	b, err := json.Marshal(leg)
	if err != nil {
		return ""
	}
	return string(b)
}

func toChannelRow(c *store.Channel) *channelpb.ChannelRow {
	if c == nil {
		return nil
	}
	return &channelpb.ChannelRow{
		Id:                    c.ID,
		Name:                  c.Name,
		PayinType:             c.PayinType,
		GatewayUrl:            c.GatewayUrl,
		ChannelMerchantNo:     c.ChannelMerchantNo,
		RsaPrivateKey:         c.RsaPrivateKey,
		SignSecret:            c.SignSecret,
		ChannelConfig:         channelConfigJSONForAPI(c),
		Weight:                c.Weight,
		MinAmount:             c.MinAmount,
		MaxAmount:             c.MaxAmount,
		Enabled:               c.Enabled,
		FuseEnabled:           c.FuseEnabled,
		SupportsPayin:         c.SupportsPayin,
		SupportsPayout:        c.SupportsPayout,
		ChannelPayinRateBps:   c.ChannelPayinRateBps,
		ChannelPayoutRateBps:  c.ChannelPayoutRateBps,
		ChannelPayoutFeeMode:  c.ChannelPayoutFeeMode,
		ChannelPayoutFixedFee: c.ChannelPayoutFixedFee,
	}
}

func syncChannelConfigKV(ctx context.Context, store *consulx.ConfigStore, channelID int64, configJSON string) {
	if store == nil || channelID <= 0 {
		return
	}
	key := consulx.ChannelConfigKVKey(channelID)
	if key == "" {
		return
	}
	configJSON = strings.TrimSpace(configJSON)
	if configJSON == "" {
		_ = store.Delete(ctx, key)
		return
	}
	_ = store.PutBytes(ctx, key, []byte(configJSON))
}

func validateChannelConfigJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return status.Error(codes.InvalidArgument, "channel_config must be valid JSON")
	}
	return nil
}

func fromUpsertReq(req *channelpb.UpsertChannelReq) *store.Channel {
	feeMode := req.GetChannelPayoutFeeMode()
	if feeMode < 1 || feeMode > 3 {
		feeMode = 1
	}
	fixedFee := req.GetChannelPayoutFixedFee()
	if fixedFee < 0 {
		fixedFee = 0
	}
	return &store.Channel{
		Name:                  req.GetName(),
		PayinType:             req.GetPayinType(),
		GatewayUrl:            "",
		ChannelMerchantNo:     "",
		RsaPrivateKey:         "",
		SignSecret:            "",
		ChannelConfig:         req.GetChannelConfig(),
		Weight:                req.GetWeight(),
		MinAmount:             req.GetMinAmount(),
		MaxAmount:             req.GetMaxAmount(),
		Enabled:               req.GetEnabled(),
		FuseEnabled:           req.GetFuseEnabled(),
		SupportsPayin:         req.GetSupportsPayin(),
		SupportsPayout:        req.GetSupportsPayout(),
		ChannelPayinRateBps:   req.GetChannelPayinRateBps(),
		ChannelPayoutRateBps:  req.GetChannelPayoutRateBps(),
		ChannelPayoutFeeMode:  feeMode,
		ChannelPayoutFixedFee: fixedFee,
	}
}

func payBindingToProto(b *store.PayinProductBindingAdmin) *channelpb.AdminPayinProductBindingRow {
	return &channelpb.AdminPayinProductBindingRow{
		Id:             b.ID,
		PayinProductId: b.PayinProductID,
		ChannelId:      b.ChannelID,
		ChannelName:    b.ChannelName,
		Weight:         b.Weight,
		Enabled:        b.Enabled,
	}
}

func payoutBindingToProto(b *store.PayoutProductBindingAdmin) *channelpb.AdminPayoutProductBindingRow {
	return &channelpb.AdminPayoutProductBindingRow{
		Id:              b.ID,
		PayoutProductId: b.PayoutProductID,
		ChannelId:       b.ChannelID,
		ChannelName:     b.ChannelName,
		Weight:          b.Weight,
		Enabled:         b.Enabled,
	}
}

func (s *ChannelServer) GetChannel(ctx context.Context, req *channelpb.GetChannelReq) (*channelpb.GetChannelResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	ch, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetChannelId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, err
	}
	// 通道配置以库表为准；Consul 内存缓存仅用于 OpenAPI 下单热路径（PrepareTerminalPay / 网关 GetSignSecret），不在此 RPC 覆盖。
	return &channelpb.GetChannelResp{Channel: toChannelRow(ch)}, nil
}

func (s *ChannelServer) ListChannels(ctx context.Context, _ *channelpb.ListChannelsReq) (*channelpb.ListChannelsResp, error) {
	items, err := s.svcCtx.Channels.AdminList(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.ChannelRow, 0, len(items))
	for i := range items {
		out = append(out, toChannelRow(&items[i]))
	}
	return &channelpb.ListChannelsResp{Channels: out}, nil
}

func (s *ChannelServer) CreateChannel(ctx context.Context, req *channelpb.UpsertChannelReq) (*channelpb.UpsertChannelResp, error) {
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	if req.GetWeight() < 0 || req.GetWeight() > 100 {
		return nil, status.Error(codes.InvalidArgument, "weight must be 0-100")
	}
	if req.GetMinAmount() < 0 || req.GetMaxAmount() < 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be >= 0")
	}
	if req.GetMaxAmount() > 0 && req.GetMinAmount() > req.GetMaxAmount() {
		return nil, status.Error(codes.InvalidArgument, "min_amount must be <= max_amount")
	}
	if err := validateChannelConfigJSON(req.GetChannelConfig()); err != nil {
		return nil, err
	}
	ch := fromUpsertReq(req)
	id, err := s.svcCtx.Channels.AdminCreate(ctx, ch)
	if err != nil {
		return nil, err
	}
	created, err := s.svcCtx.Channels.AdminGetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	syncChannelConfigKV(ctx, s.svcCtx.RuntimeConfig, id, created.ChannelConfig)
	return &channelpb.UpsertChannelResp{Channel: toChannelRow(created)}, nil
}

func (s *ChannelServer) UpdateChannel(ctx context.Context, req *channelpb.UpsertChannelReq) (*channelpb.UpsertChannelResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	if req.GetWeight() < 0 || req.GetWeight() > 100 {
		return nil, status.Error(codes.InvalidArgument, "weight must be 0-100")
	}
	if req.GetMinAmount() < 0 || req.GetMaxAmount() < 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be >= 0")
	}
	if req.GetMaxAmount() > 0 && req.GetMinAmount() > req.GetMaxAmount() {
		return nil, status.Error(codes.InvalidArgument, "min_amount must be <= max_amount")
	}
	if err := validateChannelConfigJSON(req.GetChannelConfig()); err != nil {
		return nil, err
	}
	ch := fromUpsertReq(req)
	if err := s.svcCtx.Channels.AdminUpdate(ctx, req.GetId(), ch); err != nil {
		return nil, err
	}
	updated, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	syncChannelConfigKV(ctx, s.svcCtx.RuntimeConfig, req.GetId(), updated.ChannelConfig)
	return &channelpb.UpsertChannelResp{Channel: toChannelRow(updated)}, nil
}

func (s *ChannelServer) GetRoutingSummary(ctx context.Context, _ *channelpb.GetRoutingSummaryReq) (*channelpb.GetRoutingSummaryResp, error) {
	rs, err := s.svcCtx.RoutingSummary.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &channelpb.GetRoutingSummaryResp{
		AlgorithmKey:                 "weighted_random_within_product",
		AlgorithmLabel:               "支付产品内加权随机（同产品多通道按权重分流）",
		EnabledPayinProducts:         rs.EnabledPayinProducts,
		EnabledPayoutProducts:        rs.EnabledPayoutProducts,
		EnabledChannels:              rs.EnabledChannels,
		ActiveBindings:               rs.ActiveBindings,
		ActivePayoutBindings:         rs.ActivePayoutBindings,
		MerchantsWithPayinWhitelist:  rs.MerchantsWithPayinWhitelist,
		MerchantsWithPayoutWhitelist: rs.MerchantsWithPayoutWhitelist,
		FusedChannels:                rs.FusedChannels,
	}, nil
}

func (s *ChannelServer) ListTerminalPayinProducts(ctx context.Context, req *channelpb.ListTerminalPayinProductsReq) (*channelpb.ListTerminalPayinProductsResp, error) {
	opts, err := s.svcCtx.PayinProducts.ListTerminalPayinProducts(ctx, req.GetMerchantId(), req.GetAmount())
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.PayinProductOption, 0, len(opts))
	for _, o := range opts {
		out = append(out, &channelpb.PayinProductOption{Code: o.Code, Name: o.Name})
	}
	return &channelpb.ListTerminalPayinProductsResp{Products: out}, nil
}

func (s *ChannelServer) MerchantHasPayinProductCode(ctx context.Context, req *channelpb.MerchantHasPayinProductCodeReq) (*channelpb.MerchantHasPayinProductCodeResp, error) {
	ok, err := s.svcCtx.PayinProducts.MerchantHasPayinProductCode(ctx, req.GetMerchantId(), req.GetPayinProductCode())
	if err != nil {
		return nil, err
	}
	return &channelpb.MerchantHasPayinProductCodeResp{Ok: ok}, nil
}

func (s *ChannelServer) ResolveLockedChannelForMerchant(ctx context.Context, req *channelpb.ResolveLockedChannelForMerchantReq) (*channelpb.ResolveLockedChannelForMerchantResp, error) {
	ppid, code, err := s.svcCtx.PayinProducts.ResolveLockedChannelForMerchant(ctx, req.GetMerchantId(), req.GetChannelId(), req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &channelpb.ResolveLockedChannelForMerchantResp{PayinProductId: ppid, PayinProductCode: code}, nil
}

func (s *ChannelServer) GetPayinProductDisplayName(ctx context.Context, req *channelpb.GetPayinProductDisplayNameReq) (*channelpb.GetPayinProductDisplayNameResp, error) {
	name, err := s.svcCtx.PayinProducts.GetPayinProductDisplayName(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &channelpb.GetPayinProductDisplayNameResp{Name: name}, nil
}

func (s *ChannelServer) AdminListPayinProducts(ctx context.Context, _ *channelpb.AdminListPayinProductsReq) (*channelpb.AdminListPayinProductsResp, error) {
	rows, err := s.svcCtx.PayinProducts.AdminListAllPayinProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayinProductRow, 0, len(rows))
	for _, p := range rows {
		out = append(out, &channelpb.AdminPayinProductRow{
			Id:        p.ID,
			Code:      p.Code,
			Name:      p.Name,
			SortOrder: p.SortOrder,
			Enabled:   p.Enabled,
		})
	}
	return &channelpb.AdminListPayinProductsResp{Products: out}, nil
}

func (s *ChannelServer) AdminCreatePayinProduct(ctx context.Context, req *channelpb.AdminCreatePayinProductReq) (*channelpb.AdminUpsertPayinProductResp, error) {
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	id, err := s.svcCtx.PayinProducts.AdminCreatePayinProduct(ctx, code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayinProducts.AdminGetPayinProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayinProductResp{Product: &channelpb.AdminPayinProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminUpdatePayinProduct(ctx context.Context, req *channelpb.AdminUpdatePayinProductReq) (*channelpb.AdminUpsertPayinProductResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	err := s.svcCtx.PayinProducts.AdminUpdatePayinProduct(ctx, req.GetId(), code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayinProducts.AdminGetPayinProduct(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	return &channelpb.AdminUpsertPayinProductResp{Product: &channelpb.AdminPayinProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminListPayinProductBindings(ctx context.Context, req *channelpb.AdminListPayinProductBindingsReq) (*channelpb.AdminListPayinProductBindingsResp, error) {
	if req.GetPayinProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if _, err := s.svcCtx.PayinProducts.AdminGetPayinProduct(ctx, req.GetPayinProductId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	rows, err := s.svcCtx.PayinProducts.AdminListBindings(ctx, req.GetPayinProductId())
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayinProductBindingRow, 0, len(rows))
	for _, b := range rows {
		out = append(out, payBindingToProto(&b))
	}
	return &channelpb.AdminListPayinProductBindingsResp{Bindings: out}, nil
}

func (s *ChannelServer) AdminUpsertPayinProductBinding(ctx context.Context, req *channelpb.AdminUpsertPayinProductBindingReq) (*channelpb.AdminUpsertPayinProductBindingResp, error) {
	if req.GetPayinProductId() <= 0 || req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "payin_product_id and channel_id required")
	}
	if req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	if _, err := s.svcCtx.PayinProducts.AdminGetPayinProduct(ctx, req.GetPayinProductId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	ok, err := s.svcCtx.PayinProducts.AdminChannelExists(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "channel not found")
	}
	sup, err := s.svcCtx.PayinProducts.AdminChannelSupportsPayin(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !sup {
		return nil, status.Error(codes.FailedPrecondition, "channel does not support payin")
	}
	bid, err := s.svcCtx.PayinProducts.AdminUpsertBinding(ctx, req.GetPayinProductId(), req.GetChannelId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		return nil, err
	}
	b, err := s.svcCtx.PayinProducts.AdminGetBindingByID(ctx, bid)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayinProductBindingResp{Binding: payBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminUpdatePayinProductBinding(ctx context.Context, req *channelpb.AdminUpdatePayinProductBindingReq) (*channelpb.AdminUpdatePayinProductBindingResp, error) {
	if req.GetId() <= 0 || req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id and positive weight required")
	}
	err := s.svcCtx.PayinProducts.AdminUpdateBinding(ctx, req.GetId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	b, err := s.svcCtx.PayinProducts.AdminGetBindingByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminUpdatePayinProductBindingResp{Binding: payBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminDeletePayinProductBinding(ctx context.Context, req *channelpb.AdminDeletePayinProductBindingReq) (*channelpb.AdminDeletePayinProductBindingResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	err := s.svcCtx.PayinProducts.AdminDeleteBinding(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminDeletePayinProductBindingResp{Ok: true}, nil
}

func (s *ChannelServer) AdminListPayoutProducts(ctx context.Context, _ *channelpb.AdminListPayoutProductsReq) (*channelpb.AdminListPayoutProductsResp, error) {
	rows, err := s.svcCtx.PayoutProducts.AdminListAllPayoutProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayoutProductRow, 0, len(rows))
	for _, p := range rows {
		out = append(out, &channelpb.AdminPayoutProductRow{
			Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
		})
	}
	return &channelpb.AdminListPayoutProductsResp{Products: out}, nil
}

func (s *ChannelServer) AdminCreatePayoutProduct(ctx context.Context, req *channelpb.AdminCreatePayoutProductReq) (*channelpb.AdminUpsertPayoutProductResp, error) {
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	id, err := s.svcCtx.PayoutProducts.AdminCreatePayoutProduct(ctx, code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayoutProducts.AdminGetPayoutProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayoutProductResp{Product: &channelpb.AdminPayoutProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminUpdatePayoutProduct(ctx context.Context, req *channelpb.AdminUpdatePayoutProductReq) (*channelpb.AdminUpsertPayoutProductResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	err := s.svcCtx.PayoutProducts.AdminUpdatePayoutProduct(ctx, req.GetId(), code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payout product not found")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayoutProducts.AdminGetPayoutProduct(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayoutProductResp{Product: &channelpb.AdminPayoutProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminListPayoutProductBindings(ctx context.Context, req *channelpb.AdminListPayoutProductBindingsReq) (*channelpb.AdminListPayoutProductBindingsResp, error) {
	if req.GetPayoutProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if _, err := s.svcCtx.PayoutProducts.AdminGetPayoutProduct(ctx, req.GetPayoutProductId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payout product not found")
		}
		return nil, err
	}
	rows, err := s.svcCtx.PayoutProducts.AdminListPayoutBindings(ctx, req.GetPayoutProductId())
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayoutProductBindingRow, 0, len(rows))
	for _, b := range rows {
		out = append(out, payoutBindingToProto(&b))
	}
	return &channelpb.AdminListPayoutProductBindingsResp{Bindings: out}, nil
}

func (s *ChannelServer) AdminUpsertPayoutProductBinding(ctx context.Context, req *channelpb.AdminUpsertPayoutProductBindingReq) (*channelpb.AdminUpsertPayoutProductBindingResp, error) {
	if req.GetPayoutProductId() <= 0 || req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "payout_product_id and channel_id required")
	}
	if req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	if _, err := s.svcCtx.PayoutProducts.AdminGetPayoutProduct(ctx, req.GetPayoutProductId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payout product not found")
		}
		return nil, err
	}
	chOk, err := s.svcCtx.PayinProducts.AdminChannelExists(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !chOk {
		return nil, status.Error(codes.NotFound, "channel not found")
	}
	ok, err := s.svcCtx.PayoutProducts.AdminChannelSupportsPayout(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Error(codes.FailedPrecondition, "channel does not support payout")
	}
	bid, err := s.svcCtx.PayoutProducts.AdminUpsertPayoutBinding(ctx, req.GetPayoutProductId(), req.GetChannelId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		return nil, err
	}
	b, err := s.svcCtx.PayoutProducts.AdminGetPayoutBindingByID(ctx, bid)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayoutProductBindingResp{Binding: payoutBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminUpdatePayoutProductBinding(ctx context.Context, req *channelpb.AdminUpdatePayoutProductBindingReq) (*channelpb.AdminUpdatePayoutProductBindingResp, error) {
	if req.GetId() <= 0 || req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id and positive weight required")
	}
	err := s.svcCtx.PayoutProducts.AdminUpdatePayoutBinding(ctx, req.GetId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	b, err := s.svcCtx.PayoutProducts.AdminGetPayoutBindingByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpdatePayoutProductBindingResp{Binding: payoutBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminDeletePayoutProductBinding(ctx context.Context, req *channelpb.AdminDeletePayoutProductBindingReq) (*channelpb.AdminDeletePayoutProductBindingResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	err := s.svcCtx.PayoutProducts.AdminDeletePayoutBinding(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminDeletePayoutProductBindingResp{Ok: true}, nil
}
