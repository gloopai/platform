package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/common/channelconfig"
	"github.com/gloopai/pay/common/model"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *orderpb.CreateOrderReq) (*orderpb.CreateOrderResp, error) {
	if in.GetMerchantId() == "" || in.GetMerchantOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and merchant_order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	existing, err := l.svcCtx.PayOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	if err == nil {
		return &orderpb.CreateOrderResp{
			Order:   toOrderInfo(existing),
			Existed: true,
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "query existing order failed")
	}

	lockKey := "idempotent:order:create:" + in.GetMerchantId() + ":" + in.GetMerchantOrderNo()
	ok, err := l.svcCtx.Redis.SetNX(l.ctx, lockKey, "1", 10*time.Minute).Result()
	if err != nil {
		return nil, status.Error(codes.Internal, "redis error")
	}
	if !ok {
		existing, err := l.svcCtx.PayOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
		if err == nil {
			return &orderpb.CreateOrderResp{
				Order:   toOrderInfo(existing),
				Existed: true,
			}, nil
		}
		return nil, status.Error(codes.Aborted, "duplicate request")
	}

	orderNo, err := newOrderNo()
	if err != nil {
		_ = l.svcCtx.Redis.Del(l.ctx, lockKey).Err()
		return nil, status.Error(codes.Internal, "generate order_no failed")
	}
	payCode := strings.TrimSpace(in.GetPayinProductCode())
	if payCode == "" {
		payCode = strings.TrimSpace(in.GetPayinType())
	}
	rec := &model.OrderRecord{
		OrderNo:          orderNo,
		MerchantId:       in.GetMerchantId(),
		MerchantOrderNo:  in.GetMerchantOrderNo(),
		Amount:           in.GetAmount(),
		Currency:         in.GetCurrency(),
		Status:           store.OrderStatusPending,
		ChannelId:        in.GetChannelId(),
		PayinProductId:   in.GetPayinProductId(),
		PayinProductCode: payCode,
		ChannelLocked:    in.GetChannelLocked(),
		PaidAmount:       0,
		FeeMode:          in.GetFeeMode(),
		FeeRateBps:       in.GetFeeRateBps(),
		FeeFixedAmount:   in.GetFeeFixedAmount(),
		FeeAmount:        in.GetFeeAmount(),
		NetAmount:        in.GetNetAmount(),
		ReturnUrl:        in.GetReturnUrl(),
		NotifyUrl:        in.GetNotifyUrl(),
	}
	if rec.Currency == "" {
		rec.Currency = "CNY"
	}
	if rec.FeeMode < 1 || rec.FeeMode > 3 {
		rec.FeeMode = 1
	}
	if rec.FeeRateBps < 0 {
		rec.FeeRateBps = 0
	}
	if rec.FeeFixedAmount < 0 {
		rec.FeeFixedAmount = 0
	}
	if rec.FeeAmount < 0 {
		rec.FeeAmount = 0
	}
	if rec.NetAmount < 0 {
		rec.NetAmount = 0
	}

	if err := l.svcCtx.PayOrders.Insert(l.ctx, rec); err != nil {
		_ = l.svcCtx.Redis.Del(l.ctx, lockKey).Err()
		return nil, status.Error(codes.Internal, "insert order failed")
	}

	_ = l.svcCtx.Redis.Expire(l.ctx, lockKey, 10*time.Minute).Err()
	created, err := l.svcCtx.PayOrders.FindByOrderNo(l.ctx, orderNo)
	if err != nil {
		return nil, status.Error(codes.Internal, "load created order failed")
	}

	return &orderpb.CreateOrderResp{
		Order:   toOrderInfo(created),
		Existed: false,
	}, nil
}

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *orderpb.GetOrderReq) (*orderpb.GetOrderResp, error) {
	var (
		rec *model.OrderRecord
		err error
	)
	switch {
	case in.GetOrderNo() != "":
		rec, err = l.svcCtx.PayOrders.FindByOrderNo(l.ctx, in.GetOrderNo())
	case in.GetMerchantId() != "" && in.GetMerchantOrderNo() != "":
		rec, err = l.svcCtx.PayOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	default:
		return nil, status.Error(codes.InvalidArgument, "order_no or (merchant_id and merchant_order_no) required")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "get order failed")
	}
	if in.GetMerchantId() != "" && rec.MerchantId != in.GetMerchantId() {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &orderpb.GetOrderResp{
		Order: toOrderInfo(rec),
	}, nil
}

type ListPayOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayOrdersLogic {
	return &ListPayOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPayOrdersLogic) ListPayOrders(in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	records, total, err := l.svcCtx.PayOrders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetOffset(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list orders failed")
	}

	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		rec := records[i]
		out = append(out, toOrderInfo(&rec))
	}
	return &orderpb.ListOrdersResp{Orders: out, Total: total}, nil
}

type TodaySummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTodaySummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TodaySummaryLogic {
	return &TodaySummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TodaySummaryLogic) TodaySummary(in *orderpb.TodaySummaryReq) (*orderpb.TodaySummaryResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	totalAmount, totalCount, successCount, err := l.svcCtx.PayOrders.TodaySummary(l.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "today summary failed")
	}

	return &orderpb.TodaySummaryResp{
		TotalAmount:  totalAmount,
		TotalCount:   totalCount,
		SuccessCount: successCount,
	}, nil
}

type MarkPaidLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkPaidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkPaidLogic {
	return &MarkPaidLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkPaidLogic) MarkPaid(in *orderpb.MarkPaidReq) (*orderpb.MarkPaidResp, error) {
	if in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	if in.GetPaidAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "paid_amount must be positive")
	}

	rec, err := l.svcCtx.PayOrders.FindByOrderNo(l.ctx, in.GetOrderNo())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "get order failed")
	}
	if rec.ChannelId != in.GetChannelId() {
		return nil, status.Error(codes.FailedPrecondition, "channel mismatch")
	}

	changed, err := l.svcCtx.PayOrders.MarkPaid(l.ctx, in.GetOrderNo(), in.GetPaidAmount(), in.GetChannelTradeNo())
	if err != nil {
		return nil, status.Error(codes.Internal, "mark paid failed")
	}
	return &orderpb.MarkPaidResp{Changed: changed}, nil
}

func newOrderNo() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "P" + time.Now().Format("20060102150405") + hex.EncodeToString(b[:8]), nil
}

func toOrderInfo(rec *model.OrderRecord) *orderpb.OrderInfo {
	return &orderpb.OrderInfo{
		OrderNo:          rec.OrderNo,
		MerchantId:       rec.MerchantId,
		MerchantOrderNo:  rec.MerchantOrderNo,
		Amount:           rec.Amount,
		Currency:         rec.Currency,
		Status:           rec.Status,
		ChannelId:        rec.ChannelId,
		PayinProductId:   rec.PayinProductId,
		PayinProductCode: rec.PayinProductCode,
		ChannelLocked:    rec.ChannelLocked,
		FeeMode:          rec.FeeMode,
		FeeRateBps:       rec.FeeRateBps,
		FeeFixedAmount:   rec.FeeFixedAmount,
		FeeAmount:        rec.FeeAmount,
		NetAmount:        rec.NetAmount,
		CreatedAt:        rec.CreatedAt.Unix(),
		UpdatedAt:        rec.UpdatedAt.Unix(),
		ReturnUrl:        rec.ReturnUrl,
		NotifyUrl:        rec.NotifyUrl,
		ChannelTradeNo:   rec.ChannelTradeNo,
		PaidAmount:       rec.PaidAmount,
		ChannelName:      rec.ChannelName,
	}
}

type PayoutOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayoutOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayoutOrderLogic {
	return &PayoutOrderLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *PayoutOrderLogic) CreatePayoutOrder(in *orderpb.CreatePayoutOrderReq) (*orderpb.CreateOrderResp, error) {
	if in.GetMerchantId() == "" || in.GetMerchantOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and merchant_order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	existing, err := l.svcCtx.PayoutOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	if err == nil {
		return &orderpb.CreateOrderResp{Order: toOrderInfo(existing), Existed: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "query existing payout order failed")
	}

	orderNo, err := newOrderNo()
	if err != nil {
		return nil, status.Error(codes.Internal, "generate order_no failed")
	}
	productCode := strings.TrimSpace(in.GetPayoutProductCode())
	if productCode == "" {
		return nil, status.Error(codes.InvalidArgument, "payout_product_code required")
	}

	rec := &model.OrderRecord{
		OrderNo:          orderNo,
		MerchantId:       in.GetMerchantId(),
		MerchantOrderNo:  in.GetMerchantOrderNo(),
		Amount:           in.GetAmount(),
		Currency:         in.GetCurrency(),
		Status:           store.OrderStatusPending,
		ChannelId:        in.GetChannelId(),
		PayinProductId:   in.GetPayoutProductId(),
		PayinProductCode: productCode,
		PaidAmount:       0,
		FeeMode:          in.GetFeeMode(),
		FeeRateBps:       in.GetFeeRateBps(),
		FeeFixedAmount:   in.GetFeeFixedAmount(),
		FeeAmount:        in.GetFeeAmount(),
		NetAmount:        in.GetNetAmount(),
		NotifyUrl:        in.GetNotifyUrl(),
	}
	if rec.Currency == "" {
		rec.Currency = "CNY"
	}
	if rec.FeeMode < 1 || rec.FeeMode > 3 {
		rec.FeeMode = 1
	}
	if rec.FeeRateBps < 0 {
		rec.FeeRateBps = 0
	}
	if rec.FeeFixedAmount < 0 {
		rec.FeeFixedAmount = 0
	}
	if rec.FeeAmount < 0 {
		rec.FeeAmount = 0
	}
	if rec.NetAmount < 0 {
		rec.NetAmount = 0
	}

	if err := l.svcCtx.PayoutOrders.Insert(l.ctx, rec); err != nil {
		var me *mysql.MySQLError
		dup := errors.As(err, &me) && me.Number == 1062
		if !dup {
			lowerErr := strings.ToLower(err.Error())
			dup = strings.Contains(lowerErr, "duplicate") || strings.Contains(lowerErr, "unique")
		}
		if dup {
			existed, ge := l.svcCtx.PayoutOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
			if ge == nil {
				return &orderpb.CreateOrderResp{Order: toOrderInfo(existed), Existed: true}, nil
			}
		}
		return nil, status.Error(codes.Internal, "insert payout order failed")
	}
	created, err := l.svcCtx.PayoutOrders.FindByOrderNo(l.ctx, orderNo)
	if err != nil {
		return nil, status.Error(codes.Internal, "load created payout order failed")
	}
	return &orderpb.CreateOrderResp{Order: toOrderInfo(created), Existed: false}, nil
}

func (l *PayoutOrderLogic) GetPayoutOrder(in *orderpb.GetOrderReq) (*orderpb.GetOrderResp, error) {
	var (
		rec *model.OrderRecord
		err error
	)
	switch {
	case in.GetOrderNo() != "":
		rec, err = l.svcCtx.PayoutOrders.FindByOrderNo(l.ctx, in.GetOrderNo())
	case in.GetMerchantId() != "" && in.GetMerchantOrderNo() != "":
		rec, err = l.svcCtx.PayoutOrders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	default:
		return nil, status.Error(codes.InvalidArgument, "order_no or (merchant_id and merchant_order_no) required")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payout order not found")
		}
		return nil, status.Error(codes.Internal, "get payout order failed")
	}
	if in.GetMerchantId() != "" && rec.MerchantId != in.GetMerchantId() {
		return nil, status.Error(codes.NotFound, "payout order not found")
	}
	return &orderpb.GetOrderResp{Order: toOrderInfo(rec)}, nil
}

func (l *PayoutOrderLogic) ListPayoutOrders(in *orderpb.ListOrdersReq) (*orderpb.ListOrdersResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	records, total, err := l.svcCtx.PayoutOrders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetOffset(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.ListOrdersResp{Orders: out, Total: total}, nil
}

func (l *PayoutOrderLogic) AdminListPayoutOrders(in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	limit := in.GetLimit()
	st := int32(-1)
	if in.Status != nil {
		st = *in.Status
		if st < -1 || st > 3 {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
	}
	records, total, err := l.svcCtx.PayoutOrders.AdminList(l.ctx, strings.TrimSpace(in.GetMerchantId()), strings.TrimSpace(in.GetKeyword()), st, in.GetOffset(), limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.AdminListOrdersResp{Orders: out, Total: total}, nil
}

type PrepareTerminalPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPrepareTerminalPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PrepareTerminalPayLogic {
	return &PrepareTerminalPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PrepareTerminalPayLogic) PrepareTerminalPay(in *orderpb.PrepareTerminalPayReq) (*orderpb.PrepareTerminalPayResp, error) {
	orderNo := strings.TrimSpace(in.GetOrderNo())
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}

	rec, err := l.svcCtx.PayOrders.FindByOrderNo(l.ctx, orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "get order failed")
	}
	if rec.Status != store.OrderStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "order not payable")
	}

	code := strings.TrimSpace(in.GetPayinProductCode())
	if rec.ChannelLocked != 0 {
		return l.prepareLockedTerminal(rec, orderNo, code)
	}

	if code == "" {
		code = strings.TrimSpace(rec.PayinProductCode)
	}
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "payin_product_code required")
	}

	mh, err := l.svcCtx.ChannelRpc.MerchantHasPayinProductCode(l.ctx, &channelpb.MerchantHasPayinProductCodeReq{
		MerchantId:       rec.MerchantId,
		PayinProductCode: code,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "check merchant pay products failed")
	}
	if !mh.GetOk() {
		return nil, status.Error(codes.PermissionDenied, "payin_product not enabled for this merchant")
	}

	route, err := l.svcCtx.ChannelRpc.Route(l.ctx, &channelpb.RouteReq{Amount: rec.Amount, PayinType: code})
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	chID := route.GetChannelId()
	payPID := route.GetPayinProductId()
	if chID <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "no available channel")
	}

	if err := l.svcCtx.PayOrders.UpdatePendingPayRoute(l.ctx, orderNo, chID, payPID, code); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "order not pending or not found")
		}
		return nil, status.Error(codes.Internal, "update order failed")
	}

	return l.terminalPaySurface(chID, payPID, code, orderNo, rec)
}

func (l *PrepareTerminalPayLogic) prepareLockedTerminal(rec *model.OrderRecord, orderNo, code string) (*orderpb.PrepareTerminalPayResp, error) {
	if rec.ChannelId <= 0 || rec.PayinProductId <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "locked order missing channel/route")
	}
	if code == "" {
		code = strings.TrimSpace(rec.PayinProductCode)
	}
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "payin_product_code required")
	}
	if want := strings.TrimSpace(rec.PayinProductCode); want != "" && code != want {
		return nil, status.Error(codes.FailedPrecondition, "payin_product_code mismatch for locked order")
	}

	chID := rec.ChannelId
	payPID := rec.PayinProductId

	if err := l.svcCtx.PayOrders.UpdatePendingPayRoute(l.ctx, orderNo, chID, payPID, code); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "order not pending or not found")
		}
		return nil, status.Error(codes.Internal, "update order failed")
	}

	return l.terminalPaySurface(chID, payPID, code, orderNo, rec)
}

// terminalPaySurface uses channeldriver when channels.payin_type matches a registered driver and Upstream.CheckoutNotifyBaseURL is set; otherwise legacy gateway_url / mock surface.
func (l *PrepareTerminalPayLogic) terminalPaySurface(chID, payPID int64, code, orderNo string, rec *model.OrderRecord) (*orderpb.PrepareTerminalPayResp, error) {
	gch, err := l.svcCtx.ChannelRpc.GetChannel(l.ctx, &channelpb.GetChannelReq{ChannelId: chID})
	if err != nil {
		return nil, status.Error(codes.Internal, "load channel failed")
	}
	chRow := gch.GetChannel()
	if chRow == nil {
		return nil, status.Error(codes.Internal, "load channel failed")
	}
	dk := strings.TrimSpace(chRow.GetPayinType())
	gw := strings.TrimSpace(chRow.GetGatewayUrl())
	if uc := strings.TrimSpace(chRow.GetChannelConfig()); uc != "" {
		if jg := channelconfig.StringFromJSONObject(uc, "gateway_url"); jg != "" {
			gw = jg
		}
	}
	notifyBase := strings.TrimSpace(l.svcCtx.Config.Upstream.CheckoutNotifyBaseURL)
	if notifyBase != "" && dk != "" {
		raw, err := channelconfig.ChannelConfigJSONForBind(
			chRow.GetChannelConfig(),
			channelconfig.LegacyChannelFields{
				GatewayURL:        chRow.GetGatewayUrl(),
				ChannelMerchantNo: chRow.GetChannelMerchantNo(),
				SignSecret:        chRow.GetSignSecret(),
				RSAPrivateKey:     chRow.GetRsaPrivateKey(),
			},
			chRow.GetSupportsPayin(), chRow.GetSupportsPayout(),
		)
		if err != nil {
			l.Errorf("terminalPaySurface channel_config channel_id=%d err=%v", chID, err)
			return nil, status.Error(codes.InvalidArgument, "invalid channel_config")
		}
		if err := channelconfig.ValidateChannelConfigJSON(raw); err != nil {
			l.Errorf("terminalPaySurface channel_config channel_id=%d err=%v", chID, err)
			return nil, status.Error(codes.InvalidArgument, "invalid channel_config")
		}
		in := channeldriver.BindInput{
			ChannelID:         chRow.GetId(),
			DriverKey:         dk,
			ChannelConfigJSON: raw,
		}
		payCh, oerr := l.svcCtx.ChannelDrivers.OpenPayin(in)
		if oerr == nil {
			notifyURL := fmt.Sprintf("%s/v1/callback/upstream/payin?channel_id=%d&order_no=%s",
				strings.TrimRight(notifyBase, "/"), chID, url.QueryEscape(orderNo))
			resp, cerr := payCh.CreatePayment(l.ctx, &channeldriver.CreatePaymentReq{
				MerchantOrderNo: orderNo,
				AmountMinor:     rec.Amount,
				PayerName:       "payin",
				PayerPhone:      "0",
				PayerEmail:      "payin@local",
				NotifyURL:       notifyURL,
			})
			if cerr == nil && resp != nil {
				payURL := strings.TrimSpace(resp.PayURL)
				if payURL == "" {
					payURL = notifyURL
				}
				return &orderpb.PrepareTerminalPayResp{
					ChannelId:        chID,
					PayinProductId:   payPID,
					PayinProductCode: code,
					PayUrl:           payURL,
					QrPayload:        payURL,
					PayMode:          "channel",
				}, nil
			}
			if cerr != nil {
				l.Errorf("channel CreatePayment channel_id=%d driver=%s err=%v", chID, dk, cerr)
			}
		}
	}

	payURL, qrPayload, payMode := buildPaySurface(orderNo, rec.Amount, gw)
	return &orderpb.PrepareTerminalPayResp{
		ChannelId:        chID,
		PayinProductId:   payPID,
		PayinProductCode: code,
		PayUrl:           payURL,
		QrPayload:        qrPayload,
		PayMode:          payMode,
	}, nil
}

func buildPaySurface(orderNo string, amount int64, gatewayURL string) (payURL, qrPayload, payMode string) {
	gatewayURL = strings.TrimSpace(gatewayURL)
	if gatewayURL == "" {
		payload := fmt.Sprintf("mock://pay?order_no=%s&amount=%d", url.QueryEscape(orderNo), amount)
		return payload, payload, "mock"
	}
	u, err := url.Parse(gatewayURL)
	if err != nil {
		payload := fmt.Sprintf("mock://pay?order_no=%s&amount=%d", url.QueryEscape(orderNo), amount)
		return payload, payload, "mock"
	}
	q := u.Query()
	q.Set("order_no", orderNo)
	u.RawQuery = q.Encode()
	s := u.String()
	return s, s, "qr"
}

type AdminTodayOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminTodayOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminTodayOverviewLogic {
	return &AdminTodayOverviewLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminTodayOverviewLogic) AdminTodayOverview(*orderpb.AdminTodayOverviewReq) (*orderpb.AdminTodayOverviewResp, error) {
	tot, prods, chs, err := l.svcCtx.OrderStats.TodayOverview(l.ctx)
	if err != nil {
		return nil, err
	}
	var enabledCh, fused int64
	if rs, rerr := l.svcCtx.ChannelRpc.GetRoutingSummary(l.ctx, &channelpb.GetRoutingSummaryReq{}); rerr == nil && rs != nil {
		enabledCh = rs.GetEnabledChannels()
		fused = rs.GetFusedChannels()
	}

	totals := &orderpb.AdminStatsTotals{
		OrderCount:             tot.OrderCount,
		PaidAmount:             tot.PaidAmount,
		PaidCount:              tot.PaidCount,
		FailedCount:            tot.FailedCount,
		PendingCount:           tot.PendingCount,
		ClosedCount:            tot.ClosedCount,
		ConversionRatePct:      store.RateConversion(tot.PaidCount, tot.OrderCount),
		TerminalSuccessRatePct: store.RateTerminalSuccess(tot.PaidCount, tot.FailedCount),
	}

	outProd := make([]*orderpb.AdminStatsProductRow, 0, len(prods))
	for _, p := range prods {
		outProd = append(outProd, &orderpb.AdminStatsProductRow{
			ProductCode:            p.ProductCode,
			ProductName:            p.ProductName,
			OrderCount:             p.OrderCount,
			PaidAmount:             p.PaidAmount,
			PaidCount:              p.PaidCount,
			FailedCount:            p.FailedCount,
			ConversionRatePct:      store.RateConversion(p.PaidCount, p.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(p.PaidCount, p.FailedCount),
		})
	}

	outCh := make([]*orderpb.AdminStatsChannelRow, 0, len(chs))
	for _, c := range chs {
		outCh = append(outCh, &orderpb.AdminStatsChannelRow{
			ChannelId:              c.ChannelID,
			ChannelName:            c.ChannelName,
			OrderCount:             c.OrderCount,
			PaidAmount:             c.PaidAmount,
			PaidCount:              c.PaidCount,
			FailedCount:            c.FailedCount,
			ConversionRatePct:      store.RateConversion(c.PaidCount, c.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(c.PaidCount, c.FailedCount),
		})
	}

	return &orderpb.AdminTodayOverviewResp{
		Range:           "today",
		Totals:          totals,
		ByPayinProduct:  outProd,
		ByChannel:       outCh,
		EnabledChannels: enabledCh,
		FusedChannels:   fused,
	}, nil
}

type ListMerchantNotifyLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantNotifyLogsLogic {
	return &ListMerchantNotifyLogsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMerchantNotifyLogsLogic) ListMerchantNotifyLogs(in *orderpb.ListMerchantNotifyLogsReq) (*orderpb.ListMerchantNotifyLogsResp, error) {
	rows, err := l.svcCtx.NotifyLogs.ListByOrder(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*orderpb.MerchantNotifyLogItem, 0, len(rows))
	for _, x := range rows {
		out = append(out, &orderpb.MerchantNotifyLogItem{
			Id:           x.Id,
			NotifyUrl:    x.NotifyUrl,
			Attempt:      x.Attempt,
			HttpStatus:   x.HttpStatus,
			ResponseBody: x.ResponseBody,
			ErrorMsg:     x.ErrorMsg,
			CreatedAt:    x.CreatedAt.Unix(),
		})
	}
	return &orderpb.ListMerchantNotifyLogsResp{Logs: out}, nil
}

type AdminListPayOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPayOrdersLogic {
	return &AdminListPayOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminListPayOrdersLogic) AdminListPayOrders(in *orderpb.AdminListOrdersReq) (*orderpb.AdminListOrdersResp, error) {
	limit := in.GetLimit()
	st := int32(-1)
	if in.Status != nil {
		st = *in.Status
		if st < -1 || st > 3 {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
	}

	records, total, err := l.svcCtx.PayOrders.AdminList(l.ctx, strings.TrimSpace(in.GetMerchantId()), strings.TrimSpace(in.GetKeyword()), st, in.GetOffset(), limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin list orders failed")
	}

	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.AdminListOrdersResp{Orders: out, Total: total}, nil
}

type AdminDayOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDayOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDayOverviewLogic {
	return &AdminDayOverviewLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminDayOverviewLogic) AdminDayOverview(in *orderpb.AdminDayOverviewReq) (*orderpb.AdminDayOverviewResp, error) {
	ds := strings.TrimSpace(in.GetDate())
	merchantID := strings.TrimSpace(in.GetMerchantId())
	if ds == "" {
		return nil, status.Error(codes.InvalidArgument, "date is required (YYYY-MM-DD)")
	}
	day, err := time.ParseInLocation("2006-01-02", ds, time.Local)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	tot, prods, chs, err := l.svcCtx.OrderStats.DayOverview(l.ctx, day, merchantID)
	if err != nil {
		return nil, err
	}

	totals := &orderpb.AdminStatsTotals{
		OrderCount:             tot.OrderCount,
		PaidAmount:             tot.PaidAmount,
		PaidCount:              tot.PaidCount,
		FailedCount:            tot.FailedCount,
		PendingCount:           tot.PendingCount,
		ClosedCount:            tot.ClosedCount,
		ConversionRatePct:      store.RateConversion(tot.PaidCount, tot.OrderCount),
		TerminalSuccessRatePct: store.RateTerminalSuccess(tot.PaidCount, tot.FailedCount),
	}

	outProd := make([]*orderpb.AdminStatsProductRow, 0, len(prods))
	for _, p := range prods {
		outProd = append(outProd, &orderpb.AdminStatsProductRow{
			ProductCode:            p.ProductCode,
			ProductName:            p.ProductName,
			OrderCount:             p.OrderCount,
			PaidAmount:             p.PaidAmount,
			PaidCount:              p.PaidCount,
			FailedCount:            p.FailedCount,
			ConversionRatePct:      store.RateConversion(p.PaidCount, p.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(p.PaidCount, p.FailedCount),
		})
	}

	outCh := make([]*orderpb.AdminStatsChannelRow, 0, len(chs))
	for _, c := range chs {
		outCh = append(outCh, &orderpb.AdminStatsChannelRow{
			ChannelId:              c.ChannelID,
			ChannelName:            c.ChannelName,
			OrderCount:             c.OrderCount,
			PaidAmount:             c.PaidAmount,
			PaidCount:              c.PaidCount,
			FailedCount:            c.FailedCount,
			ConversionRatePct:      store.RateConversion(c.PaidCount, c.OrderCount),
			TerminalSuccessRatePct: store.RateTerminalSuccess(c.PaidCount, c.FailedCount),
		})
	}

	return &orderpb.AdminDayOverviewResp{
		Date:           ds,
		Totals:         totals,
		ByPayinProduct: outProd,
		ByChannel:      outCh,
	}, nil
}
