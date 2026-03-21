package logic

import (
	"context"

	"github.com/gloopai/pay/settle/internal/svc"
	"github.com/gloopai/pay/settle/settle/settle"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreditLogic {
	return &CreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreditLogic) Credit(in *settle.CreditReq) (*settle.CreditResp, error) {
	if in.GetMerchantId() == "" || in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	changed, balance, err := l.svcCtx.Settle.Credit(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetAmount(), in.GetReason())
	if err != nil {
		return nil, status.Error(codes.Internal, "credit failed")
	}
	return &settle.CreditResp{
		Changed: changed,
		Balance: balance,
	}, nil
}
