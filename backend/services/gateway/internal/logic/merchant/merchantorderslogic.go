package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
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
	r, err := l.svcCtx.OrderRpc.ListOrders(l.ctx, &orderclient.ListOrdersReq{
		MerchantId: merchantId,
		Keyword:    req.OrderNo,
		Status:     status,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}
	items := r.GetOrders()
	out := make([]types.MerchantOrderItem, 0, len(items))
	for _, o := range items {
		out = append(out, types.MerchantOrderItem{
			OrderNo:          o.GetOrderNo(),
			MerchantOrderNo:  o.GetMerchantOrderNo(),
			Amount:           o.GetAmount(),
			Currency:         o.GetCurrency(),
			Status:           o.GetStatus(),
			ChannelId:        o.GetChannelId(),
			PayProductCode:   o.GetPayProductCode(),
			PaidAmount:       o.GetPaidAmount(),
			UpstreamTradeNo:  o.GetUpstreamTradeNo(),
			CreatedAt:        o.GetCreatedAt(),
		})
	}
	return &types.MerchantOrdersResp{Orders: out}, nil
}
