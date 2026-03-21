package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/order/orderclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantOrderDetailLogic {
	return &MerchantOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantOrderDetailLogic) MerchantOrderDetail(req *types.MerchantOrderDetailReq) (*types.MerchantOrderDetailResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(l.ctx))
	r, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderReq{
		MerchantId: merchantId,
		OrderNo:    req.OrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	logs, err := l.svcCtx.NotifyLogs.ListByOrder(l.ctx, merchantId, req.OrderNo, 50)
	if err != nil {
		return nil, err
	}
	outLogs := make([]types.MerchantNotifyLogItem, 0, len(logs))
	for _, x := range logs {
		outLogs = append(outLogs, types.MerchantNotifyLogItem{
			Id:           x.Id,
			NotifyUrl:    x.NotifyUrl,
			Attempt:      x.Attempt,
			HttpStatus:   x.HttpStatus,
			ResponseBody: x.ResponseBody,
			ErrorMsg:     x.ErrorMsg,
			CreatedAt:    x.CreatedAt.Unix(),
		})
	}

	return &types.MerchantOrderDetailResp{
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
		Logs: outLogs,
	}, nil
}
