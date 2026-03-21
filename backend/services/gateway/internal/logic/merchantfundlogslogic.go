package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantFundLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantFundLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantFundLogsLogic {
	return &MerchantFundLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantFundLogsLogic) MerchantFundLogs(req *types.MerchantFundLogsReq) (*types.MerchantFundLogsResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(l.ctx))
	logs, err := l.svcCtx.FundLogs.ListByMerchant(l.ctx, merchantId, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]types.MerchantFundLogItem, 0, len(logs))
	for _, f := range logs {
		out = append(out, types.MerchantFundLogItem{
			Id:            f.Id,
			OrderNo:       f.OrderNo,
			ChangeType:    f.ChangeType,
			Amount:        f.Amount,
			BalanceBefore: f.BalanceBefore,
			BalanceAfter:  f.BalanceAfter,
			Reason:        f.Reason,
			CreatedAt:     f.CreatedAt.Unix(),
		})
	}
	return &types.MerchantFundLogsResp{Logs: out}, nil
}
