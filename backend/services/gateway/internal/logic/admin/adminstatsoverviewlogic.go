package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminStatsOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminStatsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminStatsOverviewLogic {
	return &AdminStatsOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminStatsOverviewLogic) AdminStatsOverview() (*types.AdminStatsOverviewResp, error) {
	tot, prods, chs, err := l.svcCtx.OrderStats.TodayOverview(l.ctx)
	if err != nil {
		return nil, err
	}
	rs, _ := l.svcCtx.RoutingSummary.Get(l.ctx)

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
