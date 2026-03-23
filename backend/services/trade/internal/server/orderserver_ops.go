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

func (s *OrderServer) AdminListOrders(ctx context.Context, in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	l := logic.NewAdminListOrdersLogic(ctx, s.svcCtx)
	return l.AdminListOrders(in)
}
