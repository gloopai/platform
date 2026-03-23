package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/common/grpcclient/orderclient"
	"github.com/gloopai/pay/common/grpcclient/settleclient"
	channelpb "github.com/gloopai/pay/common/pb/channel"
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
	payType := strings.TrimSpace(req.PayType)
	channelID := req.ChannelId

	var (
		route          *channelclient.RouteResp
		payProductCode string
		channelLocked  int32
		cid            int64
		ppid           int64
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
		ppid = rl.GetPayProductId()
		payProductCode = rl.GetPayProductCode()

	case payType != "":
		has, err := c.svcCtx.ChannelRpc.MerchantHasPayProductCode(c.ctx, &channelpb.MerchantHasPayProductCodeReq{
			MerchantId: merchantID, PayProductCode: payType,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !has.GetOk() {
			return nil, status.Error(codes.PermissionDenied, "pay_type not enabled for this merchant")
		}
		route, err = c.svcCtx.ChannelRpc.Route(c.ctx, &channelclient.RouteReq{
			Amount:  req.Amount,
			PayType: payType,
		})
		if err != nil {
			return nil, err
		}
		cid = route.GetChannelId()
		ppid = route.GetPayProductId()
		payProductCode = payType
		channelLocked = 0

	default:
		cid, ppid = 0, 0
		payProductCode = ""
		channelLocked = 0
	}

	r, err := c.svcCtx.OrderRpc.CreateOrder(c.ctx, &orderclient.CreateOrderReq{
		MerchantId:      req.MerchantId,
		MerchantOrderNo: req.MerchantOrderNo,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Subject:         req.Subject,
		ReturnUrl:       req.ReturnUrl,
		NotifyUrl:       req.NotifyUrl,
		PayType:         payType,
		ChannelId:       cid,
		PayProductId:    ppid,
		PayProductCode:  payProductCode,
		ChannelLocked:   channelLocked,
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
		OrderNo:        orderInfo.GetOrderNo(),
		Status:         orderInfo.GetStatus(),
		ChannelId:      orderInfo.GetChannelId(),
		PayProductId:   orderInfo.GetPayProductId(),
		PayProductCode: orderInfo.GetPayProductCode(),
		CheckoutUrl:    checkoutURL,
		ChannelLocked:  orderInfo.GetChannelLocked(),
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

	var items []types.PayProductItem
	if o.GetChannelLocked() != 0 {
		code := o.GetPayProductCode()
		name := code
		if code != "" {
			if dn, err := c.svcCtx.ChannelRpc.GetPayProductDisplayName(c.ctx, &channelpb.GetPayProductDisplayNameReq{Code: code}); err == nil && dn.GetName() != "" {
				name = dn.GetName()
			}
		}
		if code != "" {
			items = []types.PayProductItem{{Code: code, Name: name}}
		}
	} else {
		lr, err := c.svcCtx.ChannelRpc.ListTerminalPayProducts(c.ctx, &channelpb.ListTerminalPayProductsReq{
			MerchantId: o.GetMerchantId(), Amount: o.GetAmount(),
		})
		if err != nil {
			return nil, err
		}
		opts := lr.GetProducts()
		items = make([]types.PayProductItem, 0, len(opts))
		for _, p := range opts {
			items = append(items, types.PayProductItem{Code: p.GetCode(), Name: p.GetName()})
		}
	}

	return &types.TerminalOrderResp{
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
		PayProducts: items,
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
	code := strings.TrimSpace(req.PayProductCode)

	if o.GetChannelLocked() == 0 {
		if code == "" {
			return nil, status.Error(codes.InvalidArgument, "pay_product_code required")
		}
		ok, err := c.svcCtx.ChannelRpc.MerchantHasPayProductCode(c.ctx, &channelpb.MerchantHasPayProductCodeReq{
			MerchantId: o.GetMerchantId(), PayProductCode: code,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "check merchant pay products failed")
		}
		if !ok.GetOk() {
			return nil, status.Error(codes.PermissionDenied, "pay_product_code not enabled for merchant")
		}
	}

	r, err := c.svcCtx.OrderRpc.PrepareTerminalPay(c.ctx, &orderclient.PrepareTerminalPayReq{
		OrderNo:        orderNo,
		PayProductCode: code,
	})
	if err != nil {
		return nil, err
	}
	return &types.TerminalPayResp{
		ChannelId:      r.GetChannelId(),
		PayProductId:   r.GetPayProductId(),
		PayProductCode: r.GetPayProductCode(),
		PayUrl:         r.GetPayUrl(),
		QrPayload:      r.GetQrPayload(),
		PayMode:        r.GetPayMode(),
	}, nil
}

func (c *Checkout) UpstreamNotify(req *types.UpstreamNotifyReq) (resp *types.UpstreamNotifyResp, err error) {
	if strings.TrimSpace(req.OrderNo) == "" || strings.TrimSpace(req.UpstreamTradeNo) == "" || req.ChannelId <= 0 || req.PaidAmount <= 0 {
		return notifyFail(NotifyCodeInvalidNotifyParams, "invalid notify params"), nil
	}

	signResp, err := c.svcCtx.ChannelRpc.GetSignSecret(c.ctx, &channelclient.GetSignSecretReq{ChannelId: req.ChannelId})
	if err != nil {
		return notifyFail(NotifyCodeChannelNotFound, "channel not found"), nil
	}

	expect := md5Sign(map[string]string{
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

	// 幂等与重放控制：
	// - 已支付订单仅接受与已落库支付快照完全一致的重复通知
	// - 非待支付（失败/关闭）不再接受支付成功通知
	if o.GetStatus() == 1 {
		if samePaidSnapshot(o, req) {
			return notifyOK(NotifyCodeIdempotentReplayAccepted, "idempotent replay accepted"), nil
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
		return notifyFail(NotifyCodeMarkPaidFailed, "mark paid failed"), nil
	}

	if !markResp.GetChanged() {
		// 并发场景：若另一条回调已先落库，允许同快照重放成功。
		latest, ge := c.svcCtx.OrderRpc.GetOrder(c.ctx, &orderclient.GetOrderReq{OrderNo: req.OrderNo})
		if ge != nil {
			return notifyFail(NotifyCodeMarkPaidRace, "mark paid race"), nil
		}
		if samePaidSnapshot(latest.GetOrder(), req) {
			return notifyOK(NotifyCodeIdempotentRaceAccepted, "idempotent race accepted"), nil
		}
		return notifyFail(NotifyCodeMarkPaidRaceMismatch, "mark paid race mismatch"), nil
	}

	_, _ = c.svcCtx.SettleRpc.Credit(c.ctx, &settleclient.CreditReq{
		MerchantId: o.GetMerchantId(),
		OrderNo:    o.GetOrderNo(),
		Amount:     req.PaidAmount,
		Reason:     "ORDER_PAID",
	})

	body, _ := json.Marshal(map[string]any{
		"merchant_id": o.GetMerchantId(),
		"order_no":    o.GetOrderNo(),
		"attempt":     0,
	})
	_ = c.svcCtx.NsqProducer.Publish(c.svcCtx.Config.Nsq.Topic, body)

	return notifyOK("", ""), nil
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

func md5Sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		k = strings.ToLower(k)
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	if b.Len() > 0 {
		b.WriteByte('&')
	}
	b.WriteString("key=")
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
