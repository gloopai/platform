package server

import (
	"context"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/logic"
	"github.com/gloopai/pay/core/internal/svc"
)

type SettleServer struct {
	svcCtx *svc.ServiceContext
	settlepb.UnimplementedSettleServer
}

func NewSettleServer(svcCtx *svc.ServiceContext) *SettleServer {
	return &SettleServer{
		svcCtx: svcCtx,
	}
}

func (s *SettleServer) Credit(ctx context.Context, in *settlepb.CreditReq) (*settlepb.CreditResp, error) {
	l := logic.NewCreditLogic(ctx, s.svcCtx)
	return l.Credit(in)
}
