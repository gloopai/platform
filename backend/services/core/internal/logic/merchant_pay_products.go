package logic

import (
	"context"
	"strings"

	merchantpb "github.com/gloopai/pay/common/pb/merchant"
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
	if err := l.svcCtx.MerchantPayProducts.Replace(l.ctx, mid, in.GetPayProductIds()); err != nil {
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
	ids, err := l.svcCtx.MerchantPayProducts.ListProductIDs(l.ctx, in.GetMerchantId())
	if err != nil {
		return nil, err
	}
	return &merchantpb.ListMerchantPayProductIdsResp{PayProductIds: ids}, nil
}
