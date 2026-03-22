package server

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/logic"
)

func (s *OrderServer) ListOrders(ctx context.Context, in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	l := logic.NewListOrdersLogic(ctx, s.svcCtx)
	return l.ListOrders(in)
}

func (s *OrderServer) TodaySummary(ctx context.Context, in *orderpb.TodaySummaryReq) (*orderpb.TodaySummaryResp, error) {
	l := logic.NewTodaySummaryLogic(ctx, s.svcCtx)
	return l.TodaySummary(in)
}
