package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/model"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/core/internal/configsync"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReplaceMerchantPayinProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplaceMerchantPayinProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplaceMerchantPayinProductsLogic {
	return &ReplaceMerchantPayinProductsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ReplaceMerchantPayinProductsLogic) ReplaceMerchantPayinProducts(in *merchantpb.ReplaceMerchantPayinProductsReq) (*merchantpb.ReplaceMerchantPayinProductsResp, error) {
	mid := strings.TrimSpace(in.GetMerchantId())
	if mid == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	var grants []model.PayinGrant
	for _, g := range in.GetGrants() {
		if g == nil || g.GetPayinProductId() <= 0 {
			continue
		}
		cg := model.PayinGrant{PayinProductID: g.GetPayinProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			cg.RateBps = &v
		}
		grants = append(grants, cg)
	}
	if err := l.svcCtx.MerchantPayinProducts.Replace(l.ctx, mid, grants); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_ = configsync.SyncMerchantPayinGrants(l.ctx, l.svcCtx.RuntimeConfig, l.svcCtx.MerchantPayinProducts, mid)
	return &merchantpb.ReplaceMerchantPayinProductsResp{Ok: true}, nil
}

type ListMerchantPayinProductIdsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantPayinProductIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantPayinProductIdsLogic {
	return &ListMerchantPayinProductIdsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMerchantPayinProductIdsLogic) ListMerchantPayinProductIds(in *merchantpb.ListMerchantPayinProductIdsReq) (*merchantpb.ListMerchantPayinProductIdsResp, error) {
	mid := strings.TrimSpace(in.GetMerchantId())
	if l.svcCtx.MerchantPayinGrantsSnapshot != nil {
		if g, ok := l.svcCtx.MerchantPayinGrantsSnapshot.Get(mid); ok && g != nil {
			return listPayinGrantsPBFromKV(g), nil
		}
	}
	grants, err := l.svcCtx.MerchantPayinProducts.ListPayinGrants(l.ctx, mid)
	if err != nil {
		return nil, err
	}
	return listPayinGrantsPB(grants), nil
}

func listPayinGrantsPBFromKV(g *configkv.MerchantPayinGrantsKV) *merchantpb.ListMerchantPayinProductIdsResp {
	if g == nil {
		return &merchantpb.ListMerchantPayinProductIdsResp{}
	}
	grants := kvcache.PayinGrantsModelFromKV(g)
	return listPayinGrantsPB(grants)
}

func listPayinGrantsPB(grants []model.PayinGrant) *merchantpb.ListMerchantPayinProductIdsResp {
	out := make([]*merchantpb.MerchantPayinGrant, 0, len(grants))
	for _, g := range grants {
		row := &merchantpb.MerchantPayinGrant{PayinProductId: g.PayinProductID}
		if g.RateBps != nil {
			v := *g.RateBps
			row.MerchantRateBps = &v
		}
		out = append(out, row)
	}
	return &merchantpb.ListMerchantPayinProductIdsResp{Grants: out}
}

type ReplaceMerchantPayoutProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplaceMerchantPayoutProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplaceMerchantPayoutProductsLogic {
	return &ReplaceMerchantPayoutProductsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ReplaceMerchantPayoutProductsLogic) ReplaceMerchantPayoutProducts(in *merchantpb.ReplaceMerchantPayoutProductsReq) (*merchantpb.ReplaceMerchantPayoutProductsResp, error) {
	mid := strings.TrimSpace(in.GetMerchantId())
	if mid == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	var grants []model.PayoutGrant
	for _, g := range in.GetGrants() {
		if g == nil || g.GetPayoutProductId() <= 0 {
			continue
		}
		pg := model.PayoutGrant{PayoutProductID: g.GetPayoutProductId()}
		pg.FeeMode = g.GetFeeMode()
		if pg.FeeMode <= 0 {
			pg.FeeMode = 1
		}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			pg.RateBps = &v
		}
		pg.FixedFeeAmount = g.GetFeeFixedAmount()
		grants = append(grants, pg)
	}
	if err := l.svcCtx.MerchantPayoutProducts.Replace(l.ctx, mid, grants); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_ = configsync.SyncMerchantPayoutGrants(l.ctx, l.svcCtx.RuntimeConfig, l.svcCtx.MerchantPayoutProducts, mid)
	return &merchantpb.ReplaceMerchantPayoutProductsResp{Ok: true}, nil
}

type ListMerchantPayoutProductIdsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantPayoutProductIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantPayoutProductIdsLogic {
	return &ListMerchantPayoutProductIdsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMerchantPayoutProductIdsLogic) ListMerchantPayoutProductIds(in *merchantpb.ListMerchantPayoutProductIdsReq) (*merchantpb.ListMerchantPayoutProductIdsResp, error) {
	mid := strings.TrimSpace(in.GetMerchantId())
	if l.svcCtx.MerchantPayoutGrantsSnapshot != nil {
		if g, ok := l.svcCtx.MerchantPayoutGrantsSnapshot.Get(mid); ok && g != nil {
			return listPayoutGrantsPB(kvcache.PayoutGrantsModelFromKV(g)), nil
		}
	}
	grants, err := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, mid)
	if err != nil {
		return nil, err
	}
	return listPayoutGrantsPB(grants), nil
}

func listPayoutGrantsPB(grants []model.PayoutGrant) *merchantpb.ListMerchantPayoutProductIdsResp {
	out := make([]*merchantpb.MerchantPayoutGrant, 0, len(grants))
	for _, g := range grants {
		row := &merchantpb.MerchantPayoutGrant{PayoutProductId: g.PayoutProductID}
		if g.RateBps != nil {
			v := *g.RateBps
			row.MerchantRateBps = &v
		}
		row.FeeMode = g.FeeMode
		if row.FeeMode <= 0 {
			row.FeeMode = 1
		}
		row.FeeFixedAmount = g.FixedFeeAmount
		out = append(out, row)
	}
	return &merchantpb.ListMerchantPayoutProductIdsResp{Grants: out}
}
