package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if !errors.Is(err, sql.ErrNoRows) {
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
		payCode = strings.TrimSpace(in.GetPayType())
	}
	rec := &store.OrderRecord{
		OrderNo:         orderNo,
		MerchantId:      in.GetMerchantId(),
		MerchantOrderNo: in.GetMerchantOrderNo(),
		Amount:          in.GetAmount(),
		Currency:        in.GetCurrency(),
		Status:          store.OrderStatusPending,
		ChannelId:       in.GetChannelId(),
		PayinProductId:    in.GetPayinProductId(),
		PayinProductCode:  payCode,
		ChannelLocked:   in.GetChannelLocked(),
		PaidAmount:      0,
		FeeMode:         in.GetFeeMode(),
		FeeRateBps:      in.GetFeeRateBps(),
		FeeFixedAmount:  in.GetFeeFixedAmount(),
		FeeAmount:       in.GetFeeAmount(),
		NetAmount:       in.GetNetAmount(),
		ReturnUrl:       in.GetReturnUrl(),
		NotifyUrl:       in.GetNotifyUrl(),
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
		rec *store.OrderRecord
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
		if errors.Is(err, sql.ErrNoRows) {
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

	records, err := l.svcCtx.PayOrders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list orders failed")
	}

	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		rec := records[i]
		out = append(out, toOrderInfo(&rec))
	}
	return &orderpb.ListOrdersResp{Orders: out}, nil
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

	changed, err := l.svcCtx.PayOrders.MarkPaid(l.ctx, in.GetOrderNo(), in.GetPaidAmount(), in.GetUpstreamTradeNo(), in.GetChannelId())
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

func toOrderInfo(rec *store.OrderRecord) *orderpb.OrderInfo {
	return &orderpb.OrderInfo{
		OrderNo:         rec.OrderNo,
		MerchantId:      rec.MerchantId,
		MerchantOrderNo: rec.MerchantOrderNo,
		Amount:          rec.Amount,
		Currency:        rec.Currency,
		Status:          rec.Status,
		ChannelId:       rec.ChannelId,
		PayinProductId:    rec.PayinProductId,
		PayinProductCode:  rec.PayinProductCode,
		ChannelLocked:   rec.ChannelLocked,
		FeeMode:         rec.FeeMode,
		FeeRateBps:      rec.FeeRateBps,
		FeeFixedAmount:  rec.FeeFixedAmount,
		FeeAmount:       rec.FeeAmount,
		NetAmount:       rec.NetAmount,
		CreatedAt:       rec.CreatedAt.Unix(),
		UpdatedAt:       rec.UpdatedAt.Unix(),
		ReturnUrl:       rec.ReturnUrl,
		NotifyUrl:       rec.NotifyUrl,
		UpstreamTradeNo: rec.UpstreamTradeNo,
		PaidAmount:      rec.PaidAmount,
	}
}
