package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/trade/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toChannelRow(c *store.Channel) *channelpb.ChannelRow {
	if c == nil {
		return nil
	}
	return &channelpb.ChannelRow{
		Id:                     c.ID,
		Name:                   c.Name,
		PayType:                c.PayType,
		GatewayUrl:             c.GatewayUrl,
		UpstreamMerchantNo:     c.UpstreamMerchantNo,
		RsaPrivateKey:          c.RsaPrivateKey,
		SignSecret:             c.SignSecret,
		Weight:                 c.Weight,
		MinAmount:              c.MinAmount,
		MaxAmount:              c.MaxAmount,
		Enabled:                c.Enabled,
		FuseEnabled:            c.FuseEnabled,
		SupportsCollect:        c.SupportsCollect,
		SupportsPayout:         c.SupportsPayout,
		UpstreamCollectRateBps: c.UpstreamCollectRateBps,
		UpstreamPayoutRateBps:  c.UpstreamPayoutRateBps,
		UpstreamPayoutFeeMode:  c.UpstreamPayoutFeeMode,
		UpstreamPayoutFixedFee: c.UpstreamPayoutFixedFee,
	}
}

func fromUpsertReq(req *channelpb.UpsertChannelReq) *store.Channel {
	feeMode := req.GetUpstreamPayoutFeeMode()
	if feeMode < 1 || feeMode > 3 {
		feeMode = 1
	}
	fixedFee := req.GetUpstreamPayoutFixedFee()
	if fixedFee < 0 {
		fixedFee = 0
	}
	return &store.Channel{
		Name:                   req.GetName(),
		PayType:                req.GetPayType(),
		GatewayUrl:             req.GetGatewayUrl(),
		UpstreamMerchantNo:     req.GetUpstreamMerchantNo(),
		RsaPrivateKey:          req.GetRsaPrivateKey(),
		SignSecret:             req.GetSignSecret(),
		Weight:                 req.GetWeight(),
		MinAmount:              req.GetMinAmount(),
		MaxAmount:              req.GetMaxAmount(),
		Enabled:                req.GetEnabled(),
		FuseEnabled:            req.GetFuseEnabled(),
		SupportsCollect:        req.GetSupportsCollect(),
		SupportsPayout:         req.GetSupportsPayout(),
		UpstreamCollectRateBps: req.GetUpstreamCollectRateBps(),
		UpstreamPayoutRateBps:  req.GetUpstreamPayoutRateBps(),
		UpstreamPayoutFeeMode:  feeMode,
		UpstreamPayoutFixedFee: fixedFee,
	}
}

func payBindingToProto(b *store.PayProductBindingAdmin) *channelpb.AdminPayProductBindingRow {
	return &channelpb.AdminPayProductBindingRow{
		Id:           b.ID,
		PayProductId: b.PayProductID,
		ChannelId:    b.ChannelID,
		ChannelName:  b.ChannelName,
		Weight:       b.Weight,
		Enabled:      b.Enabled,
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
	ch := fromUpsertReq(req)
	id, err := s.svcCtx.Channels.AdminCreate(ctx, ch)
	if err != nil {
		return nil, err
	}
	created, err := s.svcCtx.Channels.AdminGetByID(ctx, id)
	if err != nil {
		return nil, err
	}
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
	ch := fromUpsertReq(req)
	if err := s.svcCtx.Channels.AdminUpdate(ctx, req.GetId(), ch); err != nil {
		return nil, err
	}
	updated, err := s.svcCtx.Channels.AdminGetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &channelpb.UpsertChannelResp{Channel: toChannelRow(updated)}, nil
}

func (s *ChannelServer) GetRoutingSummary(ctx context.Context, _ *channelpb.GetRoutingSummaryReq) (*channelpb.GetRoutingSummaryResp, error) {
	rs, err := s.svcCtx.RoutingSummary.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &channelpb.GetRoutingSummaryResp{
		AlgorithmKey:                  "weighted_random_within_product",
		AlgorithmLabel:                "支付产品内加权随机（同产品多上游按权重分流）",
		EnabledPayProducts:            rs.EnabledPayProducts,
		EnabledPayoutProducts:         rs.EnabledPayoutProducts,
		EnabledChannels:               rs.EnabledChannels,
		ActiveBindings:                rs.ActiveBindings,
		ActivePayoutBindings:          rs.ActivePayoutBindings,
		MerchantsWithCollectWhitelist: rs.MerchantsWithCollectWhitelist,
		MerchantsWithPayoutWhitelist:  rs.MerchantsWithPayoutWhitelist,
		FusedChannels:                 rs.FusedChannels,
	}, nil
}

func (s *ChannelServer) ListTerminalPayProducts(ctx context.Context, req *channelpb.ListTerminalPayProductsReq) (*channelpb.ListTerminalPayProductsResp, error) {
	opts, err := s.svcCtx.PayProducts.ListTerminalPayProducts(ctx, req.GetMerchantId(), req.GetAmount())
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.PayProductOption, 0, len(opts))
	for _, o := range opts {
		out = append(out, &channelpb.PayProductOption{Code: o.Code, Name: o.Name})
	}
	return &channelpb.ListTerminalPayProductsResp{Products: out}, nil
}

func (s *ChannelServer) MerchantHasPayProductCode(ctx context.Context, req *channelpb.MerchantHasPayProductCodeReq) (*channelpb.MerchantHasPayProductCodeResp, error) {
	ok, err := s.svcCtx.PayProducts.MerchantHasPayProductCode(ctx, req.GetMerchantId(), req.GetPayProductCode())
	if err != nil {
		return nil, err
	}
	return &channelpb.MerchantHasPayProductCodeResp{Ok: ok}, nil
}

func (s *ChannelServer) ResolveLockedChannelForMerchant(ctx context.Context, req *channelpb.ResolveLockedChannelForMerchantReq) (*channelpb.ResolveLockedChannelForMerchantResp, error) {
	ppid, code, err := s.svcCtx.PayProducts.ResolveLockedChannelForMerchant(ctx, req.GetMerchantId(), req.GetChannelId(), req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &channelpb.ResolveLockedChannelForMerchantResp{PayProductId: ppid, PayProductCode: code}, nil
}

func (s *ChannelServer) GetPayProductDisplayName(ctx context.Context, req *channelpb.GetPayProductDisplayNameReq) (*channelpb.GetPayProductDisplayNameResp, error) {
	name, err := s.svcCtx.PayProducts.GetPayProductDisplayName(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &channelpb.GetPayProductDisplayNameResp{Name: name}, nil
}

func (s *ChannelServer) AdminListPayProducts(ctx context.Context, _ *channelpb.AdminListPayProductsReq) (*channelpb.AdminListPayProductsResp, error) {
	rows, err := s.svcCtx.PayProducts.AdminListAllPayProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayProductRow, 0, len(rows))
	for _, p := range rows {
		out = append(out, &channelpb.AdminPayProductRow{
			Id:        p.ID,
			Code:      p.Code,
			Name:      p.Name,
			SortOrder: p.SortOrder,
			Enabled:   p.Enabled,
		})
	}
	return &channelpb.AdminListPayProductsResp{Products: out}, nil
}

func (s *ChannelServer) AdminCreatePayProduct(ctx context.Context, req *channelpb.AdminCreatePayProductReq) (*channelpb.AdminUpsertPayProductResp, error) {
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	id, err := s.svcCtx.PayProducts.AdminCreatePayProduct(ctx, code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayProducts.AdminGetPayProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayProductResp{Product: &channelpb.AdminPayProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminUpdatePayProduct(ctx context.Context, req *channelpb.AdminUpdatePayProductReq) (*channelpb.AdminUpsertPayProductResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	err := s.svcCtx.PayProducts.AdminUpdatePayProduct(ctx, req.GetId(), code, name, req.GetSortOrder(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := s.svcCtx.PayProducts.AdminGetPayProduct(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	return &channelpb.AdminUpsertPayProductResp{Product: &channelpb.AdminPayProductRow{
		Id: p.ID, Code: p.Code, Name: p.Name, SortOrder: p.SortOrder, Enabled: p.Enabled,
	}}, nil
}

func (s *ChannelServer) AdminListPayProductBindings(ctx context.Context, req *channelpb.AdminListPayProductBindingsReq) (*channelpb.AdminListPayProductBindingsResp, error) {
	if req.GetPayProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if _, err := s.svcCtx.PayProducts.AdminGetPayProduct(ctx, req.GetPayProductId()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	rows, err := s.svcCtx.PayProducts.AdminListBindings(ctx, req.GetPayProductId())
	if err != nil {
		return nil, err
	}
	out := make([]*channelpb.AdminPayProductBindingRow, 0, len(rows))
	for _, b := range rows {
		out = append(out, payBindingToProto(&b))
	}
	return &channelpb.AdminListPayProductBindingsResp{Bindings: out}, nil
}

func (s *ChannelServer) AdminUpsertPayProductBinding(ctx context.Context, req *channelpb.AdminUpsertPayProductBindingReq) (*channelpb.AdminUpsertPayProductBindingResp, error) {
	if req.GetPayProductId() <= 0 || req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "pay_product_id and channel_id required")
	}
	if req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	if _, err := s.svcCtx.PayProducts.AdminGetPayProduct(ctx, req.GetPayProductId()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	ok, err := s.svcCtx.PayProducts.AdminChannelExists(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "channel not found")
	}
	sup, err := s.svcCtx.PayProducts.AdminChannelSupportsCollect(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	if !sup {
		return nil, status.Error(codes.FailedPrecondition, "channel does not support collect")
	}
	bid, err := s.svcCtx.PayProducts.AdminUpsertBinding(ctx, req.GetPayProductId(), req.GetChannelId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		return nil, err
	}
	b, err := s.svcCtx.PayProducts.AdminGetBindingByID(ctx, bid)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminUpsertPayProductBindingResp{Binding: payBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminUpdatePayProductBinding(ctx context.Context, req *channelpb.AdminUpdatePayProductBindingReq) (*channelpb.AdminUpdatePayProductBindingResp, error) {
	if req.GetId() <= 0 || req.GetWeight() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id and positive weight required")
	}
	err := s.svcCtx.PayProducts.AdminUpdateBinding(ctx, req.GetId(), req.GetWeight(), req.GetEnabled())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	b, err := s.svcCtx.PayProducts.AdminGetBindingByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminUpdatePayProductBindingResp{Binding: payBindingToProto(b)}, nil
}

func (s *ChannelServer) AdminDeletePayProductBinding(ctx context.Context, req *channelpb.AdminDeletePayProductBindingReq) (*channelpb.AdminDeletePayProductBindingResp, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	err := s.svcCtx.PayProducts.AdminDeleteBinding(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminDeletePayProductBindingResp{Ok: true}, nil
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "payout product not found")
		}
		return nil, err
	}
	chOk, err := s.svcCtx.PayProducts.AdminChannelExists(ctx, req.GetChannelId())
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &channelpb.AdminDeletePayoutProductBindingResp{Ok: true}, nil
}
