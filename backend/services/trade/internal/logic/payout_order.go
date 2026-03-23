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

type PayoutOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayoutOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayoutOrderLogic {
	return &PayoutOrderLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *PayoutOrderLogic) ListPayoutOrders(in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	records, err := l.svcCtx.PayoutOrders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.ListOrdersResp{Orders: out}, nil
}

func (l *PayoutOrderLogic) AdminListPayoutOrders(in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	limit := in.GetLimit()
	st := int32(-1)
	if in.Status != nil {
		st = *in.Status
		if st < -1 || st > 3 {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
	}
	records, err := l.svcCtx.PayoutOrders.AdminList(l.ctx, strings.TrimSpace(in.GetMerchantId()), strings.TrimSpace(in.GetKeyword()), st, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.AdminListOrdersResp{Orders: out}, nil
}
