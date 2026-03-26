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
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
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
		TodayAmount:      sum.GetTotalAmount(),
		TodayCount:       sum.GetTotalCount(),
		SuccessRate:      rate,
		PayinBalance:     auth.GetPayinBalance(),
		AvailableBalance: auth.GetAvailableBalance(),
		Status:           auth.GetStatus(),
		MerchantId:       merchantId,
		AppId:            auth.GetAppId(),
		Email:            auth.GetEmail(),
		AppSecret:        auth.GetAppSecret(),
		NotifyUrl:        auth.GetNotifyUrl(),
		ReturnUrl:        auth.GetReturnUrl(),
		IpWhitelist:      auth.GetIpWhitelist(),
	}, nil
}

func (c *MerchantConsole) MerchantUpdateConfig(req *types.MerchantUpdateConfigReq) (*types.MerchantUpdateConfigResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	if merchantId == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant not authenticated")
	}
	current, err := c.svcCtx.MerchantRpc.GetMerchant(c.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	m := current.GetMerchant()
	if m == nil {
		return nil, status.Error(codes.NotFound, "merchant not found")
	}
	updated, err := c.svcCtx.MerchantRpc.UpdateMerchant(c.ctx, &merchantclient.UpdateMerchantReq{
		MerchantId:           merchantId,
		AppSecret:            m.GetAppSecret(),
		Status:               m.GetStatus(),
		DefaultPayinRateBps:  m.GetDefaultPayinRateBps(),
		DefaultPayoutRateBps: m.GetDefaultPayoutRateBps(),
		NotifyUrl:            strings.TrimSpace(req.NotifyUrl),
		ReturnUrl:            m.GetReturnUrl(),
		IpWhitelist:          strings.TrimSpace(req.IpWhitelist),
	})
	if err != nil {
		return nil, err
	}
	out := updated.GetMerchant()
	if out == nil {
		out = m
	}
	return &types.MerchantUpdateConfigResp{
		MerchantId:  out.GetMerchantId(),
		AppId:       out.GetAppId(),
		Email:       out.GetEmail(),
		AppSecret:   out.GetAppSecret(),
		NotifyUrl:   out.GetNotifyUrl(),
		IpWhitelist: out.GetIpWhitelist(),
	}, nil
}

func (c *MerchantConsole) MerchantDisplaySettings(req *types.MerchantDisplaySettingsReq) (*types.MerchantDisplaySettingsResp, error) {
	row, err := c.svcCtx.ServiceHub.GetDisplaySettings(c.ctx)
	if err != nil {
		return nil, err
	}
	return &types.MerchantDisplaySettingsResp{
		CountryCode:    row.GetCountryCode(),
		CurrencyCode:   row.GetCurrencyCode(),
		CurrencySymbol: row.GetCurrencySymbol(),
		FiatToUsdtRate: row.GetFiatToUsdtRate(),
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
		Offset:     req.Offset,
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
		code := o.GetPayinProductCode()
		out = append(out, types.MerchantOrderItem{
			OrderNo:          o.GetOrderNo(),
			MerchantOrderNo:  o.GetMerchantOrderNo(),
			Amount:           o.GetAmount(),
			Currency:         o.GetCurrency(),
			Status:           o.GetStatus(),
			ChannelId:        o.GetChannelId(),
			PayinProductCode: code,
			PayinProductName: lookupPayinProductName(nameByCode, code),
			PaidAmount:       o.GetPaidAmount(),
			FeeMode:          o.GetFeeMode(),
			FeeRateBps:       o.GetFeeRateBps(),
			FeeFixedAmount:   o.GetFeeFixedAmount(),
			FeeAmount:        o.GetFeeAmount(),
			NetAmount:        o.GetNetAmount(),
			UpstreamTradeNo:  o.GetUpstreamTradeNo(),
			CreatedAt:        o.GetCreatedAt(),
		})
	}
	return &types.MerchantOrdersResp{Orders: out, Total: r.GetTotal()}, nil
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
	rows := r.GetByPayinProduct()
	items := make([]types.MerchantProductStatsItem, 0, len(rows))
	for _, x := range rows {
		code := strings.TrimSpace(x.GetProductCode())
		name := strings.TrimSpace(x.GetProductName())
		if name == "" {
			name = lookupPayinProductName(nameByCode, code)
		}
		if name == "" {
			name = code
		}
		items = append(items, types.MerchantProductStatsItem{
			PayinProductCode: code,
			PayinProductName: name,
			OrderCount:       x.GetOrderCount(),
			PaidAmount:       x.GetPaidAmount(),
			PaidCount:        x.GetPaidCount(),
			FailedCount:      x.GetFailedCount(),
			SuccessRatePct:   x.GetTerminalSuccessRatePct(),
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

func (c *MerchantConsole) MerchantOpenedProducts() (*types.MerchantOpenedProductsResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	if merchantId == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant not authenticated")
	}
	payinIDsResp, err := c.svcCtx.MerchantRpc.ListMerchantPayinProductIds(c.ctx, &merchantpb.ListMerchantPayinProductIdsReq{
		MerchantId: merchantId,
	})
	if err != nil {
		return nil, err
	}
	payoutIDsResp, err := c.svcCtx.MerchantRpc.ListMerchantPayoutProductIds(c.ctx, &merchantpb.ListMerchantPayoutProductIdsReq{
		MerchantId: merchantId,
	})
	if err != nil {
		return nil, err
	}
	payinByID := c.payinProductByID(c.ctx)
	payoutByID := c.payoutProductByID(c.ctx)
	items := make([]types.MerchantOpenedProductItem, 0, len(payinIDsResp.GetGrants())+len(payoutIDsResp.GetGrants()))
	for _, g := range payinIDsResp.GetGrants() {
		if g == nil || g.GetPayinProductId() <= 0 {
			continue
		}
		row := payinByID[g.GetPayinProductId()]
		var feeRateBps *int64
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			feeRateBps = &v
		}
		items = append(items, types.MerchantOpenedProductItem{
			ProductType: "payin",
			ProductId:   g.GetPayinProductId(),
			ProductCode: productCodeFromPayinRow(row),
			ProductName: productNameFromPayinRow(row, g.GetPayinProductId()),
			Enabled:     productEnabledFromPayinRow(row),
			FeeMode:     1,
			FeeRateBps:  feeRateBps,
		})
	}
	for _, g := range payoutIDsResp.GetGrants() {
		if g == nil || g.GetPayoutProductId() <= 0 {
			continue
		}
		row := payoutByID[g.GetPayoutProductId()]
		var feeRateBps *int64
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			feeRateBps = &v
		}
		feeMode := g.GetFeeMode()
		if feeMode <= 0 {
			feeMode = 1
		}
		items = append(items, types.MerchantOpenedProductItem{
			ProductType: "payout",
			ProductId:   g.GetPayoutProductId(),
			ProductCode: productCodeFromPayoutRow(row),
			ProductName: productNameFromPayoutRow(row, g.GetPayoutProductId()),
			Enabled:     productEnabledFromPayoutRow(row),
			FeeMode:     feeMode,
			FeeRateBps:  feeRateBps,
			FeeFixed:    g.GetFeeFixedAmount(),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ProductType == items[j].ProductType {
			if items[i].ProductName == items[j].ProductName {
				return items[i].ProductId < items[j].ProductId
			}
			return items[i].ProductName < items[j].ProductName
		}
		return items[i].ProductType < items[j].ProductType
	})
	return &types.MerchantOpenedProductsResp{
		MerchantId: merchantId,
		Products:   items,
	}, nil
}

func (c *MerchantConsole) payProductNameByCode(ctx context.Context) map[string]string {
	r, err := c.svcCtx.ChannelRpc.AdminListPayinProducts(ctx, &channelpb.AdminListPayinProductsReq{})
	if err != nil {
		c.Errorf("AdminListPayinProducts: %v", err)
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

func (c *MerchantConsole) payinProductByID(ctx context.Context) map[int64]*channelpb.AdminPayinProductRow {
	r, err := c.svcCtx.ChannelRpc.AdminListPayinProducts(ctx, &channelpb.AdminListPayinProductsReq{})
	if err != nil {
		c.Errorf("AdminListPayinProducts: %v", err)
		return nil
	}
	m := make(map[int64]*channelpb.AdminPayinProductRow, len(r.GetProducts()))
	for _, row := range r.GetProducts() {
		if row.GetId() > 0 {
			m[row.GetId()] = row
		}
	}
	return m
}

func (c *MerchantConsole) payoutProductByID(ctx context.Context) map[int64]*channelpb.AdminPayoutProductRow {
	r, err := c.svcCtx.ChannelRpc.AdminListPayoutProducts(ctx, &channelpb.AdminListPayoutProductsReq{})
	if err != nil {
		c.Errorf("AdminListPayoutProducts: %v", err)
		return nil
	}
	m := make(map[int64]*channelpb.AdminPayoutProductRow, len(r.GetProducts()))
	for _, row := range r.GetProducts() {
		if row.GetId() > 0 {
			m[row.GetId()] = row
		}
	}
	return m
}

func lookupPayinProductName(byCode map[string]string, code string) string {
	if code == "" || byCode == nil {
		return ""
	}
	return byCode[code]
}

func productCodeFromPayinRow(row *channelpb.AdminPayinProductRow) string {
	if row == nil {
		return ""
	}
	return row.GetCode()
}

func productNameFromPayinRow(row *channelpb.AdminPayinProductRow, id int64) string {
	if row == nil {
		return "代收产品#" + strconv.FormatInt(id, 10)
	}
	if strings.TrimSpace(row.GetName()) != "" {
		return row.GetName()
	}
	if strings.TrimSpace(row.GetCode()) != "" {
		return row.GetCode()
	}
	return "代收产品#" + strconv.FormatInt(id, 10)
}

func productEnabledFromPayinRow(row *channelpb.AdminPayinProductRow) bool {
	if row == nil {
		return true
	}
	return row.GetEnabled()
}

func productCodeFromPayoutRow(row *channelpb.AdminPayoutProductRow) string {
	if row == nil {
		return ""
	}
	return row.GetCode()
}

func productNameFromPayoutRow(row *channelpb.AdminPayoutProductRow, id int64) string {
	if row == nil {
		return "代付产品#" + strconv.FormatInt(id, 10)
	}
	if strings.TrimSpace(row.GetName()) != "" {
		return row.GetName()
	}
	if strings.TrimSpace(row.GetCode()) != "" {
		return row.GetCode()
	}
	return "代付产品#" + strconv.FormatInt(id, 10)
}

func productEnabledFromPayoutRow(row *channelpb.AdminPayoutProductRow) bool {
	if row == nil {
		return true
	}
	return row.GetEnabled()
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

func (c *MerchantConsole) MerchantTransferPayinToPayout(req *types.MerchantTransferPayinToPayoutReq) (*types.MerchantTransferPayinToPayoutResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(c.ctx))
	if merchantId == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant not authenticated")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	r, err := c.svcCtx.SettleRpc.TransferPayinToPayout(c.ctx, &settlepb.TransferPayinToPayoutReq{
		MerchantId: merchantId,
		Amount:     req.Amount,
		Reason:     strings.TrimSpace(req.Reason),
	})
	if err != nil {
		return nil, err
	}
	return &types.MerchantTransferPayinToPayoutResp{
		Ok:               r.GetChanged(),
		PayinBalance:     r.GetPayinBalance(),
		AvailableBalance: r.GetAvailableBalance(),
	}, nil
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
	if code := o.GetPayinProductCode(); code != "" {
		if dn, err := c.svcCtx.ChannelRpc.GetPayinProductDisplayName(c.ctx, &channelpb.GetPayinProductDisplayNameReq{Code: code}); err == nil && dn != nil {
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
			OrderNo:          o.GetOrderNo(),
			MerchantId:       o.GetMerchantId(),
			MerchantOrderNo:  o.GetMerchantOrderNo(),
			Amount:           o.GetAmount(),
			Currency:         o.GetCurrency(),
			Status:           o.GetStatus(),
			ChannelId:        o.GetChannelId(),
			PayinProductId:   o.GetPayinProductId(),
			PayinProductCode: o.GetPayinProductCode(),
			PayinProductName: payProductName,
			ChannelLocked:    o.GetChannelLocked(),
			PaidAmount:       o.GetPaidAmount(),
			FeeMode:          o.GetFeeMode(),
			FeeRateBps:       o.GetFeeRateBps(),
			FeeFixedAmount:   o.GetFeeFixedAmount(),
			FeeAmount:        o.GetFeeAmount(),
			NetAmount:        o.GetNetAmount(),
			ReturnUrl:        o.GetReturnUrl(),
			NotifyUrl:        o.GetNotifyUrl(),
			UpstreamTradeNo:  o.GetUpstreamTradeNo(),
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
