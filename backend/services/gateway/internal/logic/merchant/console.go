package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MerchantConsole 商户控制台：订单、资金、通知等（需登录态）。
type MerchantConsole struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantConsole(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantConsole {
	return &MerchantConsole{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *MerchantConsole) MerchantSummary(req *types.MerchantSummaryReq) (*types.MerchantSummaryResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	auth, err := c.svcCtx.MerchantRpc.GetAuthInfo(c.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	sum, err := c.svcCtx.OrderRpc.TodaySummary(c.ctx, &orderclient.TodaySummaryReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	var rate float64
	if sum.GetTotalCount() > 0 {
		rate = float64(sum.GetSuccessCount()) / float64(sum.GetTotalCount())
	}
	return &types.MerchantSummaryResp{
		TodayAmount: sum.GetTotalAmount(),
		TodayCount:  sum.GetTotalCount(),
		SuccessRate: rate,
		Balance:     auth.GetBalance(),
		MerchantId:  merchantId,
		NotifyUrl:   auth.GetNotifyUrl(),
		IpWhitelist: auth.GetIpWhitelist(),
	}, nil
}

func (c *MerchantConsole) MerchantOrders(req *types.MerchantOrdersReq) (*types.MerchantOrdersResp, error) {
	orderStatus := int32(-1)
	if strings.TrimSpace(req.Status) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 32)
		if err != nil {
			return nil, err
		}
		orderStatus = int32(v)
	}
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	r, err := c.svcCtx.OrderRpc.ListOrders(c.ctx, &orderclient.ListOrdersReq{
		MerchantId: merchantId,
		Keyword:    req.OrderNo,
		Status:     orderStatus,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}
	items := r.GetOrders()
	out := make([]types.MerchantOrderItem, 0, len(items))
	for _, o := range items {
		out = append(out, types.MerchantOrderItem{
			OrderNo:         o.GetOrderNo(),
			MerchantOrderNo: o.GetMerchantOrderNo(),
			Amount:          o.GetAmount(),
			Currency:        o.GetCurrency(),
			Status:          o.GetStatus(),
			ChannelId:       o.GetChannelId(),
			PayProductCode:  o.GetPayProductCode(),
			PaidAmount:      o.GetPaidAmount(),
			UpstreamTradeNo: o.GetUpstreamTradeNo(),
			CreatedAt:       o.GetCreatedAt(),
		})
	}
	return &types.MerchantOrdersResp{Orders: out}, nil
}

func (c *MerchantConsole) MerchantFundLogs(req *types.MerchantFundLogsReq) (*types.MerchantFundLogsResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	logs, err := c.svcCtx.FundLogs.ListByMerchant(c.ctx, merchantId, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]types.MerchantFundLogItem, 0, len(logs))
	for _, f := range logs {
		out = append(out, types.MerchantFundLogItem{
			Id:            f.Id,
			OrderNo:       f.OrderNo,
			ChangeType:    f.ChangeType,
			Amount:        f.Amount,
			BalanceBefore: f.BalanceBefore,
			BalanceAfter:  f.BalanceAfter,
			Reason:        f.Reason,
			CreatedAt:     f.CreatedAt.Unix(),
		})
	}
	return &types.MerchantFundLogsResp{Logs: out}, nil
}

func (c *MerchantConsole) MerchantOrderDetail(req *types.MerchantOrderDetailReq) (*types.MerchantOrderDetailResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	r, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		MerchantId: merchantId,
		OrderNo:    req.OrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	logs, err := c.svcCtx.NotifyLogs.ListByOrder(c.ctx, merchantId, req.OrderNo, 50)
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
			PayProductId:    o.GetPayProductId(),
			PayProductCode:  o.GetPayProductCode(),
			ChannelLocked:   o.GetChannelLocked(),
			PaidAmount:      o.GetPaidAmount(),
			ReturnUrl:       o.GetReturnUrl(),
			NotifyUrl:       o.GetNotifyUrl(),
			UpstreamTradeNo: o.GetUpstreamTradeNo(),
		},
		Logs: outLogs,
	}, nil
}

func (c *MerchantConsole) MerchantRetryNotify(req *types.MerchantRetryNotifyReq) (*types.MerchantRetryNotifyResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	if merchantId == "" || req.OrderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	body, _ := json.Marshal(map[string]any{
		"merchant_id": merchantId,
		"order_no":    req.OrderNo,
		"attempt":     0,
	})
	if err := c.svcCtx.NsqProducer.Publish(c.svcCtx.Config.Nsq.Topic, body); err != nil {
		return nil, err
	}
	return &types.MerchantRetryNotifyResp{Ok: true}, nil
}
