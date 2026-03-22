// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderLogic {
	return &QueryOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOrderLogic) QueryOrder(req *types.QueryOrderReq) (resp *types.QueryOrderResp, err error) {
	r, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderReq{
		MerchantId:      req.MerchantId,
		OrderNo:         req.OrderNo,
		MerchantOrderNo: req.MerchantOrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	return &types.QueryOrderResp{
		Order: types.OrderInfo{
			OrderNo:          o.GetOrderNo(),
			MerchantId:       o.GetMerchantId(),
			MerchantOrderNo:  o.GetMerchantOrderNo(),
			Amount:           o.GetAmount(),
			Currency:         o.GetCurrency(),
			Status:           o.GetStatus(),
			ChannelId:        o.GetChannelId(),
			PayProductId:     o.GetPayProductId(),
			PayProductCode:   o.GetPayProductCode(),
			ChannelLocked:    o.GetChannelLocked(),
			PaidAmount:       o.GetPaidAmount(),
			ReturnUrl:        o.GetReturnUrl(),
			NotifyUrl:        o.GetNotifyUrl(),
			UpstreamTradeNo:  o.GetUpstreamTradeNo(),
		},
	}, nil
}
