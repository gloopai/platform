package server

import (
	"context"

	"github.com/gloopai/pay/core/internal/logic"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
)

func (s *MerchantServer) ReplaceMerchantPayProducts(ctx context.Context, in *merchantpb.ReplaceMerchantPayProductsReq) (*merchantpb.ReplaceMerchantPayProductsResp, error) {
	l := logic.NewReplaceMerchantPayProductsLogic(ctx, s.svcCtx)
	return l.ReplaceMerchantPayProducts(in)
}

func (s *MerchantServer) ListMerchantPayProductIds(ctx context.Context, in *merchantpb.ListMerchantPayProductIdsReq) (*merchantpb.ListMerchantPayProductIdsResp, error) {
	l := logic.NewListMerchantPayProductIdsLogic(ctx, s.svcCtx)
	return l.ListMerchantPayProductIds(in)
}

func (s *MerchantServer) ReplaceMerchantPayoutProducts(ctx context.Context, in *merchantpb.ReplaceMerchantPayoutProductsReq) (*merchantpb.ReplaceMerchantPayoutProductsResp, error) {
	l := logic.NewReplaceMerchantPayoutProductsLogic(ctx, s.svcCtx)
	return l.ReplaceMerchantPayoutProducts(in)
}

func (s *MerchantServer) ListMerchantPayoutProductIds(ctx context.Context, in *merchantpb.ListMerchantPayoutProductIdsReq) (*merchantpb.ListMerchantPayoutProductIdsResp, error) {
	l := logic.NewListMerchantPayoutProductIdsLogic(ctx, s.svcCtx)
	return l.ListMerchantPayoutProductIds(in)
}
