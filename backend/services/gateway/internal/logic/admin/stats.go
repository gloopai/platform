package logic

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
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
	r, err := s.svcCtx.OrderRpc.AdminTodayOverview(s.ctx, &orderpb.AdminTodayOverviewReq{})
	if err != nil {
		return nil, err
	}
	t := r.GetTotals()
	totals := types.AdminStatsTotals{
		OrderCount:             t.GetOrderCount(),
		PaidAmount:             t.GetPaidAmount(),
		PaidCount:              t.GetPaidCount(),
		FailedCount:            t.GetFailedCount(),
		PendingCount:           t.GetPendingCount(),
		ClosedCount:            t.GetClosedCount(),
		ConversionRatePct:      t.GetConversionRatePct(),
		TerminalSuccessRatePct: t.GetTerminalSuccessRatePct(),
	}

	outProd := make([]types.AdminStatsProductRow, 0, len(r.GetByPayinProduct()))
	for _, p := range r.GetByPayinProduct() {
		outProd = append(outProd, types.AdminStatsProductRow{
			ProductCode:            p.GetProductCode(),
			ProductName:            p.GetProductName(),
			OrderCount:             p.GetOrderCount(),
			PaidAmount:             p.GetPaidAmount(),
			PaidCount:              p.GetPaidCount(),
			FailedCount:            p.GetFailedCount(),
			ConversionRatePct:      p.GetConversionRatePct(),
			TerminalSuccessRatePct: p.GetTerminalSuccessRatePct(),
		})
	}

	outCh := make([]types.AdminStatsChannelRow, 0, len(r.GetByChannel()))
	for _, c := range r.GetByChannel() {
		outCh = append(outCh, types.AdminStatsChannelRow{
			ChannelId:              c.GetChannelId(),
			ChannelName:            c.GetChannelName(),
			OrderCount:             c.GetOrderCount(),
			PaidAmount:             c.GetPaidAmount(),
			PaidCount:              c.GetPaidCount(),
			FailedCount:            c.GetFailedCount(),
			ConversionRatePct:      c.GetConversionRatePct(),
			TerminalSuccessRatePct: c.GetTerminalSuccessRatePct(),
		})
	}

	return &types.AdminStatsOverviewResp{
		Range:           r.GetRange(),
		Totals:          totals,
		ByPayinProduct:    outProd,
		ByChannel:       outCh,
		EnabledChannels: r.GetEnabledChannels(),
		FusedChannels:   r.GetFusedChannels(),
	}, nil
}

func (s *AdminStats) AdminDayOverview(req *types.AdminDayOverviewReq) (*types.AdminDayOverviewResp, error) {
	r, err := s.svcCtx.OrderRpc.AdminDayOverview(s.ctx, &orderpb.AdminDayOverviewReq{
		Date:       req.Date,
		MerchantId: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	t := r.GetTotals()
	totals := types.AdminStatsTotals{
		OrderCount:             t.GetOrderCount(),
		PaidAmount:             t.GetPaidAmount(),
		PaidCount:              t.GetPaidCount(),
		FailedCount:            t.GetFailedCount(),
		PendingCount:           t.GetPendingCount(),
		ClosedCount:            t.GetClosedCount(),
		ConversionRatePct:      t.GetConversionRatePct(),
		TerminalSuccessRatePct: t.GetTerminalSuccessRatePct(),
	}

	outProd := make([]types.AdminStatsProductRow, 0, len(r.GetByPayinProduct()))
	for _, p := range r.GetByPayinProduct() {
		outProd = append(outProd, types.AdminStatsProductRow{
			ProductCode:            p.GetProductCode(),
			ProductName:            p.GetProductName(),
			OrderCount:             p.GetOrderCount(),
			PaidAmount:             p.GetPaidAmount(),
			PaidCount:              p.GetPaidCount(),
			FailedCount:            p.GetFailedCount(),
			ConversionRatePct:      p.GetConversionRatePct(),
			TerminalSuccessRatePct: p.GetTerminalSuccessRatePct(),
		})
	}

	outCh := make([]types.AdminStatsChannelRow, 0, len(r.GetByChannel()))
	for _, c := range r.GetByChannel() {
		outCh = append(outCh, types.AdminStatsChannelRow{
			ChannelId:              c.GetChannelId(),
			ChannelName:            c.GetChannelName(),
			OrderCount:             c.GetOrderCount(),
			PaidAmount:             c.GetPaidAmount(),
			PaidCount:              c.GetPaidCount(),
			FailedCount:            c.GetFailedCount(),
			ConversionRatePct:      c.GetConversionRatePct(),
			TerminalSuccessRatePct: c.GetTerminalSuccessRatePct(),
		})
	}

	return &types.AdminDayOverviewResp{
		Date:         r.GetDate(),
		Totals:       totals,
		ByPayinProduct: outProd,
		ByChannel:    outCh,
	}, nil
}
