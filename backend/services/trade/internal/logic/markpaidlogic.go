package logic

import (
	"context"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MarkPaidLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkPaidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkPaidLogic {
	return &MarkPaidLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkPaidLogic) MarkPaid(in *orderpb.MarkPaidReq) (*orderpb.MarkPaidResp, error) {
	if in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	if in.GetPaidAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "paid_amount must be positive")
	}

	changed, err := l.svcCtx.Orders.MarkPaid(l.ctx, in.GetOrderNo(), in.GetPaidAmount(), in.GetUpstreamTradeNo(), in.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.Internal, "mark paid failed")
	}
	return &orderpb.MarkPaidResp{Changed: changed}, nil
}
