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

type AdminListPayOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPayOrdersLogic {
	return &AdminListPayOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminListPayOrdersLogic) AdminListPayOrders(in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	limit := in.GetLimit()
	st := int32(-1)
	if in.Status != nil {
		st = *in.Status
		if st < -1 || st > 3 {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
	}

	records, err := l.svcCtx.PayOrders.AdminList(l.ctx, strings.TrimSpace(in.GetMerchantId()), strings.TrimSpace(in.GetKeyword()), st, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin list orders failed")
	}

	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.AdminListOrdersResp{Orders: out}, nil
}
