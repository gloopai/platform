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

type TerminalOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTerminalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TerminalOrderLogic {
	return &TerminalOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TerminalOrderLogic) TerminalOrder(req *types.TerminalOrderReq) (resp *types.TerminalOrderResp, err error) {
	r, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	return &types.TerminalOrderResp{
		Order: types.OrderInfo{
			OrderNo:         o.GetOrderNo(),
			MerchantId:      o.GetMerchantId(),
			MerchantOrderNo: o.GetMerchantOrderNo(),
			Amount:          o.GetAmount(),
			Currency:        o.GetCurrency(),
			Status:          o.GetStatus(),
			ChannelId:       o.GetChannelId(),
			ReturnUrl:       o.GetReturnUrl(),
			NotifyUrl:       o.GetNotifyUrl(),
			UpstreamTradeNo: o.GetUpstreamTradeNo(),
		},
	}, nil
}
