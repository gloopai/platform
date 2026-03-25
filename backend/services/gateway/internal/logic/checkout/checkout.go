package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/common/grpcclient/settleclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/requestx"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Checkout 开放收银与上游异步通知（API / 终端 / 回调）。
type Checkout struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckout(ctx context.Context, svcCtx *svc.ServiceContext) *Checkout {
	return &Checkout{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *Checkout) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	merchantID := strings.TrimSpace(req.MerchantId)
	payinType := strings.TrimSpace(req.PayinType)
	channelID := req.ChannelId

	var (
		route            *channelclient.RouteResp
		payinProductCode string
		channelLocked    int32
		cid              int64
		ppid             int64
	)

	switch {
	case channelID > 0:
		rl, err := c.svcCtx.ChannelRpc.ResolveLockedChannelForMerchant(c.ctx, &channelpb.ResolveLockedChannelForMerchantReq{
			MerchantId: merchantID, ChannelId: channelID, Amount: req.Amount,
		})
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		channelLocked = 1
		cid = channelID
		ppid = rl.GetPayinProductId()
		payinProductCode = rl.GetPayinProductCode()

	case payinType != "":
		has, err := c.svcCtx.ChannelRpc.MerchantHasPayinProductCode(c.ctx, &channelpb.MerchantHasPayinProductCodeReq{
			MerchantId: merchantID, PayinProductCode: payinType,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !has.GetOk() {
			return nil, status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
		}
		route, err = c.svcCtx.ChannelRpc.Route(c.ctx, &channelclient.RouteReq{
			Amount:    req.Amount,
			PayinType: payinType,
		})
		if err != nil {
			return nil, err
		}
		cid = route.GetChannelId()
		ppid = route.GetPayinProductId()
		payinProductCode = payinType
		channelLocked = 0

	default:
		cid, ppid = 0, 0
		payinProductCode = ""
		channelLocked = 0
	}

	feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount := int64(1), int64(0), int64(0), int64(0), req.Amount
	if info, ge := c.svcCtx.MerchantRpc.GetMerchant(c.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantID}); ge == nil {
		feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount = calcPayinFeeSnapshot(info.GetMerchant(), ppid, req.Amount)
	}

	r, err := c.svcCtx.OrderRpc.CreateOrder(c.ctx, &orderclient.CreateOrderReq{
		MerchantId:       req.MerchantId,
		MerchantOrderNo:  req.MerchantOrderNo,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Subject:          req.Subject,
		ReturnUrl:        req.ReturnUrl,
		NotifyUrl:        req.NotifyUrl,
		PayinType:        payinType,
		ChannelId:        cid,
		PayinProductId:   ppid,
		PayinProductCode: payinProductCode,
		ChannelLocked:    channelLocked,
		FeeMode:          feeMode,
		FeeRateBps:       feeRateBps,
		FeeFixedAmount:   feeFixedAmount,
		FeeAmount:        feeAmount,
		NetAmount:        netAmount,
	})
	if err != nil {
		return nil, err
	}

	orderInfo := r.GetOrder()
	base := strings.TrimSpace(c.svcCtx.Config.CheckoutBaseUrl)
	if base == "" {
		base = "http://127.0.0.1:5174/"
	}
	base = strings.TrimRight(base, "/")
	checkoutURL := base + "/?order_no=" + orderInfo.GetOrderNo()

	return &types.CreateOrderResp{
		OrderNo:          orderInfo.GetOrderNo(),
		Status:           orderInfo.GetStatus(),
		ChannelId:        orderInfo.GetChannelId(),
		PayinProductId:   orderInfo.GetPayinProductId(),
		PayinProductCode: orderInfo.GetPayinProductCode(),
		CheckoutUrl:      checkoutURL,
		ChannelLocked:    orderInfo.GetChannelLocked(),
	}, nil
}

func (c *Checkout) QueryOrder(req *types.QueryOrderReq) (resp *types.QueryOrderResp, err error) {
	r, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		MerchantId:      req.MerchantId,
		OrderNo:         req.OrderNo,
		MerchantOrderNo: req.MerchantOrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	return &types.QueryOrderResp{
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
	}, nil
}

func (c *Checkout) CreatePayoutOrder(req *types.CreatePayinOrderReq) (*types.CreateOrderResp, error) {
	reqID := requestx.FromContext(c.ctx)
	merchantID := strings.TrimSpace(req.MerchantId)
	if merchantID == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	payoutCode := strings.TrimSpace(req.PayoutProductCode)
	if payoutCode == "" {
		payoutCode = strings.TrimSpace(req.PayinType)
	}
	if payoutCode == "" {
		return nil, status.Error(codes.InvalidArgument, "payout_product_code required")
	}
	payoutCodeNorm := payoutCode
	pl, err := c.svcCtx.ChannelRpc.AdminListPayoutProducts(c.ctx, &channelpb.AdminListPayoutProductsReq{})
	if err != nil {
		return nil, status.Error(codes.Internal, "list payout products failed")
	}
	var resolvedPayoutProductID int64
	for _, p := range pl.GetProducts() {
		if p == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(p.GetCode()), payoutCodeNorm) {
			resolvedPayoutProductID = p.GetId()
			break
		}
	}
	if resolvedPayoutProductID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "unknown payout_product_code")
	}
	info, ge := c.svcCtx.MerchantRpc.GetMerchant(c.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantID})
	if ge != nil {
		return nil, ge
	}
	m := info.GetMerchant()
	if m == nil {
		return nil, status.Error(codes.Internal, "merchant load failed")
	}
	grantOk := false
	for _, g := range m.GetPayoutGrants() {
		if g != nil && g.GetPayoutProductId() == resolvedPayoutProductID {
			grantOk = true
			break
		}
	}
	if !grantOk {
		for _, id := range m.GetPayoutProductIds() {
			if id == resolvedPayoutProductID {
				grantOk = true
				break
			}
		}
	}
	if !grantOk {
		return nil, status.Error(codes.PermissionDenied, "payout_product_code not enabled for merchant")
	}
	feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount := calcPayoutFeeSnapshot(m, resolvedPayoutProductID, req.Amount)
	payoutProductID := resolvedPayoutProductID
	r, err := c.svcCtx.OrderRpc.CreatePayoutOrder(c.ctx, &orderclient.CreatePayoutOrderReq{
		MerchantId:        merchantID,
		MerchantOrderNo:   req.MerchantOrderNo,
		Amount:            req.Amount,
		Currency:          req.Currency,
		NotifyUrl:         req.NotifyUrl,
		ChannelId:         req.ChannelId,
		PayoutProductId:   payoutProductID,
		PayoutProductCode: payoutCode,
		FeeMode:           feeMode,
		FeeRateBps:        feeRateBps,
		FeeFixedAmount:    feeFixedAmount,
		FeeAmount:         feeAmount,
		NetAmount:         netAmount,
	})
	if err != nil {
		c.Errorf("request_id=%s action=create_payout_order stage=create_order merchant_id=%s merchant_order_no=%s err=%v", reqID, merchantID, req.MerchantOrderNo, err)
		return nil, err
	}
	o := r.GetOrder()
	if r.GetExisted() && o.GetStatus() == 0 {
		c.Infof("request_id=%s action=create_payout_order stage=idempotent_pending merchant_id=%s merchant_order_no=%s order_no=%s", reqID, merchantID, req.MerchantOrderNo, o.GetOrderNo())
		return nil, status.Error(codes.FailedPrecondition, "payout order already exists and pending; use new merchant_order_no")
	}
	// 轻量幂等：仅在首次创建代付单时扣款；同 merchant_order_no 重试不重复扣款。
	if !r.GetExisted() {
		totalDebit := req.Amount + feeAmount
		if _, derr := c.svcCtx.SettleRpc.DebitPayout(c.ctx, &settleclient.DebitPayoutReq{
			MerchantId: merchantID,
			OrderNo:    o.GetOrderNo(),
			Amount:     totalDebit,
			Reason:     "PAYOUT_ORDER_DEBIT",
		}); derr != nil {
			// Avoid leaving a pending payout order when any debit step fails.
			changed, markErr := c.svcCtx.ServiceHub.MarkPayoutFailed(c.ctx, o.GetOrderNo())
			if markErr != nil {
				c.Errorf("request_id=%s action=create_payout_order stage=mark_failed_error merchant_id=%s order_no=%s err=%v", reqID, merchantID, o.GetOrderNo(), markErr)
			} else {
				c.Infof("request_id=%s action=create_payout_order stage=mark_failed merchant_id=%s order_no=%s changed=%v", reqID, merchantID, o.GetOrderNo(), changed)
			}
			c.Errorf("request_id=%s action=create_payout_order stage=debit_failed merchant_id=%s merchant_order_no=%s order_no=%s err=%v", reqID, merchantID, req.MerchantOrderNo, o.GetOrderNo(), derr)
			return nil, derr
		}
		c.Infof("request_id=%s action=create_payout_order stage=debit_ok merchant_id=%s merchant_order_no=%s order_no=%s debit_amount=%d", reqID, merchantID, req.MerchantOrderNo, o.GetOrderNo(), totalDebit)
	}
	c.Infof("request_id=%s action=create_payout_order stage=done merchant_id=%s merchant_order_no=%s order_no=%s status=%d", reqID, merchantID, req.MerchantOrderNo, o.GetOrderNo(), o.GetStatus())
	return &types.CreateOrderResp{
		OrderNo:          o.GetOrderNo(),
		Status:           o.GetStatus(),
		ChannelId:        o.GetChannelId(),
		PayinProductId:   o.GetPayinProductId(),
		PayinProductCode: o.GetPayinProductCode(),
		CheckoutUrl:      "",
		ChannelLocked:    0,
	}, nil
}

func calcPayoutFeeSnapshot(m *merchantpb.MerchantInfo, payoutProductID int64, amount int64) (feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount int64) {
	feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount = 1, 0, 0, 0, amount
	if m == nil || amount <= 0 || payoutProductID <= 0 {
		return
	}
	feeRateBps = m.GetDefaultPayoutRateBps()
	var matched *merchantpb.MerchantPayoutGrant
	for _, g := range m.GetPayoutGrants() {
		if g == nil || g.GetPayoutProductId() != payoutProductID {
			continue
		}
		matched = g
		break
	}
	if matched != nil {
		mode := matched.GetFeeMode()
		if mode < 1 || mode > 3 {
			mode = 1
		}
		feeMode = mode
		if matched.MerchantRateBps != nil {
			feeRateBps = matched.GetMerchantRateBps()
		}
		feeFixedAmount = matched.GetFeeFixedAmount()
	}
	if feeRateBps < 0 {
		feeRateBps = 0
	}
	if feeFixedAmount < 0 {
		feeFixedAmount = 0
	}
	switch feeMode {
	case 2:
		feeAmount = feeFixedAmount
	case 3:
		feeAmount = feeFixedAmount + amount*feeRateBps/10000
	default:
		feeAmount = amount * feeRateBps / 10000
	}
	if feeAmount < 0 {
		feeAmount = 0
	}
	netAmount = amount - feeAmount
	if netAmount < 0 {
		netAmount = 0
	}
	return
}

func (c *Checkout) QueryPayoutOrder(req *types.QueryOrderReq) (*types.QueryOrderResp, error) {
	r, err := c.svcCtx.OrderRpc.GetPayoutOrder(c.ctx, &orderclient.GetOrderReq{
		MerchantId:      req.MerchantId,
		OrderNo:         req.OrderNo,
		MerchantOrderNo: req.MerchantOrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()
	return &types.QueryOrderResp{
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
			PaidAmount:       o.GetPaidAmount(),
			FeeMode:          o.GetFeeMode(),
			FeeRateBps:       o.GetFeeRateBps(),
			FeeFixedAmount:   o.GetFeeFixedAmount(),
			FeeAmount:        o.GetFeeAmount(),
			NetAmount:        o.GetNetAmount(),
			NotifyUrl:        o.GetNotifyUrl(),
			UpstreamTradeNo:  o.GetUpstreamTradeNo(),
		},
	}, nil
}

func (c *Checkout) QueryMerchantBalance(req *types.MerchantBalanceQueryReq) (*types.MerchantBalanceQueryResp, error) {
	merchantID := strings.TrimSpace(req.MerchantId)
	if merchantID == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	auth, err := c.svcCtx.MerchantRpc.GetAuthInfo(c.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantID})
	if err != nil {
		return nil, err
	}
	return &types.MerchantBalanceQueryResp{
		MerchantId:    merchantID,
		PayinBalance:  auth.GetPayinBalance(),
		AvailableBalance: auth.GetAvailableBalance(),
	}, nil
}

func (c *Checkout) TerminalOrder(req *types.TerminalOrderReq) (resp *types.TerminalOrderResp, err error) {
	r, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return nil, err
	}
	o := r.GetOrder()

	var items []types.PayinProductItem
	if o.GetChannelLocked() != 0 {
		code := o.GetPayinProductCode()
		name := code
		if code != "" {
			if dn, err := c.svcCtx.ChannelRpc.GetPayinProductDisplayName(c.ctx, &channelpb.GetPayinProductDisplayNameReq{Code: code}); err == nil && dn.GetName() != "" {
				name = dn.GetName()
			}
		}
		if code != "" {
			items = []types.PayinProductItem{{Code: code, Name: name}}
		}
	} else {
		lr, err := c.svcCtx.ChannelRpc.ListTerminalPayinProducts(c.ctx, &channelpb.ListTerminalPayinProductsReq{
			MerchantId: o.GetMerchantId(), Amount: o.GetAmount(),
		})
		if err != nil {
			return nil, err
		}
		opts := lr.GetProducts()
		items = make([]types.PayinProductItem, 0, len(opts))
		for _, p := range opts {
			items = append(items, types.PayinProductItem{Code: p.GetCode(), Name: p.GetName()})
		}
	}

	return &types.TerminalOrderResp{
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
		PayinProducts: items,
	}, nil
}

func (c *Checkout) TerminalPay(req *types.TerminalPayReq) (*types.TerminalPayResp, error) {
	orderNo := strings.TrimSpace(req.OrderNo)
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}

	gr, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: orderNo})
	if err != nil {
		return nil, err
	}
	o := gr.GetOrder()
	code := strings.TrimSpace(req.PayinProductCode)

	if o.GetChannelLocked() == 0 {
		if code == "" {
			return nil, status.Error(codes.InvalidArgument, "payin_product_code required")
		}
		ok, err := c.svcCtx.ChannelRpc.MerchantHasPayinProductCode(c.ctx, &channelpb.MerchantHasPayinProductCodeReq{
			MerchantId: o.GetMerchantId(), PayinProductCode: code,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !ok.GetOk() {
			return nil, status.Error(codes.PermissionDenied, "payin_product_code not enabled for merchant")
		}
	}

	r, err := c.svcCtx.OrderRpc.PrepareTerminalPay(c.ctx, &orderclient.PrepareTerminalPayReq{
		OrderNo:          orderNo,
		PayinProductCode: code,
	})
	if err != nil {
		return nil, err
	}
	return &types.TerminalPayResp{
		ChannelId:        r.GetChannelId(),
		PayinProductId:   r.GetPayinProductId(),
		PayinProductCode: r.GetPayinProductCode(),
		PayUrl:           r.GetPayUrl(),
		QrPayload:        r.GetQrPayload(),
		PayMode:          r.GetPayMode(),
	}, nil
}

func (c *Checkout) UpstreamNotify(req *types.UpstreamNotifyReq) (resp *types.UpstreamNotifyResp, err error) {
	reqID := requestx.FromContext(c.ctx)
	if strings.TrimSpace(req.OrderNo) == "" || strings.TrimSpace(req.UpstreamTradeNo) == "" || req.ChannelId <= 0 || req.PaidAmount <= 0 {
		return notifyFail(NotifyCodeInvalidNotifyParams, "invalid notify params"), nil
	}

	signResp, err := c.svcCtx.ChannelRpc.GetSignSecret(c.ctx, &channelclient.GetSignSecretReq{ChannelId: req.ChannelId})
	if err != nil {
		return notifyFail(NotifyCodeChannelNotFound, "channel not found"), nil
	}

	expect := middleware.Md5Sign(map[string]string{
		"order_no":          req.OrderNo,
		"paid_amount":       strconv.FormatInt(req.PaidAmount, 10),
		"upstream_trade_no": req.UpstreamTradeNo,
		"channel_id":        strconv.FormatInt(req.ChannelId, 10),
		"sign":              req.Sign,
	}, signResp.GetSignSecret())
	if !strings.EqualFold(expect, req.Sign) {
		return notifyFail(NotifyCodeInvalidSign, "invalid sign"), nil
	}

	getResp, err := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return notifyFail(NotifyCodeOrderNotFound, "order not found"), nil
	}
	o := getResp.GetOrder()
	c.Infof("request_id=%s action=upstream_notify stage=received order_no=%s merchant_id=%s paid_amount=%d channel_id=%d", reqID, req.OrderNo, o.GetMerchantId(), req.PaidAmount, req.ChannelId)

	// 幂等与重放控制：
	// - 已支付订单仅接受与已落库支付快照完全一致的重复通知
	// - 非待支付（失败/关闭）不再接受支付成功通知
	if o.GetStatus() == 1 {
		if samePaidSnapshot(o, req) {
			return c.settlePaidOrderAndNotify(reqID, o, req, NotifyCodeIdempotentReplayAccepted, "idempotent replay accepted")
		}
		return notifyFail(NotifyCodeReplayPayloadMismatch, "replay payload mismatch"), nil
	}
	if o.GetStatus() != 0 {
		return notifyFail(NotifyCodeOrderNotPending, "order not pending"), nil
	}

	markResp, err := c.svcCtx.OrderRpc.MarkPaid(c.ctx, &orderclient.MarkPaidReq{
		OrderNo:         req.OrderNo,
		PaidAmount:      req.PaidAmount,
		UpstreamTradeNo: req.UpstreamTradeNo,
		ChannelId:       req.ChannelId,
	})
	if err != nil {
		c.Errorf("request_id=%s action=upstream_notify stage=mark_paid_failed order_no=%s err=%v", reqID, req.OrderNo, err)
		return notifyFail(NotifyCodeMarkPaidFailed, "mark paid failed"), nil
	}

	if !markResp.GetChanged() {
		// 并发场景：若另一条回调已先落库，允许同快照重放成功。
		latest, ge := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: req.OrderNo})
		if ge != nil {
			return notifyFail(NotifyCodeMarkPaidRace, "mark paid race"), nil
		}
		if samePaidSnapshot(latest.GetOrder(), req) {
			return c.settlePaidOrderAndNotify(reqID, latest.GetOrder(), req, NotifyCodeIdempotentRaceAccepted, "idempotent race accepted")
		}
		return notifyFail(NotifyCodeMarkPaidRaceMismatch, "mark paid race mismatch"), nil
	}

	return c.settlePaidOrderAndNotify(reqID, o, req, "", "")
}

func (c *Checkout) settlePaidOrderAndNotify(reqID string, o *orderclient.OrderInfo, req *types.UpstreamNotifyReq, okCode, okReason string) (*types.UpstreamNotifyResp, error) {
	creditAmount := req.PaidAmount
	if o.GetNetAmount() > 0 {
		creditAmount = o.GetNetAmount()
	}
	creditResp, err := c.svcCtx.SettleRpc.Credit(c.ctx, &settleclient.CreditReq{
		MerchantId: o.GetMerchantId(),
		OrderNo:    o.GetOrderNo(),
		Amount:     creditAmount,
		Reason:     "ORDER_PAID",
	})
	if err != nil {
		c.Errorf("request_id=%s action=upstream_notify stage=credit_rpc_error order_no=%s merchant_id=%s amount=%d err=%v",
			reqID, o.GetOrderNo(), o.GetMerchantId(), creditAmount, err)
		return notifyFail(NotifyCodeCreditFailed, "credit failed"), nil
	}
	if creditResp.GetChanged() {
		c.Infof("request_id=%s action=upstream_notify stage=credit_applied order_no=%s merchant_id=%s amount=%d balance=%d",
			reqID, o.GetOrderNo(), o.GetMerchantId(), creditAmount, creditResp.GetBalance())
	} else {
		c.Infof("request_id=%s action=upstream_notify stage=credit_idempotent order_no=%s merchant_id=%s amount=%d",
			reqID, o.GetOrderNo(), o.GetMerchantId(), creditAmount)
	}
	body, err := json.Marshal(map[string]any{
		"merchant_id": o.GetMerchantId(),
		"order_no":    o.GetOrderNo(),
		"attempt":     0,
	})
	if err != nil {
		c.Errorf("request_id=%s action=upstream_notify stage=notify_marshal_error order_no=%s err=%v", reqID, o.GetOrderNo(), err)
		return notifyFail(NotifyCodeNotifyMarshalFailed, "notify marshal failed"), nil
	}
	if err := c.svcCtx.NsqProducer.Publish(c.svcCtx.Config.Nsq.Topic, body); err != nil {
		c.Errorf("request_id=%s action=upstream_notify stage=notify_publish_error order_no=%s err=%v", reqID, o.GetOrderNo(), err)
		return notifyFail(NotifyCodeNotifyPublishFailed, "notify publish failed"), nil
	}
	return notifyOK(okCode, okReason), nil
}

func samePaidSnapshot(o *orderclient.OrderInfo, req *types.UpstreamNotifyReq) bool {
	return o != nil &&
		o.GetStatus() == 1 &&
		o.GetPaidAmount() == req.PaidAmount &&
		o.GetChannelId() == req.ChannelId &&
		strings.EqualFold(strings.TrimSpace(o.GetUpstreamTradeNo()), strings.TrimSpace(req.UpstreamTradeNo))
}

func notifyFail(code, reason string) *types.UpstreamNotifyResp {
	return &types.UpstreamNotifyResp{
		Ok:         false,
		ReasonCode: code,
		Reason:     reason,
	}
}

func notifyOK(code, reason string) *types.UpstreamNotifyResp {
	return &types.UpstreamNotifyResp{
		Ok:         true,
		ReasonCode: code,
		Reason:     reason,
	}
}

func calcPayinFeeSnapshot(m *merchantpb.MerchantInfo, payinProductID, amount int64) (feeMode, feeRateBps, feeFixedAmount, feeAmount, netAmount int64) {
	feeMode = 1
	feeRateBps = 0
	feeFixedAmount = 0
	feeAmount = 0
	netAmount = amount
	if m == nil || amount <= 0 {
		return
	}
	feeRateBps = m.GetDefaultPayinRateBps()
	for _, g := range m.GetPayinGrants() {
		if g == nil || g.GetPayinProductId() != payinProductID {
			continue
		}
		if g.MerchantRateBps != nil {
			feeRateBps = g.GetMerchantRateBps()
		}
		break
	}
	if feeRateBps < 0 {
		feeRateBps = 0
	}
	feeAmount = amount * feeRateBps / 10000
	if feeAmount < 0 {
		feeAmount = 0
	}
	if feeAmount > amount {
		feeAmount = amount
	}
	netAmount = amount - feeAmount
	return
}
