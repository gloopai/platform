package server

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/logic"
)

func (s *OrderServer) AdminTodayOverview(ctx context.Context, in *orderpb.AdminTodayOverviewReq) (*orderpb.AdminTodayOverviewResp, error) {
	l := logic.NewAdminTodayOverviewLogic(ctx, s.svcCtx)
	return l.AdminTodayOverview(in)
}

func (s *OrderServer) ListMerchantNotifyLogs(ctx context.Context, in *orderpb.ListMerchantNotifyLogsReq) (*orderpb.ListMerchantNotifyLogsResp, error) {
	l := logic.NewListMerchantNotifyLogsLogic(ctx, s.svcCtx)
	return l.ListMerchantNotifyLogs(in)
}

func (s *OrderServer) ListPayOrders(ctx context.Context, in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	l := logic.NewListPayOrdersLogic(ctx, s.svcCtx)
	return l.ListPayOrders(in)
}

func (s *OrderServer) ListPayoutOrders(ctx context.Context, in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	l := logic.NewPayoutOrderLogic(ctx, s.svcCtx)
	return l.ListPayoutOrders(in)
}

func (s *OrderServer) AdminListPayOrders(ctx context.Context, in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	l := logic.NewAdminListPayOrdersLogic(ctx, s.svcCtx)
	return l.AdminListPayOrders(in)
}

func (s *OrderServer) AdminListPayoutOrders(ctx context.Context, in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	l := logic.NewPayoutOrderLogic(ctx, s.svcCtx)
	return l.AdminListPayoutOrders(in)
}

func (s *OrderServer) AdminDayOverview(ctx context.Context, in *orderpb.AdminDayOverviewReq) (*orderpb.AdminDayOverviewResp, error) {
	l := logic.NewAdminDayOverviewLogic(ctx, s.svcCtx)
	return l.AdminDayOverview(in)
}
