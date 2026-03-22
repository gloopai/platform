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

	var items []types.PayProductItem
	if o.GetChannelLocked() != 0 {
		code := o.GetPayProductCode()
		name := code
		if code != "" {
			if dn, err := l.svcCtx.PayProducts.GetPayProductDisplayName(l.ctx, code); err == nil && dn != "" {
				name = dn
			}
		}
		if code != "" {
			items = []types.PayProductItem{{Code: code, Name: name}}
		}
	} else {
		opts, err := l.svcCtx.PayProducts.ListTerminalPayProducts(l.ctx, o.GetMerchantId(), o.GetAmount())
		if err != nil {
			return nil, err
		}
		items = make([]types.PayProductItem, 0, len(opts))
		for _, p := range opts {
			items = append(items, types.PayProductItem{Code: p.Code, Name: p.Name})
		}
	}

	return &types.TerminalOrderResp{
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
		PayProducts: items,
	}, nil
}
