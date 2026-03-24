package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	if !errors.Is(err, sql.ErrNoRows) {
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

	rec := &store.OrderRecord{
		OrderNo:         orderNo,
		MerchantId:      in.GetMerchantId(),
		MerchantOrderNo: in.GetMerchantOrderNo(),
		Amount:          in.GetAmount(),
		Currency:        in.GetCurrency(),
		Status:          store.OrderStatusPending,
		ChannelId:       in.GetChannelId(),
		PayinProductId:    in.GetPayoutProductId(),
		PayinProductCode:  productCode,
		PaidAmount:      0,
		FeeMode:         in.GetFeeMode(),
		FeeRateBps:      in.GetFeeRateBps(),
		FeeFixedAmount:  in.GetFeeFixedAmount(),
		FeeAmount:       in.GetFeeAmount(),
		NetAmount:       in.GetNetAmount(),
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
		rec *store.OrderRecord
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
		if errors.Is(err, sql.ErrNoRows) {
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
	records, err := l.svcCtx.PayoutOrders.ListByMerchant(l.ctx, merchantId, in.GetKeyword(), in.GetStatus(), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.ListOrdersResp{Orders: out}, nil
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
	records, err := l.svcCtx.PayoutOrders.AdminList(l.ctx, strings.TrimSpace(in.GetMerchantId()), strings.TrimSpace(in.GetKeyword()), st, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin list payout orders failed")
	}
	out := make([]*orderpb.OrderInfo, 0, len(records))
	for i := range records {
		out = append(out, toOrderInfo(&records[i]))
	}
	return &orderpb.AdminListOrdersResp{Orders: out}, nil
}
