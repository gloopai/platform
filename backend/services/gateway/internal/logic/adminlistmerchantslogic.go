package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/merchant/merchantclient"
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
	r, err := l.svcCtx.MerchantRpc.ListMerchants(l.ctx, &merchantclient.ListMerchantsReq{Limit: 200})
	if err != nil {
		return nil, err
	}
	items := r.GetMerchants()
	out := make([]types.AdminMerchantInfo, 0, len(items))
	for _, m := range items {
		out = append(out, types.AdminMerchantInfo{
			MerchantId:  m.GetMerchantId(),
			ApiSecret:   m.GetApiSecret(),
			Status:      m.GetStatus(),
			RateBps:     m.GetRateBps(),
			NotifyUrl:   m.GetNotifyUrl(),
			ReturnUrl:   m.GetReturnUrl(),
			IpWhitelist: m.GetIpWhitelist(),
			Balance:     m.GetBalance(),
		})
	}
	return &types.AdminListMerchantsResp{Merchants: out}, nil
}
