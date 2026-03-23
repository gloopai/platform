package logic

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminTodayOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminTodayOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminTodayOverviewLogic {
	return &AdminTodayOverviewLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminTodayOverviewLogic) AdminTodayOverview(*orderpb.AdminTodayOverviewReq) (*orderpb.AdminTodayOverviewResp, error) {
	tot, prods, chs, err := l.svcCtx.OrderStats.TodayOverview(l.ctx)
	if err != nil {
		return nil, err
	}
	rs, _ := l.svcCtx.RoutingSummary.Get(l.ctx)

	totals := &orderpb.AdminStatsTotals{
		OrderCount:             tot.OrderCount,
		PaidAmount:             tot.PaidAmount,
		PaidCount:              tot.PaidCount,
		FailedCount:            tot.FailedCount,
		PendingCount:           tot.PendingCount,
		ClosedCount:            tot.ClosedCount,
		ConversionRatePct:      store.RateConversion(tot.PaidCount, tot.OrderCount),
		TerminalSuccessRatePct: store.RateTerminalSuccess(tot.PaidCount, tot.FailedCount),
	}

	outProd := make([]*orderpb.AdminStatsProductRow, 0, len(prods))
	for _, p := range prods {
		outProd = append(outProd, &orderpb.AdminStatsProductRow{
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

	outCh := make([]*orderpb.AdminStatsChannelRow, 0, len(chs))
	for _, c := range chs {
		outCh = append(outCh, &orderpb.AdminStatsChannelRow{
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

	return &orderpb.AdminTodayOverviewResp{
		Range:           "today",
		Totals:          totals,
		ByPayinProduct:    outProd,
		ByChannel:       outCh,
		EnabledChannels: rs.EnabledChannels,
		FusedChannels:   rs.FusedChannels,
	}, nil
}

type ListMerchantNotifyLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantNotifyLogsLogic {
	return &ListMerchantNotifyLogsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMerchantNotifyLogsLogic) ListMerchantNotifyLogs(in *orderpb.ListMerchantNotifyLogsReq) (*orderpb.ListMerchantNotifyLogsResp, error) {
	rows, err := l.svcCtx.NotifyLogs.ListByOrder(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*orderpb.MerchantNotifyLogItem, 0, len(rows))
	for _, x := range rows {
		out = append(out, &orderpb.MerchantNotifyLogItem{
			Id:           x.Id,
			NotifyUrl:    x.NotifyUrl,
			Attempt:      x.Attempt,
			HttpStatus:   x.HttpStatus,
			ResponseBody: x.ResponseBody,
			ErrorMsg:     x.ErrorMsg,
			CreatedAt:    x.CreatedAt.Unix(),
		})
	}
	return &orderpb.ListMerchantNotifyLogsResp{Logs: out}, nil
}
