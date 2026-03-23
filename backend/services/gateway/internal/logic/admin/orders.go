package logic

import (
	"context"
	"strconv"
	"strings"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminOrders 管理台全站订单列表（只读，MVP）。
type AdminOrders struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrders(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrders {
	return &AdminOrders{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminOrders) AdminListOrders(req *types.AdminOrdersReq) (*types.AdminOrdersResp, error) {
	pbreq := &orderpb.AdminListOrdersReq{
		MerchantId: strings.TrimSpace(req.MerchantId),
		Keyword:    strings.TrimSpace(req.Keyword),
		Limit:      req.Limit,
	}
	if strings.TrimSpace(req.Status) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 32)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		s := int32(v)
		pbreq.Status = &s
	}

	r, err := a.svcCtx.OrderRpc.AdminListOrders(a.ctx, pbreq)
	if err != nil {
		return nil, err
	}
	rows := r.GetOrders()
	out := make([]types.AdminOrderRow, 0, len(rows))
	for _, o := range rows {
		out = append(out, types.AdminOrderRow{
			OrderNo:         o.GetOrderNo(),
			MerchantId:      o.GetMerchantId(),
			MerchantOrderNo: o.GetMerchantOrderNo(),
			Amount:          o.GetAmount(),
			Currency:        o.GetCurrency(),
			Status:          o.GetStatus(),
			ChannelId:       o.GetChannelId(),
			PayProductId:    o.GetPayProductId(),
			PayProductCode:  o.GetPayProductCode(),
			PaidAmount:      o.GetPaidAmount(),
			FeeMode:         o.GetFeeMode(),
			FeeRateBps:      o.GetFeeRateBps(),
			FeeFixedAmount:  o.GetFeeFixedAmount(),
			FeeAmount:       o.GetFeeAmount(),
			NetAmount:       o.GetNetAmount(),
			UpstreamTradeNo: o.GetUpstreamTradeNo(),
			CreatedAt:       o.GetCreatedAt(),
		})
	}
	return &types.AdminOrdersResp{Orders: out}, nil
}
