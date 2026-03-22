package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminStats 管理后台订单与转化统计（仪表盘）。
type AdminStats struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminStats(ctx context.Context, svcCtx *svc.ServiceContext) *AdminStats {
	return &AdminStats{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (s *AdminStats) AdminStatsOverview() (*types.AdminStatsOverviewResp, error) {
	tot, prods, chs, err := s.svcCtx.OrderStats.TodayOverview(s.ctx)
	if err != nil {
		return nil, err
	}
	rs, _ := s.svcCtx.RoutingSummary.Get(s.ctx)

	totals := types.AdminStatsTotals{
		OrderCount:             tot.OrderCount,
		PaidAmount:             tot.PaidAmount,
		PaidCount:              tot.PaidCount,
		FailedCount:            tot.FailedCount,
		PendingCount:           tot.PendingCount,
		ClosedCount:            tot.ClosedCount,
		ConversionRatePct:      store.RateConversion(tot.PaidCount, tot.OrderCount),
		TerminalSuccessRatePct: store.RateTerminalSuccess(tot.PaidCount, tot.FailedCount),
	}

	outProd := make([]types.AdminStatsProductRow, 0, len(prods))
	for _, p := range prods {
		outProd = append(outProd, types.AdminStatsProductRow{
			ProductCode:            p.ProductCode,
			ProductName:            p.ProductName,
			OrderCount:             p.OrderCount,
			PaidAmount:             p.PaidAmount,
			PaidCount:              p.PaidCount,
			FailedCount:            p.FailedCount,
			ConversionRatePct:      store.RateConversion(p.PaidCount, p.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(p.PaidCount, p.FailedCount),
		})
	}

	outCh := make([]types.AdminStatsChannelRow, 0, len(chs))
	for _, c := range chs {
		outCh = append(outCh, types.AdminStatsChannelRow{
			ChannelId:              c.ChannelID,
			ChannelName:            c.ChannelName,
			OrderCount:             c.OrderCount,
			PaidAmount:             c.PaidAmount,
			PaidCount:              c.PaidCount,
			FailedCount:            c.FailedCount,
			ConversionRatePct:      store.RateConversion(c.PaidCount, c.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(c.PaidCount, c.FailedCount),
		})
	}

	return &types.AdminStatsOverviewResp{
		Range:           "today",
		Totals:          totals,
		ByPayProduct:    outProd,
		ByChannel:       outCh,
		EnabledChannels: rs.EnabledChannels,
		FusedChannels:   rs.FusedChannels,
	}, nil
}
