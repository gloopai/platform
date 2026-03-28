package server

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/gloopai/pay/common/channelconfig"
	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/model"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/core/internal/configsync"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/logic"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ChannelServer struct {
	svcCtx *svc.ServiceContext
	channelpb.UnimplementedChannelServer
}

func NewChannelServer(svcCtx *svc.ServiceContext) *ChannelServer {
	return &ChannelServer{
		svcCtx: svcCtx,
	}
}

func (s *ChannelServer) Route(ctx context.Context, in *channelpb.RouteReq) (*channelpb.RouteResp, error) {
	l := logic.NewRouteLogic(ctx, s.svcCtx)
	return l.Route(in)
}

func (s *ChannelServer) PreparePayinOrder(ctx context.Context, in *channelpb.PreparePayinOrderReq) (*channelpb.PreparePayinOrderResp, error) {
	return logic.PreparePayinOrder(ctx, s.svcCtx, in)
}

func (s *ChannelServer) GetSignSecret(ctx context.Context, in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	l := logic.NewGetSignSecretLogic(ctx, s.svcCtx)
	return l.GetSignSecret(in)
}

func toChannelRow(c *model.Channel) *channelpb.ChannelRow {
	if c == nil {
		return nil
	}
	return &channelpb.ChannelRow{
		Id:                c.ID,
		Name:              c.Name,
		PayinType:         c.PayinType,
		GatewayUrl:        c.GatewayUrl,
		ChannelMerchantNo: c.ChannelMerchantNo,
		RsaPrivateKey:     c.RsaPrivateKey,
		SignSecret:        c.SignSecret,
		ChannelConfig: channelconfig.ChannelConfigJSONForAPI(c.ChannelConfig, channelconfig.LegacyChannelFields{
			GatewayURL:        c.GatewayUrl,
			ChannelMerchantNo: c.ChannelMerchantNo,
			SignSecret:        c.SignSecret,
			RSAPrivateKey:     c.RsaPrivateKey,
		}),
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

func syncChannelKV(ctx context.Context, cfg *consulx.ConfigStore, ch *model.Channel) {
	if cfg == nil || ch == nil || ch.ID <= 0 {
		return
	}
	key := configkv.ChannelSnapshotKVKey(ch.ID)
	if key == "" {
		return
	}
	kv := store.ChannelToKV(ch)
	b, err := json.Marshal(kv)
	if err != nil {
		return
	}
	_ = cfg.PutBytes(ctx, key, b)
}

func validateProductConfigJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return status.Error(codes.InvalidArgument, "product_config must be valid JSON")
	}
	return nil
}

func syncPayinProductKV(ctx context.Context, cfg *consulx.ConfigStore, p *model.PayinProductAdmin) {
	if cfg == nil || p == nil || p.ID <= 0 {
		return
	}
	key := configkv.PayinProductSnapshotKVKey(p.ID)
	if key == "" {
		return
	}
	kv := store.PayinProductAdminToKV(p)
	b, err := json.Marshal(kv)
	if err != nil {
		return
	}
	_ = cfg.PutBytes(ctx, key, b)
}

func syncPayoutProductKV(ctx context.Context, cfg *consulx.ConfigStore, p *model.PayoutProductAdmin) {
	if cfg == nil || p == nil || p.ID <= 0 {
		return
	}
	key := configkv.PayoutProductSnapshotKVKey(p.ID)
	if key == "" {
		return
	}
	kv := store.PayoutProductAdminToKV(p)
	b, err := json.Marshal(kv)
	if err != nil {
		return
	}
	_ = cfg.PutBytes(ctx, key, b)
}

func fromUpsertReq(req *channelpb.UpsertChannelReq) *model.Channel {
	feeMode := req.GetChannelPayoutFeeMode()
	if feeMode < 1 || feeMode > 3 {
		feeMode = 1
	}
	fixedFee := req.GetChannelPayoutFixedFee()
	if fixedFee < 0 {
		fixedFee = 0
	}
	return &model.Channel{
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

func payBindingToProto(b *model.PayinProductBindingAdmin) *channelpb.AdminPayinProductBindingRow {
	return &channelpb.AdminPayinProductBindingRow{
		Id:             b.ID,
		PayinProductId: b.PayinProductID,
		ChannelId:      b.ChannelID,
		ChannelName:    b.ChannelName,
		Weight:         b.Weight,
		Enabled:        b.Enabled,
	}
}

func payoutBindingToProto(b *model.PayoutProductBindingAdmin) *channelpb.AdminPayoutProductBindingRow {
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
	if req.GetAuthoritativeDb() {
		ch, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetChannelId())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Error(codes.NotFound, "channel not found")
			}
			return nil, err
		}
		return &channelpb.GetChannelResp{Channel: toChannelRow(ch)}, nil
	}
	if snap, ok := s.svcCtx.ChannelSnapshot.Get(req.GetChannelId()); ok && snap != nil {
		ch := store.KVToChannel(snap)
		effective := *ch
		effective.ChannelConfig = kvcache.PickChannelConfig(s.svcCtx.ChannelSnapshot, ch.ID, ch.ChannelConfig)
		return &channelpb.GetChannelResp{Channel: toChannelRow(&effective)}, nil
	}
	ch, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetChannelId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, err
	}
	effective := *ch
	effective.ChannelConfig = kvcache.PickChannelConfig(s.svcCtx.ChannelSnapshot, ch.ID, ch.ChannelConfig)
	return &channelpb.GetChannelResp{Channel: toChannelRow(&effective)}, nil
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
	if err := channelconfig.ValidateChannelConfigJSON(req.GetChannelConfig()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	syncChannelKV(ctx, s.svcCtx.RuntimeConfig, created)
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
	if err := channelconfig.ValidateChannelConfigJSON(req.GetChannelConfig()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ch := fromUpsertReq(req)
	if err := s.svcCtx.Channels.AdminUpdate(ctx, req.GetId(), ch); err != nil {
		return nil, err
	}
	updated, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	syncChannelKV(ctx, s.svcCtx.RuntimeConfig, updated)
	s.svcCtx.InvalidateChannelDriverCache(req.GetId())
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
	if s.svcCtx.OpenAPIMemoryReady() {
		opts := kvcache.ListTerminalPayinProductsMemory(
			req.GetMerchantId(),
			req.GetAmount(),
			s.svcCtx.MerchantPayinGrantsSnapshot,
			s.svcCtx.PayinProductSnapshot,
			s.svcCtx.PayinProductBindingsSnapshot,
			s.svcCtx.ChannelSnapshot,
		)
		out := make([]*channelpb.PayinProductOption, 0, len(opts))
		for _, o := range opts {
			out = append(out, &channelpb.PayinProductOption{Code: o.Code, Name: o.Name})
		}
		return &channelpb.ListTerminalPayinProductsResp{Products: out}, nil
	}
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
	if s.svcCtx.OpenAPIMemoryReady() {
		ok := kvcache.MerchantHasPayinProductCodeMemory(
			req.GetMerchantId(),
			req.GetPayinProductCode(),
			s.svcCtx.MerchantPayinGrantsSnapshot,
			s.svcCtx.PayinProductSnapshot,
		)
		return &channelpb.MerchantHasPayinProductCodeResp{Ok: ok}, nil
	}
	ok, err := s.svcCtx.PayinProducts.MerchantHasPayinProductCode(ctx, req.GetMerchantId(), req.GetPayinProductCode())
	if err != nil {
		return nil, err
	}
	return &channelpb.MerchantHasPayinProductCodeResp{Ok: ok}, nil
}

func (s *ChannelServer) ResolveLockedChannelForMerchant(ctx context.Context, req *channelpb.ResolveLockedChannelForMerchantReq) (*channelpb.ResolveLockedChannelForMerchantResp, error) {
	if s.svcCtx.OpenAPIMemoryReady() {
		ppid, code, err := kvcache.ResolveLockedChannelForMerchantMemory(
			req.GetMerchantId(),
			req.GetChannelId(),
			req.GetAmount(),
			s.svcCtx.MerchantPayinGrantsSnapshot,
			s.svcCtx.PayinProductSnapshot,
			s.svcCtx.PayinProductBindingsSnapshot,
			s.svcCtx.ChannelSnapshot,
		)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return &channelpb.ResolveLockedChannelForMerchantResp{PayinProductId: ppid, PayinProductCode: code}, nil
	}
	ppid, code, err := s.svcCtx.PayinProducts.ResolveLockedChannelForMerchant(ctx, req.GetMerchantId(), req.GetChannelId(), req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &channelpb.ResolveLockedChannelForMerchantResp{PayinProductId: ppid, PayinProductCode: code}, nil
}

func (s *ChannelServer) GetPayinProductDisplayName(ctx context.Context, req *channelpb.GetPayinProductDisplayNameReq) (*channelpb.GetPayinProductDisplayNameResp, error) {
	if !req.GetAuthoritativeDb() && s.svcCtx.PayinProductSnapshot != nil && s.svcCtx.RuntimeConfig != nil {
		name := kvcache.GetPayinProductDisplayNameMemory(req.GetCode(), s.svcCtx.PayinProductSnapshot)
		return &channelpb.GetPayinProductDisplayNameResp{Name: name}, nil
	}
	var snap *kvcache.PayinProductSnapshot
	if !req.GetAuthoritativeDb() {
		snap = s.svcCtx.PayinProductSnapshot
	}
	name, err := s.svcCtx.PayinProducts.GetPayinProductDisplayName(ctx, req.GetCode(), snap)
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
			Id:            p.ID,
			Code:          p.Code,
			Name:          p.Name,
			SortOrder:     p.SortOrder,
			Enabled:       p.Enabled,
			ProductConfig: p.ProductConfig,
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
	pc := strings.TrimSpace(req.GetProductConfig())
	if err := validateProductConfigJSON(pc); err != nil {
		return nil, err
	}
	id, err := s.svcCtx.PayinProducts.AdminCreatePayinProduct(ctx, code, name, req.GetSortOrder(), req.GetEnabled(), pc)
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
	syncPayinProductKV(ctx, s.svcCtx.RuntimeConfig, p)
	return &channelpb.AdminUpsertPayinProductResp{Product: &channelpb.AdminPayinProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
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
	pc := strings.TrimSpace(req.GetProductConfig())
	if err := validateProductConfigJSON(pc); err != nil {
		return nil, err
	}
	err := s.svcCtx.PayinProducts.AdminUpdatePayinProduct(ctx, req.GetId(), code, name, req.GetSortOrder(), req.GetEnabled(), pc)
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
	syncPayinProductKV(ctx, s.svcCtx.RuntimeConfig, p)
	return &channelpb.AdminUpsertPayinProductResp{Product: &channelpb.AdminPayinProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
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
	_ = configsync.SyncPayinProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayinProducts, req.GetPayinProductId())
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
	_ = configsync.SyncPayinProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayinProducts, b.PayinProductID)
	return &channelpb.AdminUpdatePayinProductBindingResp{Binding: payBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminDeletePayinProductBinding(ctx context.Context, req *channelpb.AdminDeletePayinProductBindingReq) (*channelpb.AdminDeletePayinProductBindingResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	prev, err := s.svcCtx.PayinProducts.AdminGetBindingByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	pid := prev.PayinProductID
	err = s.svcCtx.PayinProducts.AdminDeleteBinding(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	_ = configsync.SyncPayinProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayinProducts, pid)
	return &channelpb.AdminDeletePayinProductBindingResp{Ok: true}, nil
}

func (s *ChannelServer) AdminListPayoutProducts(ctx context.Context, _ *channelpb.AdminListPayoutProductsReq) (*channelpb.AdminListPayoutProductsResp, error) {
	if s.svcCtx.PayoutProductSnapshot != nil && s.svcCtx.RuntimeConfig != nil {
		type row struct {
			id        int64
			sortOrder int64
			p         *configkv.PayoutProductKV
		}
		var tmp []row
		s.svcCtx.PayoutProductSnapshot.ForEach(func(id int64, p *configkv.PayoutProductKV) {
			if p != nil {
				tmp = append(tmp, row{id: id, sortOrder: p.SortOrder, p: p})
			}
		})
		sort.Slice(tmp, func(i, j int) bool {
			if tmp[i].sortOrder != tmp[j].sortOrder {
				return tmp[i].sortOrder < tmp[j].sortOrder
			}
			return tmp[i].id < tmp[j].id
		})
		out := make([]*channelpb.AdminPayoutProductRow, 0, len(tmp))
		for _, r := range tmp {
			p := r.p
			out = append(out, &channelpb.AdminPayoutProductRow{
				Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
			})
		}
		return &channelpb.AdminListPayoutProductsResp{Products: out}, nil
	}
	rows, err := s.svcCtx.PayoutProducts.AdminListAllPayoutProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayoutProductRow, 0, len(rows))
	for _, p := range rows {
		out = append(out, &channelpb.AdminPayoutProductRow{
			Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
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
	pc := strings.TrimSpace(req.GetProductConfig())
	if err := validateProductConfigJSON(pc); err != nil {
		return nil, err
	}
	id, err := s.svcCtx.PayoutProducts.AdminCreatePayoutProduct(ctx, code, name, req.GetSortOrder(), req.GetEnabled(), pc)
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
	syncPayoutProductKV(ctx, s.svcCtx.RuntimeConfig, p)
	return &channelpb.AdminUpsertPayoutProductResp{Product: &channelpb.AdminPayoutProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
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
	pc := strings.TrimSpace(req.GetProductConfig())
	if err := validateProductConfigJSON(pc); err != nil {
		return nil, err
	}
	err := s.svcCtx.PayoutProducts.AdminUpdatePayoutProduct(ctx, req.GetId(), code, name, req.GetSortOrder(), req.GetEnabled(), pc)
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
	syncPayoutProductKV(ctx, s.svcCtx.RuntimeConfig, p)
	return &channelpb.AdminUpsertPayoutProductResp{Product: &channelpb.AdminPayoutProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled, ProductConfig: p.ProductConfig,
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
	_ = configsync.SyncPayoutProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayoutProducts, req.GetPayoutProductId())
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
	_ = configsync.SyncPayoutProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayoutProducts, b.PayoutProductID)
	return &channelpb.AdminUpdatePayoutProductBindingResp{Binding: payoutBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminDeletePayoutProductBinding(ctx context.Context, req *channelpb.AdminDeletePayoutProductBindingReq) (*channelpb.AdminDeletePayoutProductBindingResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	prev, err := s.svcCtx.PayoutProducts.AdminGetPayoutBindingByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	pid := prev.PayoutProductID
	err = s.svcCtx.PayoutProducts.AdminDeletePayoutBinding(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	_ = configsync.SyncPayoutProductChannelBindings(ctx, s.svcCtx.RuntimeConfig, s.svcCtx.PayoutProducts, pid)
	return &channelpb.AdminDeletePayoutProductBindingResp{Ok: true}, nil
}
