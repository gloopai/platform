package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gloopai/pay/order/internal/store"
	"github.com/gloopai/pay/order/internal/svc"
	"github.com/gloopai/pay/order/order/order"

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

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	if in.GetMerchantId() == "" || in.GetMerchantOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and merchant_order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	existing, err := l.svcCtx.Orders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
	if err == nil {
		return &order.CreateOrderResp{
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
		existing, err := l.svcCtx.Orders.FindByMerchantOrderNo(l.ctx, in.GetMerchantId(), in.GetMerchantOrderNo())
		if err == nil {
			return &order.CreateOrderResp{
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
	rec := &store.OrderRecord{
		OrderNo:         orderNo,
		MerchantId:      in.GetMerchantId(),
		MerchantOrderNo: in.GetMerchantOrderNo(),
		Amount:          in.GetAmount(),
		Currency:        in.GetCurrency(),
		Status:          store.OrderStatusPending,
		ChannelId:       in.GetChannelId(),
		ReturnUrl:       in.GetReturnUrl(),
		NotifyUrl:       in.GetNotifyUrl(),
	}
	if rec.Currency == "" {
		rec.Currency = "CNY"
	}

	if err := l.svcCtx.Orders.Insert(l.ctx, rec); err != nil {
		_ = l.svcCtx.Redis.Del(l.ctx, lockKey).Err()
		return nil, status.Error(codes.Internal, "insert order failed")
	}

	_ = l.svcCtx.Redis.Expire(l.ctx, lockKey, 10*time.Minute).Err()
	created, err := l.svcCtx.Orders.FindByOrderNo(l.ctx, orderNo)
	if err != nil {
		return nil, status.Error(codes.Internal, "load created order failed")
	}

	return &order.CreateOrderResp{
		Order:   toOrderInfo(created),
		Existed: false,
	}, nil
}

func newOrderNo() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "P" + time.Now().UTC().Format("20060102150405") + hex.EncodeToString(b[:8]), nil
}

func toOrderInfo(rec *store.OrderRecord) *order.OrderInfo {
	return &order.OrderInfo{
		OrderNo:         rec.OrderNo,
		MerchantId:      rec.MerchantId,
		MerchantOrderNo: rec.MerchantOrderNo,
		Amount:          rec.Amount,
		Currency:        rec.Currency,
		Status:          rec.Status,
		ChannelId:       rec.ChannelId,
		CreatedAt:       rec.CreatedAt.Unix(),
		UpdatedAt:       rec.UpdatedAt.Unix(),
		ReturnUrl:       rec.ReturnUrl,
		NotifyUrl:       rec.NotifyUrl,
		UpstreamTradeNo: rec.UpstreamTradeNo,
	}
}
