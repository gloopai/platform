// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	channelId := req.ChannelId
	if channelId <= 0 {
		r, err := l.svcCtx.ChannelRpc.Route(l.ctx, &channelclient.RouteReq{
			Amount:  req.Amount,
			PayType: req.PayType,
		})
		if err != nil {
			return nil, err
		}
		channelId = r.ChannelId
	}

	r, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &orderclient.CreateOrderReq{
		MerchantId:      req.MerchantId,
		MerchantOrderNo: req.MerchantOrderNo,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Subject:         req.Subject,
		ReturnUrl:       req.ReturnUrl,
		NotifyUrl:       req.NotifyUrl,
		PayType:         req.PayType,
		ChannelId:       channelId,
	})
	if err != nil {
		return nil, err
	}

	orderInfo := r.GetOrder()
	base := strings.TrimSpace(l.svcCtx.Config.CheckoutBaseUrl)
	if base == "" {
		base = "http://127.0.0.1:5174/"
	}
	base = strings.TrimRight(base, "/")
	checkoutUrl := base + "/?order_no=" + orderInfo.GetOrderNo()

	return &types.CreateOrderResp{
		OrderNo:     orderInfo.GetOrderNo(),
		Status:      orderInfo.GetStatus(),
		ChannelId:   orderInfo.GetChannelId(),
		CheckoutUrl: checkoutUrl,
	}, nil
}
