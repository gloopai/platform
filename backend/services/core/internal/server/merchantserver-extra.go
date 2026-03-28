package server

import (
	"context"

	"github.com/gloopai/pay/core/internal/logic"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
)

func (s *MerchantServer) ReplaceMerchantPayinProducts(ctx context.Context, in *merchantpb.ReplaceMerchantPayinProductsReq) (*merchantpb.ReplaceMerchantPayinProductsResp, error) {
	l := logic.NewReplaceMerchantPayinProductsLogic(ctx, s.svcCtx)
	return l.ReplaceMerchantPayinProducts(in)
}

func (s *MerchantServer) ListMerchantPayinProductIds(ctx context.Context, in *merchantpb.ListMerchantPayinProductIdsReq) (*merchantpb.ListMerchantPayinProductIdsResp, error) {
	l := logic.NewListMerchantPayinProductIdsLogic(ctx, s.svcCtx)
	return l.ListMerchantPayinProductIds(in)
}

func (s *MerchantServer) ReplaceMerchantPayoutProducts(ctx context.Context, in *merchantpb.ReplaceMerchantPayoutProductsReq) (*merchantpb.ReplaceMerchantPayoutProductsResp, error) {
	l := logic.NewReplaceMerchantPayoutProductsLogic(ctx, s.svcCtx)
	return l.ReplaceMerchantPayoutProducts(in)
}

func (s *MerchantServer) ListMerchantPayoutProductIds(ctx context.Context, in *merchantpb.ListMerchantPayoutProductIdsReq) (*merchantpb.ListMerchantPayoutProductIdsResp, error) {
	l := logic.NewListMerchantPayoutProductIdsLogic(ctx, s.svcCtx)
	return l.ListMerchantPayoutProductIds(in)
}
