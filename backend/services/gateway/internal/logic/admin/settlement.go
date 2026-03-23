package logic

import (
	"context"
	"strings"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminSettlement 管理台结算中心（MVP：平台资金流水只读）。
type AdminSettlement struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSettlement(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSettlement {
	return &AdminSettlement{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminSettlement) AdminSettlementLogs(req *types.AdminSettlementLogsReq) (*types.AdminSettlementLogsResp, error) {
	r, err := a.svcCtx.SettleRpc.ListFundLogs(a.ctx, &settlepb.ListFundLogsReq{
		MerchantId: strings.TrimSpace(req.MerchantId),
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminSettlementLogItem, 0, len(r.GetLogs()))
	for _, x := range r.GetLogs() {
		out = append(out, types.AdminSettlementLogItem{
			Id:            x.GetId(),
			MerchantId:    x.GetMerchantId(),
			OrderNo:       x.GetOrderNo(),
			ChangeType:    x.GetChangeType(),
			Amount:        x.GetAmount(),
			BalanceBefore: x.GetBalanceBefore(),
			BalanceAfter:  x.GetBalanceAfter(),
			Reason:        x.GetReason(),
			CreatedAt:     x.GetCreatedAt(),
		})
	}
	return &types.AdminSettlementLogsResp{Logs: out}, nil
}
