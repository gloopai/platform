package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminRouting 管理后台路由策略说明与实时汇总（只读）。
type AdminRouting struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRouting(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRouting {
	return &AdminRouting{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (r *AdminRouting) AdminRoutingSummary() (*types.AdminRoutingSummaryResp, error) {
	s, err := r.svcCtx.RoutingSummary.Get(r.ctx)
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
