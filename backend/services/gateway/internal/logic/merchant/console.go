package logic

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	orderpb "github.com/gloopai/pay/common/pb/order"
	settlepb "github.com/gloopai/pay/common/pb/settle"
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

func (c *MerchantConsole) MerchantDisplaySettings(req *types.MerchantDisplaySettingsReq) (*types.MerchantDisplaySettingsResp, error) {
	row, err := c.svcCtx.GlobalSettings.GetDisplaySettings(c.ctx)
	if err != nil {
		return nil, err
	}
	return &types.MerchantDisplaySettingsResp{
		CountryCode:    row.CountryCode,
		CurrencyCode:   row.CurrencyCode,
		CurrencySymbol: row.CurrencySymbol,
	}, nil
}

func (c *MerchantConsole) MerchantPayOrders(req *types.MerchantOrdersReq) (*types.MerchantOrdersResp, error) {
	return c.merchantOrders(req, false)
}

func (c *MerchantConsole) MerchantPayoutOrders(req *types.MerchantOrdersReq) (*types.MerchantOrdersResp, error) {
	return c.merchantOrders(req, true)
}

func (c *MerchantConsole) merchantOrders(req *types.MerchantOrdersReq, payout bool) (*types.MerchantOrdersResp, error) {
	orderStatus := int32(-1)
	if strings.TrimSpace(req.Status) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 32)
		if err != nil {
			return nil, err
		}
		orderStatus = int32(v)
	}
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	pbReq := &orderclient.ListOrdersReq{
		MerchantId: merchantId,
		Keyword:    req.OrderNo,
		Status:     orderStatus,
		Limit:      req.Limit,
	}
	var (
		r   *orderclient.ListOrdersResp
		err error
	)
	if payout {
		r, err = c.svcCtx.OrderRpc.ListPayoutOrders(c.ctx, pbReq)
	} else {
		r, err = c.svcCtx.OrderRpc.ListPayOrders(c.ctx, pbReq)
	}
	if err != nil {
		return nil, err
	}
	items := r.GetOrders()
	nameByCode := c.payProductNameByCode(c.ctx)
	out := make([]types.MerchantOrderItem, 0, len(items))
	for _, o := range items {
		code := o.GetPayProductCode()
		out = append(out, types.MerchantOrderItem{
			OrderNo:         o.GetOrderNo(),
			MerchantOrderNo: o.GetMerchantOrderNo(),
			Amount:          o.GetAmount(),
			Currency:        o.GetCurrency(),
			Status:          o.GetStatus(),
			ChannelId:       o.GetChannelId(),
			PayProductCode:  code,
			PayProductName:  lookupPayProductName(nameByCode, code),
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
	return &types.MerchantOrdersResp{Orders: out}, nil
}

func (c *MerchantConsole) MerchantProductStats(req *types.MerchantProductStatsReq) (*types.MerchantProductStatsResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	date := time.Now().Format("2006-01-02")
	r, err := c.svcCtx.OrderRpc.AdminDayOverview(c.ctx, &orderpb.AdminDayOverviewReq{
		Date:       date,
		MerchantId: merchantId,
	})
	if err != nil {
		return nil, err
	}
	nameByCode := c.payProductNameByCode(c.ctx)
	rows := r.GetByPayProduct()
	items := make([]types.MerchantProductStatsItem, 0, len(rows))
	for _, x := range rows {
		code := strings.TrimSpace(x.GetProductCode())
		name := strings.TrimSpace(x.GetProductName())
		if name == "" {
			name = lookupPayProductName(nameByCode, code)
		}
		if name == "" {
			name = code
		}
		items = append(items, types.MerchantProductStatsItem{
			PayProductCode: code,
			PayProductName: name,
			OrderCount:     x.GetOrderCount(),
			PaidAmount:     x.GetPaidAmount(),
			PaidCount:      x.GetPaidCount(),
			FailedCount:    x.GetFailedCount(),
			SuccessRatePct: x.GetTerminalSuccessRatePct(),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PaidAmount == items[j].PaidAmount {
			return items[i].OrderCount > items[j].OrderCount
		}
		return items[i].PaidAmount > items[j].PaidAmount
	})
	return &types.MerchantProductStatsResp{
		Date:       r.GetDate(),
		MerchantId: merchantId,
		Items:      items,
	}, nil
}

func (c *MerchantConsole) payProductNameByCode(ctx context.Context) map[string]string {
	r, err := c.svcCtx.ChannelRpc.AdminListPayProducts(ctx, &channelpb.AdminListPayProductsReq{})
	if err != nil {
		c.Errorf("AdminListPayProducts: %v", err)
		return nil
	}
	m := make(map[string]string, len(r.GetProducts()))
	for _, row := range r.GetProducts() {
		if row.GetCode() != "" {
			m[row.GetCode()] = row.GetName()
		}
	}
	return m
}

func lookupPayProductName(byCode map[string]string, code string) string {
	if code == "" || byCode == nil {
		return ""
	}
	return byCode[code]
}

func (c *MerchantConsole) MerchantFundLogs(req *types.MerchantFundLogsReq) (*types.MerchantFundLogsResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	r, err := c.svcCtx.SettleRpc.ListFundLogs(c.ctx, &settlepb.ListFundLogsReq{
		MerchantId: merchantId,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.MerchantFundLogItem, 0, len(r.GetLogs()))
	for _, f := range r.GetLogs() {
		out = append(out, types.MerchantFundLogItem{
			Id:            f.GetId(),
			OrderNo:       f.GetOrderNo(),
			ChangeType:    f.GetChangeType(),
			Amount:        f.GetAmount(),
			BalanceBefore: f.GetBalanceBefore(),
			BalanceAfter:  f.GetBalanceAfter(),
			Reason:        f.GetReason(),
			CreatedAt:     f.GetCreatedAt(),
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

	payProductName := ""
	if code := o.GetPayProductCode(); code != "" {
		if dn, err := c.svcCtx.ChannelRpc.GetPayProductDisplayName(c.ctx, &channelpb.GetPayProductDisplayNameReq{Code: code}); err == nil && dn != nil {
			payProductName = dn.GetName()
		}
	}

	nlr, err := c.svcCtx.OrderRpc.ListMerchantNotifyLogs(c.ctx, &orderpb.ListMerchantNotifyLogsReq{
		MerchantId: merchantId,
		OrderNo:    req.OrderNo,
		Limit:      50,
	})
	if err != nil {
		return nil, err
	}
	outLogs := make([]types.MerchantNotifyLogItem, 0, len(nlr.GetLogs()))
	for _, x := range nlr.GetLogs() {
		outLogs = append(outLogs, types.MerchantNotifyLogItem{
			Id:           x.GetId(),
			NotifyUrl:    x.GetNotifyUrl(),
			Attempt:      x.GetAttempt(),
			HttpStatus:   x.GetHttpStatus(),
			ResponseBody: x.GetResponseBody(),
			ErrorMsg:     x.GetErrorMsg(),
			CreatedAt:    x.GetCreatedAt(),
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
			PayProductName:  payProductName,
			ChannelLocked:   o.GetChannelLocked(),
			PaidAmount:      o.GetPaidAmount(),
			FeeMode:         o.GetFeeMode(),
			FeeRateBps:      o.GetFeeRateBps(),
			FeeFixedAmount:  o.GetFeeFixedAmount(),
			FeeAmount:       o.GetFeeAmount(),
			NetAmount:       o.GetNetAmount(),
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
