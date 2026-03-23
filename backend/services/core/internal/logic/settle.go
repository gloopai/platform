package logic

import (
	"context"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type DebitPayoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDebitPayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DebitPayoutLogic {
	return &DebitPayoutLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DebitPayoutLogic) DebitPayout(in *settlepb.DebitPayoutReq) (*settlepb.DebitPayoutResp, error) {
	if in.GetMerchantId() == "" || in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	changed, payoutBalance, err := l.svcCtx.Settle.DebitPayout(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetAmount(), in.GetReason())
	if err != nil {
		if err == store.ErrInsufficientBalance {
			return nil, status.Error(codes.FailedPrecondition, "insufficient payout balance")
		}
		return nil, status.Error(codes.Internal, "debit payout failed")
	}
	return &settlepb.DebitPayoutResp{Changed: changed, PayoutBalance: payoutBalance}, nil
}

type TransferCollectToPayoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferCollectToPayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferCollectToPayoutLogic {
	return &TransferCollectToPayoutLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *TransferCollectToPayoutLogic) TransferCollectToPayout(in *settlepb.TransferCollectToPayoutReq) (*settlepb.TransferCollectToPayoutResp, error) {
	if in.GetMerchantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	changed, collectBalance, payoutBalance, err := l.svcCtx.Settle.TransferCollectToPayout(l.ctx, in.GetMerchantId(), in.GetAmount(), in.GetReason())
	if err != nil {
		if err == store.ErrInsufficientBalance {
			return nil, status.Error(codes.FailedPrecondition, "insufficient collect balance")
		}
		return nil, status.Error(codes.Internal, "transfer collect to payout failed")
	}
	return &settlepb.TransferCollectToPayoutResp{
		Changed:        changed,
		CollectBalance: collectBalance,
		PayoutBalance:  payoutBalance,
	}, nil
}

func NewCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreditLogic {
	return &CreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreditLogic) Credit(in *settlepb.CreditReq) (*settlepb.CreditResp, error) {
	if in.GetMerchantId() == "" || in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and order_no required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	changed, balance, err := l.svcCtx.Settle.Credit(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetAmount(), in.GetReason())
	if err != nil {
		return nil, status.Error(codes.Internal, "credit failed")
	}
	return &settlepb.CreditResp{
		Changed: changed,
		Balance: balance,
	}, nil
}
