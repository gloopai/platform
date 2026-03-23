package server

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/logic"
)

func (s *OrderServer) TodaySummary(ctx context.Context, in *orderpb.TodaySummaryReq) (*orderpb.TodaySummaryResp, error) {
	l := logic.NewTodaySummaryLogic(ctx, s.svcCtx)
	return l.TodaySummary(in)
}

func (s *OrderServer) PrepareTerminalPay(ctx context.Context, in *orderpb.PrepareTerminalPayReq) (*orderpb.PrepareTerminalPayResp, error) {
	l := logic.NewPrepareTerminalPayLogic(ctx, s.svcCtx)
	return l.PrepareTerminalPay(in)
}
