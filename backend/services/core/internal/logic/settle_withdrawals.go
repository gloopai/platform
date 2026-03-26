package logic

import (
	"context"
	"strings"

	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WithdrawalsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalsLogic {
	return &WithdrawalsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func toPbWithdrawal(x store.WithdrawalRow) *settlepb.WithdrawalItem {
	var reviewedAt, payoutedAt int64
	if x.ReviewedAt != nil {
		reviewedAt = x.ReviewedAt.Unix()
	}
	if x.PayoutedAt != nil {
		payoutedAt = x.PayoutedAt.Unix()
	}
	return &settlepb.WithdrawalItem{
		Id:             x.Id,
		WithdrawNo:     x.WithdrawNo,
		MerchantId:     x.MerchantId,
		ApplyAmount:    x.ApplyAmount,
		FeeAmount:      x.FeeAmount,
		NetAmount:      x.NetAmount,
		FiatDebitAmount: x.FiatDebitAmount,
		Status:         x.Status,
		ReceiveAccount: x.ReceiveAccount,
		ReceiveName:    x.ReceiveName,
		BankName:       x.BankName,
		ApplyNote:      x.ApplyNote,
		ReviewNote:     x.ReviewNote,
		PayoutNote:     x.PayoutNote,
		ReviewedBy:     x.ReviewedBy,
		ReviewedAt:     reviewedAt,
		PayoutedBy:     x.PayoutedBy,
		PayoutedAt:     payoutedAt,
		CreatedAt:      x.CreatedAt.Unix(),
		UpdatedAt:      x.UpdatedAt.Unix(),
	}
}

func (l *WithdrawalsLogic) CreateWithdrawal(in *settlepb.CreateWithdrawalReq) (*settlepb.CreateWithdrawalResp, error) {
	if strings.TrimSpace(in.GetWithdrawNo()) == "" || strings.TrimSpace(in.GetMerchantId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "withdraw_no and merchant_id required")
	}
	if in.GetApplyAmount() <= 0 || in.GetFeeAmount() < 0 || in.GetFeeAmount() > in.GetApplyAmount() {
		return nil, status.Error(codes.InvalidArgument, "invalid apply_amount/fee_amount")
	}
	if in.GetFiatDebitAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "fiat_debit_amount must be positive")
	}
	row, err := l.svcCtx.Settle.CreateWithdrawal(l.ctx, store.WithdrawalRow{
		WithdrawNo:     strings.TrimSpace(in.GetWithdrawNo()),
		MerchantId:     strings.TrimSpace(in.GetMerchantId()),
		ApplyAmount:    in.GetApplyAmount(),
		FeeAmount:      in.GetFeeAmount(),
		FiatDebitAmount: in.GetFiatDebitAmount(),
		ReceiveAccount: strings.TrimSpace(in.GetReceiveAccount()),
		ReceiveName:    strings.TrimSpace(in.GetReceiveName()),
		BankName:       strings.TrimSpace(in.GetBankName()),
		ApplyNote:      strings.TrimSpace(in.GetApplyNote()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "create withdrawal failed")
	}
	return &settlepb.CreateWithdrawalResp{Item: toPbWithdrawal(row)}, nil
}

func (l *WithdrawalsLogic) ListWithdrawals(in *settlepb.ListWithdrawalsReq) (*settlepb.ListWithdrawalsResp, error) {
	rows, err := l.svcCtx.Settle.ListWithdrawals(l.ctx, strings.TrimSpace(in.GetMerchantId()), in.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "list withdrawals failed")
	}
	items := make([]*settlepb.WithdrawalItem, 0, len(rows))
	for _, x := range rows {
		items = append(items, toPbWithdrawal(x))
	}
	return &settlepb.ListWithdrawalsResp{Items: items}, nil
}

func (l *WithdrawalsLogic) ReviewWithdrawal(in *settlepb.ReviewWithdrawalReq) (*settlepb.ReviewWithdrawalResp, error) {
	withdrawNo := strings.TrimSpace(in.GetWithdrawNo())
	if withdrawNo == "" {
		return nil, status.Error(codes.InvalidArgument, "withdraw_no required")
	}
	row, _, _, err := l.svcCtx.Settle.ReviewWithdrawal(
		l.ctx,
		withdrawNo,
		in.GetApproved(),
		strings.TrimSpace(in.GetReviewNote()),
		strings.TrimSpace(in.GetOperator()),
	)
	if err != nil {
		if err == store.ErrInsufficientBalance {
			return nil, status.Error(codes.FailedPrecondition, "insufficient available balance")
		}
		if err == store.ErrWithdrawalNotFound {
			return nil, status.Error(codes.NotFound, "withdrawal not found")
		}
		return nil, status.Error(codes.FailedPrecondition, "review withdrawal failed")
	}
	return &settlepb.ReviewWithdrawalResp{Item: toPbWithdrawal(row)}, nil
}

func (l *WithdrawalsLogic) PayoutWithdrawal(in *settlepb.PayoutWithdrawalReq) (*settlepb.PayoutWithdrawalResp, error) {
	withdrawNo := strings.TrimSpace(in.GetWithdrawNo())
	if withdrawNo == "" {
		return nil, status.Error(codes.InvalidArgument, "withdraw_no required")
	}
	row, _, err := l.svcCtx.Settle.MarkWithdrawalPayoutSuccess(
		l.ctx,
		withdrawNo,
		strings.TrimSpace(in.GetPayoutNote()),
		strings.TrimSpace(in.GetOperator()),
	)
	if err != nil {
		if err == store.ErrWithdrawalNotFound {
			return nil, status.Error(codes.NotFound, "withdrawal not found")
		}
		if err == store.ErrInvalidWithdrawalStatus {
			return nil, status.Error(codes.FailedPrecondition, "withdrawal status invalid")
		}
		return nil, status.Error(codes.Internal, "payout withdrawal failed")
	}
	return &settlepb.PayoutWithdrawalResp{Item: toPbWithdrawal(row)}, nil
}
