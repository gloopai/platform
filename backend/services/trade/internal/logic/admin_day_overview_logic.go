package logic

import (
	"context"
	"strings"
	"time"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminDayOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDayOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDayOverviewLogic {
	return &AdminDayOverviewLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminDayOverviewLogic) AdminDayOverview(in *orderpb.AdminDayOverviewReq) (*orderpb.AdminDayOverviewResp, error) {
	ds := strings.TrimSpace(in.GetDate())
	if ds == "" {
		return nil, status.Error(codes.InvalidArgument, "date is required (YYYY-MM-DD)")
	}
	day, err := time.ParseInLocation("2006-01-02", ds, time.Local)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	tot, prods, chs, err := l.svcCtx.OrderStats.DayOverview(l.ctx, day)
	if err != nil {
		return nil, err
	}

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

	return &orderpb.AdminDayOverviewResp{
		Date:         ds,
		Totals:       totals,
		ByPayProduct: outProd,
		ByChannel:    outCh,
	}, nil
}
