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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	merchantID := strings.TrimSpace(req.MerchantId)
	payType := strings.TrimSpace(req.PayType)
	channelID := req.ChannelId

	var (
		route          *channelclient.RouteResp
		payProductCode string
		channelLocked  int32
		cid            int64
		ppid           int64
	)

	switch {
	case channelID > 0:
		var code string
		ppid, code, err = l.svcCtx.PayProducts.ResolveLockedChannelForMerchant(l.ctx, merchantID, channelID, req.Amount)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		channelLocked = 1
		cid = channelID
		payProductCode = code

	case payType != "":
		ok, err := l.svcCtx.PayProducts.MerchantHasPayProductCode(l.ctx, merchantID, payType)
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "pay_type not enabled for this merchant")
		}
		route, err = l.svcCtx.ChannelRpc.Route(l.ctx, &channelclient.RouteReq{
			Amount:  req.Amount,
			PayType: payType,
		})
		if err != nil {
			return nil, err
		}
		cid = route.ChannelId
		ppid = route.PayProductId
		payProductCode = payType
		channelLocked = 0

	default:
		// 纯预下单：不在此处选路，收银台按商户白名单展示支付方式
		cid, ppid = 0, 0
		payProductCode = ""
		channelLocked = 0
	}

	r, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &orderclient.CreateOrderReq{
		MerchantId:      req.MerchantId,
		MerchantOrderNo: req.MerchantOrderNo,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Subject:         req.Subject,
		ReturnUrl:       req.ReturnUrl,
		NotifyUrl:       req.NotifyUrl,
		PayType:         payType,
		ChannelId:       cid,
		PayProductId:    ppid,
		PayProductCode:  payProductCode,
		ChannelLocked:   channelLocked,
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
	checkoutURL := base + "/?order_no=" + orderInfo.GetOrderNo()

	return &types.CreateOrderResp{
		OrderNo:         orderInfo.GetOrderNo(),
		Status:          orderInfo.GetStatus(),
		ChannelId:       orderInfo.GetChannelId(),
		PayProductId:    orderInfo.GetPayProductId(),
		PayProductCode:  orderInfo.GetPayProductCode(),
		CheckoutUrl:     checkoutURL,
		ChannelLocked:   orderInfo.GetChannelLocked(),
	}, nil
}
