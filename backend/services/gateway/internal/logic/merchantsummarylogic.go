package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
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
	m, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}
	amount, count, successCount, err := l.svcCtx.Orders.TodaySummary(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}
	var rate float64
	if count > 0 {
		rate = float64(successCount) / float64(count)
	}
	return &types.MerchantSummaryResp{
		TodayAmount: amount,
		TodayCount:  count,
		SuccessRate: rate,
		Balance:     m.Balance,
		MerchantId:  m.MerchantId,
		NotifyUrl:   m.NotifyUrl,
		IpWhitelist: m.IpWhitelist,
	}, nil
}
