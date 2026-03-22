package server

import (
	"context"

	"github.com/gloopai/pay/core/internal/logic"
	settlepb "github.com/gloopai/pay/common/pb/settle"
)

func (s *SettleServer) ListFundLogs(ctx context.Context, in *settlepb.ListFundLogsReq) (*settlepb.ListFundLogsResp, error) {
	l := logic.NewListFundLogsLogic(ctx, s.svcCtx)
	return l.ListFundLogs(in)
}
