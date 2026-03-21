// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/order/orderclient"
	"github.com/gloopai/pay/settle/settleclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpstreamNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpstreamNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpstreamNotifyLogic {
	return &UpstreamNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpstreamNotifyLogic) UpstreamNotify(req *types.UpstreamNotifyReq) (resp *types.UpstreamNotifyResp, err error) {
	ch, err := l.svcCtx.Channels.GetByID(l.ctx, req.ChannelId)
	if err != nil {
		return &types.UpstreamNotifyResp{Ok: false}, nil
	}

	expect := md5Sign(map[string]string{
		"order_no":          req.OrderNo,
		"paid_amount":       strconv.FormatInt(req.PaidAmount, 10),
		"upstream_trade_no": req.UpstreamTradeNo,
		"channel_id":        strconv.FormatInt(req.ChannelId, 10),
		"sign":              req.Sign,
	}, ch.SignSecret)
	if !strings.EqualFold(expect, req.Sign) {
		return &types.UpstreamNotifyResp{Ok: false}, nil
	}

	getResp, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderReq{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return &types.UpstreamNotifyResp{Ok: false}, nil
	}
	o := getResp.GetOrder()

	markResp, err := l.svcCtx.OrderRpc.MarkPaid(l.ctx, &orderclient.MarkPaidReq{
		OrderNo:         req.OrderNo,
		PaidAmount:      req.PaidAmount,
		UpstreamTradeNo: req.UpstreamTradeNo,
		ChannelId:       req.ChannelId,
	})
	if err != nil {
		return &types.UpstreamNotifyResp{Ok: false}, nil
	}

	if markResp.GetChanged() {
		_, _ = l.svcCtx.SettleRpc.Credit(l.ctx, &settleclient.CreditReq{
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
		_ = l.svcCtx.NsqProducer.Publish(l.svcCtx.Config.Nsq.Topic, body)
	}

	return &types.UpstreamNotifyResp{Ok: true}, nil
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
