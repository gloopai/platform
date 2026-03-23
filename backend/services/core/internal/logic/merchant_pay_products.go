package logic

import (
	"context"
	"strings"

	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReplaceMerchantPayProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplaceMerchantPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplaceMerchantPayProductsLogic {
	return &ReplaceMerchantPayProductsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ReplaceMerchantPayProductsLogic) ReplaceMerchantPayProducts(in *merchantpb.ReplaceMerchantPayProductsReq) (*merchantpb.ReplaceMerchantPayProductsResp, error) {
	mid := strings.TrimSpace(in.GetMerchantId())
	if mid == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	var grants []store.PayinGrant
	for _, g := range in.GetGrants() {
		if g == nil || g.GetPayProductId() <= 0 {
			continue
		}
		cg := store.PayinGrant{PayinProductID: g.GetPayProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			cg.RateBps = &v
		}
		grants = append(grants, cg)
	}
	if err := l.svcCtx.MerchantPayProducts.Replace(l.ctx, mid, grants); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &merchantpb.ReplaceMerchantPayProductsResp{Ok: true}, nil
}

type ListMerchantPayProductIdsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantPayProductIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantPayProductIdsLogic {
	return &ListMerchantPayProductIdsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMerchantPayProductIdsLogic) ListMerchantPayProductIds(in *merchantpb.ListMerchantPayProductIdsReq) (*merchantpb.ListMerchantPayProductIdsResp, error) {
	grants, err := l.svcCtx.MerchantPayProducts.ListPayinGrants(l.ctx, in.GetMerchantId())
	if err != nil {
		return nil, err
	}
	out := make([]*merchantpb.MerchantPayinGrant, 0, len(grants))
	for _, g := range grants {
		row := &merchantpb.MerchantPayinGrant{PayProductId: g.PayinProductID}
		if g.RateBps != nil {
			v := *g.RateBps
			row.MerchantRateBps = &v
		}
		out = append(out, row)
	}
	return &merchantpb.ListMerchantPayProductIdsResp{Grants: out}, nil
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
	var grants []store.PayoutGrant
	for _, g := range in.GetGrants() {
		if g == nil || g.GetPayoutProductId() <= 0 {
			continue
		}
		pg := store.PayoutGrant{PayoutProductID: g.GetPayoutProductId()}
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
	grants, err := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, in.GetMerchantId())
	if err != nil {
		return nil, err
	}
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
	return &merchantpb.ListMerchantPayoutProductIdsResp{Grants: out}, nil
}
