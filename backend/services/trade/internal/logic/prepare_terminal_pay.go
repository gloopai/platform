package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "get order failed")
	}
	if rec.Status != store.OrderStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "order not payable")
	}

	code := strings.TrimSpace(in.GetPayProductCode())
	if rec.ChannelLocked != 0 {
		return l.prepareLockedTerminal(rec, orderNo, code)
	}

	if code == "" {
		code = strings.TrimSpace(rec.PayProductCode)
	}
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "payin_product_code required")
	}

	ok, err := l.svcCtx.MerchantPayProducts.MerchantHasPayProductCode(l.ctx, rec.MerchantId, code)
	if err != nil {
		return nil, status.Error(codes.Internal, "check merchant pay products failed")
	}
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "pay_product not enabled for this merchant")
	}

	chID, payPID, err := l.svcCtx.Channels.Route(l.ctx, code, rec.Amount)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if chID <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "no available channel")
	}

	if err := l.svcCtx.PayOrders.UpdatePendingPayRoute(l.ctx, orderNo, chID, payPID, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.FailedPrecondition, "order not pending or not found")
		}
		return nil, status.Error(codes.Internal, "update order failed")
	}

	gw, _, err := l.svcCtx.Channels.GetGatewayURLAndPayType(l.ctx, chID)
	if err != nil {
		return nil, status.Error(codes.Internal, "load channel failed")
	}

	payURL, qrPayload, payMode := buildPaySurface(orderNo, rec.Amount, gw)

	return &orderpb.PrepareTerminalPayResp{
		ChannelId:      chID,
		PayProductId:   payPID,
		PayProductCode: code,
		PayUrl:         payURL,
		QrPayload:      qrPayload,
		PayMode:        payMode,
	}, nil
}

func (l *PrepareTerminalPayLogic) prepareLockedTerminal(rec *store.OrderRecord, orderNo, code string) (*orderpb.PrepareTerminalPayResp, error) {
	if rec.ChannelId <= 0 || rec.PayProductId <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "locked order missing channel/route")
	}
	if code == "" {
		code = strings.TrimSpace(rec.PayProductCode)
	}
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "payin_product_code required")
	}
	if want := strings.TrimSpace(rec.PayProductCode); want != "" && code != want {
		return nil, status.Error(codes.FailedPrecondition, "payin_product_code mismatch for locked order")
	}

	chID := rec.ChannelId
	payPID := rec.PayProductId

	if err := l.svcCtx.PayOrders.UpdatePendingPayRoute(l.ctx, orderNo, chID, payPID, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.FailedPrecondition, "order not pending or not found")
		}
		return nil, status.Error(codes.Internal, "update order failed")
	}

	gw, _, err := l.svcCtx.Channels.GetGatewayURLAndPayType(l.ctx, chID)
	if err != nil {
		return nil, status.Error(codes.Internal, "load channel failed")
	}

	payURL, qrPayload, payMode := buildPaySurface(orderNo, rec.Amount, gw)

	return &orderpb.PrepareTerminalPayResp{
		ChannelId:      chID,
		PayProductId:   payPID,
		PayProductCode: code,
		PayUrl:         payURL,
		QrPayload:      qrPayload,
		PayMode:        payMode,
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
