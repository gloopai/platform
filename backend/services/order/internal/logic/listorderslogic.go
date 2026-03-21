package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/order/internal/svc"
	"github.com/gloopai/pay/order/order/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *order.ListOrdersReq) (*order.ListOrdersResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	records, err := l.svcCtx.Orders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list orders failed")
	}

	out := make([]*order.OrderInfo, 0, len(records))
	for i := range records {
		rec := records[i]
		out = append(out, toOrderInfo(&rec))
	}
	return &order.ListOrdersResp{Orders: out}, nil
}
