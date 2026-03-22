package logic

import (
	"context"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListFundLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFundLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFundLogsLogic {
	return &ListFundLogsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListFundLogsLogic) ListFundLogs(in *settlepb.ListFundLogsReq) (*settlepb.ListFundLogsResp, error) {
	rows, err := l.svcCtx.Settle.ListByMerchant(l.ctx, in.GetMerchantId(), in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*settlepb.FundLogItem, 0, len(rows))
	for _, f := range rows {
		out = append(out, &settlepb.FundLogItem{
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
	return &settlepb.ListFundLogsResp{Logs: out}, nil
}
