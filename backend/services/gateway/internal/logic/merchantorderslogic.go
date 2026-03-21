package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantOrdersLogic {
	return &MerchantOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantOrdersLogic) MerchantOrders(req *types.MerchantOrdersReq) (*types.MerchantOrdersResp, error) {
	status := int32(-1)
	if strings.TrimSpace(req.Status) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 32)
		if err != nil {
			return nil, err
		}
		status = int32(v)
	}
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(l.ctx))
	orders, err := l.svcCtx.Orders.ListByMerchant(l.ctx, merchantId, req.OrderNo, status, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]types.MerchantOrderItem, 0, len(orders))
	for _, o := range orders {
		out = append(out, types.MerchantOrderItem{
			OrderNo:         o.OrderNo,
			MerchantOrderNo: o.MerchantOrderNo,
			Amount:          o.Amount,
			Currency:        o.Currency,
			Status:          o.Status,
			ChannelId:       o.ChannelId,
			PaidAmount:      o.PaidAmount,
			UpstreamTradeNo: o.UpstreamTradeNo,
			CreatedAt:       o.CreatedAt.Unix(),
		})
	}
	return &types.MerchantOrdersResp{Orders: out}, nil
}
