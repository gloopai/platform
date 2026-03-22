package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TerminalPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTerminalPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TerminalPayLogic {
	return &TerminalPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TerminalPayLogic) TerminalPay(req *types.TerminalPayReq) (*types.TerminalPayResp, error) {
	orderNo := strings.TrimSpace(req.OrderNo)
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}

	gr, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderReq{OrderNo: orderNo})
	if err != nil {
		return nil, err
	}
	o := gr.GetOrder()
	code := strings.TrimSpace(req.PayProductCode)

	if o.GetChannelLocked() == 0 {
		if code == "" {
			return nil, status.Error(codes.InvalidArgument, "pay_product_code required")
		}
		ok, err := l.svcCtx.PayProducts.MerchantHasPayProductCode(l.ctx, o.GetMerchantId(), code)
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "pay_product_code not enabled for merchant")
		}
	}

	r, err := l.svcCtx.OrderRpc.PrepareTerminalPay(l.ctx, &orderclient.PrepareTerminalPayReq{
		OrderNo:        orderNo,
		PayProductCode: code,
	})
	if err != nil {
		return nil, err
	}
	return &types.TerminalPayResp{
		ChannelId:      r.GetChannelId(),
		PayProductId:   r.GetPayProductId(),
		PayProductCode: r.GetPayProductCode(),
		PayUrl:         r.GetPayUrl(),
		QrPayload:      r.GetQrPayload(),
		PayMode:        r.GetPayMode(),
	}, nil
}
