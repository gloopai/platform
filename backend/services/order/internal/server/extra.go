package server

import (
	"context"

	"github.com/gloopai/pay/order/internal/logic"
	"github.com/gloopai/pay/order/order/order"
)

func (s *OrderServer) ListOrders(ctx context.Context, in *order.ListOrdersReq) (*order.ListOrdersResp, error) {
	l := logic.NewListOrdersLogic(ctx, s.svcCtx)
	return l.ListOrders(in)
}

func (s *OrderServer) TodaySummary(ctx context.Context, in *order.TodaySummaryReq) (*order.TodaySummaryResp, error) {
	l := logic.NewTodaySummaryLogic(ctx, s.svcCtx)
	return l.TodaySummary(in)
}
