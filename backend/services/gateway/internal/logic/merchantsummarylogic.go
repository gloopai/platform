package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/merchant/merchantclient"
	"github.com/gloopai/pay/order/orderclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantSummaryLogic {
	return &MerchantSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantSummaryLogic) MerchantSummary(req *types.MerchantSummaryReq) (*types.MerchantSummaryResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(l.ctx))
	auth, err := l.svcCtx.MerchantRpc.GetAuthInfo(l.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	sum, err := l.svcCtx.OrderRpc.TodaySummary(l.ctx, &orderclient.TodaySummaryReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	var rate float64
	if sum.GetTotalCount() > 0 {
		rate = float64(sum.GetSuccessCount()) / float64(sum.GetTotalCount())
	}
	return &types.MerchantSummaryResp{
		TodayAmount: sum.GetTotalAmount(),
		TodayCount:  sum.GetTotalCount(),
		SuccessRate: rate,
		Balance:     auth.GetBalance(),
		MerchantId:  merchantId,
		NotifyUrl:   auth.GetNotifyUrl(),
		IpWhitelist: auth.GetIpWhitelist(),
	}, nil
}
