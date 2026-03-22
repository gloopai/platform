package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	orderpb "github.com/gloopai/pay/common/pb/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *orderpb.GetOrderReq) (*orderpb.GetOrderResp, error) {
	var (
		rec *store.OrderRecord
		err error
	)
	switch {
	case in.GetOrderNo() != "":
		rec, err = l.svcCtx.Orders.FindByOrderNo(l.ctx, in.GetOrderNo())
	case in.GetMerchantId() != "" && in.GetMerchantOrderNo() != "":
		rec, err = l.svcCtx.Orders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	default:
		return nil, status.Error(codes.InvalidArgument, "order_no or (merchant_id and merchant_order_no) required")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "get order failed")
	}
	if in.GetMerchantId() != "" && rec.MerchantId != in.GetMerchantId() {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &orderpb.GetOrderResp{
		Order: toOrderInfo(rec),
	}, nil
}
