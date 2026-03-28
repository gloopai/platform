package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (a *AdminOrders) AdminListPayOrders(req *types.AdminOrdersReq) (*types.AdminOrdersResp, error) {
	return a.adminListOrders(req, false)
}

func (a *AdminOrders) AdminListPayoutOrders(req *types.AdminOrdersReq) (*types.AdminOrdersResp, error) {
	return a.adminListOrders(req, true)
}

func (a *AdminOrders) adminListOrders(req *types.AdminOrdersReq, payout bool) (*types.AdminOrdersResp, error) {
	pbreq := &orderpb.AdminListOrdersReq{
		MerchantId: strings.TrimSpace(req.MerchantId),
		Keyword:    strings.TrimSpace(req.Keyword),
		Limit:      req.Limit,
		Offset:     req.Offset,
	}
	if strings.TrimSpace(req.Status) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 32)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		s := int32(v)
		pbreq.Status = &s
	}

	var (
		r   *orderpb.AdminListOrdersResp
		err error
	)
	if payout {
		r, err = a.svcCtx.OrderRpc.AdminListPayoutOrders(a.ctx, pbreq)
	} else {
		r, err = a.svcCtx.OrderRpc.AdminListPayOrders(a.ctx, pbreq)
	}
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
			ChannelName:     o.GetChannelName(),
			PayinProductId:    o.GetPayinProductId(),
			PayinProductCode:  o.GetPayinProductCode(),
			PaidAmount:      o.GetPaidAmount(),
			FeeMode:         o.GetFeeMode(),
			FeeRateBps:      o.GetFeeRateBps(),
			FeeFixedAmount:  o.GetFeeFixedAmount(),
			FeeAmount:       o.GetFeeAmount(),
			NetAmount:       o.GetNetAmount(),
			ChannelTradeNo: o.GetChannelTradeNo(),
			CreatedAt:       o.GetCreatedAt(),
		})
	}
	return &types.AdminOrdersResp{Orders: out, Total: r.GetTotal()}, nil
}

func (a *AdminOrders) AdminMockPayoutSuccess(req *types.AdminMockPayoutSuccessReq) (*types.AdminMockPayoutSuccessResp, error) {
	orderNo := strings.TrimSpace(req.OrderNo)
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	channelTradeNo := strings.TrimSpace(req.ChannelTradeNo)
	if channelTradeNo == "" {
		channelTradeNo = fmt.Sprintf("UP-MOCK-%d", time.Now().UnixNano())
	}
	changed, err := a.svcCtx.ServiceHub.MarkPayoutSuccess(a.ctx, orderNo, channelTradeNo)
	if err != nil {
		return nil, status.Error(codes.Internal, "mock payout success failed")
	}
	newStatus := int32(1)
	if !changed {
		// 如果未更新，可能已是成功或非待处理，返回 changed=false 给调用方判定。
		newStatus = -1
	}
	return &types.AdminMockPayoutSuccessResp{
		Ok:        true,
		OrderNo:   orderNo,
		Changed:   changed,
		NewStatus: newStatus,
	}, nil
}
