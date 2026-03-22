package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminRoutingSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRoutingSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRoutingSummaryLogic {
	return &AdminRoutingSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminRoutingSummaryLogic) AdminRoutingSummary() (*types.AdminRoutingSummaryResp, error) {
	s, err := l.svcCtx.RoutingSummary.Get(l.ctx)
	if err != nil {
		return nil, err
	}
	return &types.AdminRoutingSummaryResp{
		AlgorithmKey:           "weighted_random_within_product",
		AlgorithmLabel:         "支付产品内加权随机（同产品多上游按权重分流）",
		EnabledPayProducts:     s.EnabledPayProducts,
		EnabledChannels:        s.EnabledChannels,
		ActiveBindings:         s.ActiveBindings,
		MerchantsWithWhitelist: s.MerchantsWithWhitelist,
		FusedChannels:          s.FusedChannels,
	}, nil
}
