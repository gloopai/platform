package logic

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gloopai/pay/channeldriver"
	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
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

	ok, err := l.svcCtx.MerchantPayinProducts.MerchantHasPayinProductCode(l.ctx, rec.MerchantId, code)
	if err != nil {
		return nil, status.Error(codes.Internal, "check merchant pay products failed")
	}
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "payin_product not enabled for this merchant")
	}

	chID, payPID, err := l.svcCtx.Channels.Route(l.ctx, code, rec.Amount)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
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

func (l *PrepareTerminalPayLogic) prepareLockedTerminal(rec *store.OrderRecord, orderNo, code string) (*orderpb.PrepareTerminalPayResp, error) {
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
func (l *PrepareTerminalPayLogic) terminalPaySurface(chID, payPID int64, code, orderNo string, rec *store.OrderRecord) (*orderpb.PrepareTerminalPayResp, error) {
	chRow, err := l.svcCtx.Channels.AdminGetByID(l.ctx, chID)
	if err != nil {
		return nil, status.Error(codes.Internal, "load channel failed")
	}
	dk := strings.TrimSpace(chRow.PayinType)
	gw := chRow.GatewayUrl
	mer := chRow.UpstreamMerchantNo
	sig := chRow.SignSecret
	rsa := chRow.RsaPrivateKey
	if uc := strings.TrimSpace(chRow.UpstreamConfig); uc != "" {
		jg, jm, js, jr := channeldriver.ConfigFieldsFromUpstreamJSON(uc)
		if jg != "" {
			gw = jg
		}
		if jm != "" {
			mer = jm
		}
		if js != "" {
			sig = js
		}
		if jr != "" {
			rsa = jr
		}
	}
	notifyBase := strings.TrimSpace(l.svcCtx.Config.Upstream.CheckoutNotifyBaseURL)
	if notifyBase != "" && dk != "" {
		if drv, derr := l.svcCtx.ChannelDrivers.Payin(dk); derr == nil {
			cfg := channeldriver.ConfigFromDriverKey(
				chRow.ID, dk, gw, mer, sig, rsa,
				chRow.SupportsPayin, chRow.SupportsPayout,
			)
			notifyURL := fmt.Sprintf("%s/v1/callback/upstream/payin?channel_id=%d&order_no=%s",
				strings.TrimRight(notifyBase, "/"), chID, url.QueryEscape(orderNo))
			resp, cerr := drv.CreatePayment(l.ctx, cfg, &channeldriver.CreatePaymentReq{
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
					PayMode:          "upstream",
				}, nil
			}
			if cerr != nil {
				l.Errorf("upstream CreatePayment channel_id=%d driver=%s err=%v", chID, dk, cerr)
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
