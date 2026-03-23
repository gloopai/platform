package logic

import (
	"context"

	channelpb "github.com/gloopai/pay/common/pb/channel"
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
	s, err := r.svcCtx.ChannelRpc.GetRoutingSummary(r.ctx, &channelpb.GetRoutingSummaryReq{})
	if err != nil {
		return nil, err
	}
	return &types.AdminRoutingSummaryResp{
		AlgorithmKey:                 s.GetAlgorithmKey(),
		AlgorithmLabel:               s.GetAlgorithmLabel(),
		EnabledPayProducts:           s.GetEnabledPayProducts(),
		EnabledPayoutProducts:        s.GetEnabledPayoutProducts(),
		EnabledChannels:              s.GetEnabledChannels(),
		ActiveBindings:               s.GetActiveBindings(),
		ActivePayoutBindings:         s.GetActivePayoutBindings(),
		MerchantsWithPayinWhitelist:  s.GetMerchantsWithPayinWhitelist(),
		MerchantsWithPayoutWhitelist: s.GetMerchantsWithPayoutWhitelist(),
		FusedChannels:                s.GetFusedChannels(),
	}, nil
}
