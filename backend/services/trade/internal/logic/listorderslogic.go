package logic

import (
	"context"
	"strings"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/svc"

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

func (l *ListOrdersLogic) ListOrders(in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	records, err := l.svcCtx.Orders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list orders failed")
	}

	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		rec := records[i]
		out = append(out, toOrderInfo(&rec))
	}
	return &orderpb.ListOrdersResp{Orders: out}, nil
}
