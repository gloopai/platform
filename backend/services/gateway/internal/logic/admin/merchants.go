package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminMerchants 管理后台商户（入驻商户）与商户侧支付产品白名单。
type AdminMerchants struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminMerchants(ctx context.Context, svcCtx *svc.ServiceContext) *AdminMerchants {
	return &AdminMerchants{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func toAdminMerchantInfo(m *merchantpb.MerchantInfo) types.AdminMerchantInfo {
	if m == nil {
		return types.AdminMerchantInfo{}
	}
	cg := make([]types.MerchantCollectGrant, 0, len(m.GetCollectGrants()))
	for _, g := range m.GetCollectGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantCollectGrant{PayProductId: g.GetPayProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		cg = append(cg, row)
	}
	pg := make([]types.MerchantPayoutGrant, 0, len(m.GetPayoutGrants()))
	for _, g := range m.GetPayoutGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayoutGrant{PayoutProductId: g.GetPayoutProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		row.FeeMode = g.GetFeeMode()
		if row.FeeMode <= 0 {
			row.FeeMode = 1
		}
		row.FeeFixedAmount = g.GetFeeFixedAmount()
		pg = append(pg, row)
	}
	return types.AdminMerchantInfo{
		MerchantId:            m.GetMerchantId(),
		ApiSecret:             m.GetApiSecret(),
		Status:                m.GetStatus(),
		DefaultCollectRateBps: m.GetDefaultCollectRateBps(),
		DefaultPayoutRateBps:  m.GetDefaultPayoutRateBps(),
		NotifyUrl:             m.GetNotifyUrl(),
		ReturnUrl:             m.GetReturnUrl(),
		IpWhitelist:           m.GetIpWhitelist(),
		Balance:               m.GetBalance(),
		PayProductIds:         m.GetPayProductIds(),
		PayoutProductIds:      m.GetPayoutProductIds(),
		CollectGrants:         cg,
		PayoutGrants:          pg,
	}
}

func (m *AdminMerchants) AdminListMerchants() (*types.AdminListMerchantsResp, error) {
	r, err := m.svcCtx.MerchantRpc.ListMerchants(m.ctx, &merchantclient.ListMerchantsReq{Limit: 200})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminMerchantInfo, 0, len(r.GetMerchants()))
	for _, row := range r.GetMerchants() {
		out = append(out, toAdminMerchantInfo(row))
	}
	return &types.AdminListMerchantsResp{Merchants: out}, nil
}

func (m *AdminMerchants) AdminCreateMerchant(req *types.AdminCreateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := strings.TrimSpace(req.ApiSecret)
	r, err := m.svcCtx.MerchantRpc.CreateMerchant(m.ctx, &merchantclient.CreateMerchantReq{
		MerchantId:            merchantId,
		ApiSecret:             secret,
		Status:                1,
		DefaultCollectRateBps: req.DefaultCollectRateBps,
		DefaultPayoutRateBps:  req.DefaultPayoutRateBps,
		NotifyUrl:             req.NotifyUrl,
		ReturnUrl:             req.ReturnUrl,
		IpWhitelist:           req.IpWhitelist,
		PayProductIds:         req.PayProductIds,
		PayoutProductIds:      req.PayoutProductIds,
	})
	if err != nil {
		return nil, err
	}
	created := r.GetMerchant()
	return &types.AdminUpsertMerchantResp{
		Merchant: toAdminMerchantInfo(created),
	}, nil
}

func (m *AdminMerchants) AdminUpdateMerchant(req *types.AdminUpdateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := ""
	if req.ResetSecret {
		tok, err := shared.NewToken()
		if err != nil {
			return nil, err
		}
		secret = tok
	}
	r, err := m.svcCtx.MerchantRpc.UpdateMerchant(m.ctx, &merchantclient.UpdateMerchantReq{
		MerchantId:            merchantId,
		ApiSecret:             secret,
		Status:                req.Status,
		DefaultCollectRateBps: req.DefaultCollectRateBps,
		DefaultPayoutRateBps:  req.DefaultPayoutRateBps,
		NotifyUrl:             req.NotifyUrl,
		ReturnUrl:             req.ReturnUrl,
		IpWhitelist:           req.IpWhitelist,
	})
	if err != nil {
		return nil, err
	}
	updated := r.GetMerchant()

	if req.CollectGrants != nil {
		pbGrants := make([]*merchantpb.MerchantCollectGrant, 0, len(req.CollectGrants))
		for _, g := range req.CollectGrants {
			row := &merchantpb.MerchantCollectGrant{PayProductId: g.PayProductId}
			if g.MerchantRateBps != nil {
				v := *g.MerchantRateBps
				row.MerchantRateBps = &v
			}
			pbGrants = append(pbGrants, row)
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayProducts(m.ctx, &merchantpb.ReplaceMerchantPayProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	} else if req.PayProductIds != nil {
		var pbGrants []*merchantpb.MerchantCollectGrant
		for _, id := range req.PayProductIds {
			if id <= 0 {
				continue
			}
			pbGrants = append(pbGrants, &merchantpb.MerchantCollectGrant{PayProductId: id})
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayProducts(m.ctx, &merchantpb.ReplaceMerchantPayProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}

	if req.PayoutGrants != nil {
		pbGrants := make([]*merchantpb.MerchantPayoutGrant, 0, len(req.PayoutGrants))
		for _, g := range req.PayoutGrants {
			row := &merchantpb.MerchantPayoutGrant{
				PayoutProductId: g.PayoutProductId,
				FeeMode:         g.FeeMode,
				FeeFixedAmount:  g.FeeFixedAmount,
			}
			if row.FeeMode <= 0 {
				row.FeeMode = 1
			}
			if g.MerchantRateBps != nil {
				v := *g.MerchantRateBps
				row.MerchantRateBps = &v
			}
			pbGrants = append(pbGrants, row)
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayoutProducts(m.ctx, &merchantpb.ReplaceMerchantPayoutProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant payout products failed")
		}
	} else if req.PayoutProductIds != nil {
		var pbGrants []*merchantpb.MerchantPayoutGrant
		for _, id := range req.PayoutProductIds {
			if id <= 0 {
				continue
			}
			pbGrants = append(pbGrants, &merchantpb.MerchantPayoutGrant{PayoutProductId: id})
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayoutProducts(m.ctx, &merchantpb.ReplaceMerchantPayoutProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant payout products failed")
		}
	}

	gr, err := m.svcCtx.MerchantRpc.ListMerchantPayProductIds(m.ctx, &merchantpb.ListMerchantPayProductIdsReq{MerchantId: merchantId})
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}
	pr, err := m.svcCtx.MerchantRpc.ListMerchantPayoutProductIds(m.ctx, &merchantpb.ListMerchantPayoutProductIdsReq{MerchantId: merchantId})
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant payout products failed")
	}
	mi := toAdminMerchantInfo(updated)
	// 覆盖为最新白名单
	var payIds []int64
	for _, g := range gr.GetGrants() {
		if g != nil {
			payIds = append(payIds, g.GetPayProductId())
		}
	}
	var payoutIds []int64
	for _, g := range pr.GetGrants() {
		if g != nil {
			payoutIds = append(payoutIds, g.GetPayoutProductId())
		}
	}
	mi.PayProductIds = payIds
	mi.PayoutProductIds = payoutIds
	mi.CollectGrants = nil
	mi.PayoutGrants = nil
	for _, g := range gr.GetGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantCollectGrant{PayProductId: g.GetPayProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		mi.CollectGrants = append(mi.CollectGrants, row)
	}
	for _, g := range pr.GetGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayoutGrant{PayoutProductId: g.GetPayoutProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		row.FeeMode = g.GetFeeMode()
		if row.FeeMode <= 0 {
			row.FeeMode = 1
		}
		row.FeeFixedAmount = g.GetFeeFixedAmount()
		mi.PayoutGrants = append(mi.PayoutGrants, row)
	}
	return &types.AdminUpsertMerchantResp{Merchant: mi}, nil
}
