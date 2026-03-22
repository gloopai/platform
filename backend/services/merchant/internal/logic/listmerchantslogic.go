package logic

import (
	"context"

	"github.com/gloopai/pay/merchant/internal/svc"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListMerchantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantsLogic {
	return &ListMerchantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMerchantsLogic) ListMerchants(in *merchantpb.ListMerchantsReq) (*merchantpb.ListMerchantsResp, error) {
	items, err := l.svcCtx.Merchants.List(l.ctx, in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*merchantpb.MerchantInfo, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, toMerchantInfo(&m))
	}
	return &merchantpb.ListMerchantsResp{Merchants: out}, nil
}
