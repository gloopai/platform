package logic

import (
	"context"
	"errors"
	"strings"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	changed, availableBalance, err := l.svcCtx.Settle.DebitPayout(l.ctx, in.GetMerchantId(), in.GetOrderNo(), in.GetAmount(), in.GetReason())
	if err != nil {
		if err == store.ErrInsufficientBalance {
			return nil, status.Error(codes.FailedPrecondition, "insufficient available balance")
		}
		return nil, status.Error(codes.Internal, "debit payout failed")
	}
	return &settlepb.DebitPayoutResp{Changed: changed, AvailableBalance: availableBalance}, nil
}

type TransferPayinToPayoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferPayinToPayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferPayinToPayoutLogic {
	return &TransferPayinToPayoutLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *TransferPayinToPayoutLogic) TransferPayinToPayout(in *settlepb.TransferPayinToPayoutReq) (*settlepb.TransferPayinToPayoutResp, error) {
	if in.GetMerchantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	changed, payinBalance, availableBalance, err := l.svcCtx.Settle.TransferPayinToPayout(l.ctx, in.GetMerchantId(), in.GetAmount(), in.GetReason())
	if err != nil {
		if err == store.ErrInsufficientBalance {
			return nil, status.Error(codes.FailedPrecondition, "insufficient payin balance")
		}
		return nil, status.Error(codes.Internal, "transfer payin to payout failed")
	}
	return &settlepb.TransferPayinToPayoutResp{
		Changed:       changed,
		PayinBalance:  payinBalance,
		AvailableBalance: availableBalance,
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

type DepositAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepositAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositAvailableLogic {
	return &DepositAvailableLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DepositAvailableLogic) DepositAvailable(in *settlepb.DepositAvailableReq) (*settlepb.DepositAvailableResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	if in.GetAmountFiatCents() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount_fiat_cents must be positive")
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" {
		reason = "AVAILABLE_DEPOSIT"
	}
	orderNo, availableAfter, err := l.svcCtx.Settle.DepositAvailable(l.ctx, merchantId, in.GetAmountFiatCents(), reason)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, status.Error(codes.Internal, "deposit failed")
	}
	return &settlepb.DepositAvailableResp{OrderNo: orderNo, AvailableBalance: availableAfter}, nil
}
