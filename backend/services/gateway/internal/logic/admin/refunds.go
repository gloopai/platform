package logic

import (
	"context"
	"sort"
	"strings"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminRefunds 管理台退款与差错（MVP：失败/关闭订单候选只读）。
type AdminRefunds struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRefunds(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRefunds {
	return &AdminRefunds{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminRefunds) AdminListRefundCandidates(req *types.AdminRefundsReq) (*types.AdminRefundsResp, error) {
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	statusFilter := strings.ToLower(strings.TrimSpace(req.Status))
	if statusFilter == "" {
		statusFilter = "all"
	}
	wantFailed := statusFilter == "all" || statusFilter == "failed"
	wantClosed := statusFilter == "all" || statusFilter == "closed"

	var rows []*orderpb.OrderInfo
	appendByStatus := func(status int32) error {
		s := status
		r, err := a.svcCtx.OrderRpc.AdminListPayOrders(a.ctx, &orderpb.AdminListOrdersReq{
			MerchantId: strings.TrimSpace(req.MerchantId),
			Keyword:    strings.TrimSpace(req.Keyword),
			Status:     &s,
			Limit:      limit,
		})
		if err != nil {
			return err
		}
		rows = append(rows, r.GetOrders()...)
		return nil
	}

	if wantFailed {
		if err := appendByStatus(2); err != nil {
			return nil, err
		}
	}
	if wantClosed {
		if err := appendByStatus(3); err != nil {
			return nil, err
		}
	}

	uniq := make(map[string]*orderpb.OrderInfo, len(rows))
	for _, o := range rows {
		if o == nil || strings.TrimSpace(o.GetOrderNo()) == "" {
			continue
		}
		uniq[o.GetOrderNo()] = o
	}

	flat := make([]*orderpb.OrderInfo, 0, len(uniq))
	for _, o := range uniq {
		flat = append(flat, o)
	}
	sort.Slice(flat, func(i, j int) bool {
		return flat[i].GetCreatedAt() > flat[j].GetCreatedAt()
	})
	if int64(len(flat)) > limit {
		flat = flat[:limit]
	}

	out := make([]types.AdminRefundRow, 0, len(flat))
	for _, o := range flat {
		label := "未知"
		switch o.GetStatus() {
		case 2:
			label = "支付失败"
		case 3:
			label = "已关闭"
		}
		out = append(out, types.AdminRefundRow{
			OrderNo:         o.GetOrderNo(),
			MerchantId:      o.GetMerchantId(),
			MerchantOrderNo: o.GetMerchantOrderNo(),
			Amount:          o.GetAmount(),
			Currency:        o.GetCurrency(),
			Status:          o.GetStatus(),
			StatusLabel:     label,
			ChannelId:       o.GetChannelId(),
			PayinProductCode:  o.GetPayinProductCode(),
			UpstreamTradeNo: o.GetUpstreamTradeNo(),
			CreatedAt:       o.GetCreatedAt(),
		})
	}

	return &types.AdminRefundsResp{Items: out}, nil
}
