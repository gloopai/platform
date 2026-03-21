package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListMerchantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListMerchantsLogic {
	return &AdminListMerchantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListMerchantsLogic) AdminListMerchants() (*types.AdminListMerchantsResp, error) {
	items, err := l.svcCtx.Merchants.List(l.ctx, 200)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminMerchantInfo, 0, len(items))
	for _, m := range items {
		out = append(out, types.AdminMerchantInfo{
			MerchantId:  m.MerchantId,
			ApiSecret:   m.ApiSecret,
			Status:      m.Status,
			RateBps:     m.RateBps,
			NotifyUrl:   m.NotifyUrl,
			ReturnUrl:   m.ReturnUrl,
			IpWhitelist: m.IpWhitelist,
			Balance:     m.Balance,
		})
	}
	return &types.AdminListMerchantsResp{Merchants: out}, nil
}

