package server

import (
	"context"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/logic"
)

func (s *SettleServer) ListFundLogs(ctx context.Context, in *settlepb.ListFundLogsReq) (*settlepb.ListFundLogsResp, error) {
	l := logic.NewListFundLogsLogic(ctx, s.svcCtx)
	return l.ListFundLogs(in)
}

func (s *SettleServer) CreateWithdrawal(ctx context.Context, in *settlepb.CreateWithdrawalReq) (*settlepb.CreateWithdrawalResp, error) {
	l := logic.NewWithdrawalsLogic(ctx, s.svcCtx)
	return l.CreateWithdrawal(in)
}

func (s *SettleServer) ListWithdrawals(ctx context.Context, in *settlepb.ListWithdrawalsReq) (*settlepb.ListWithdrawalsResp, error) {
	l := logic.NewWithdrawalsLogic(ctx, s.svcCtx)
	return l.ListWithdrawals(in)
}

func (s *SettleServer) ReviewWithdrawal(ctx context.Context, in *settlepb.ReviewWithdrawalReq) (*settlepb.ReviewWithdrawalResp, error) {
	l := logic.NewWithdrawalsLogic(ctx, s.svcCtx)
	return l.ReviewWithdrawal(in)
}

func (s *SettleServer) PayoutWithdrawal(ctx context.Context, in *settlepb.PayoutWithdrawalReq) (*settlepb.PayoutWithdrawalResp, error) {
	l := logic.NewWithdrawalsLogic(ctx, s.svcCtx)
	return l.PayoutWithdrawal(in)
}
