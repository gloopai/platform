package logic

import (
	"context"
	"strings"

	orderpb "github.com/gloopai/pay/common/pb/order"
	"github.com/gloopai/pay/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TodaySummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTodaySummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TodaySummaryLogic {
	return &TodaySummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TodaySummaryLogic) TodaySummary(in *orderpb.TodaySummaryReq) (*orderpb.TodaySummaryResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	totalAmount, totalCount, successCount, err := l.svcCtx.Orders.TodaySummary(l.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "today summary failed")
	}

	return &orderpb.TodaySummaryResp{
		TotalAmount:  totalAmount,
		TotalCount:   totalCount,
		SuccessCount: successCount,
	}, nil
}
