package logic

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/gateway/internal/logic/fundlog"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		ct := x.GetChangeType()
		out = append(out, types.AdminSettlementLogItem{
			Id:            x.GetId(),
			MerchantId:    x.GetMerchantId(),
			OrderNo:       x.GetOrderNo(),
			ChangeType:    ct,
			AccountType:   fundlog.AccountTypeFromChangeType(ct),
			Amount:        x.GetAmount(),
			BalanceBefore: x.GetBalanceBefore(),
			BalanceAfter:  x.GetBalanceAfter(),
			Reason:        x.GetReason(),
			CreatedAt:     x.GetCreatedAt(),
		})
	}
	return &types.AdminSettlementLogsResp{Logs: out}, nil
}

func mapWithdrawal(x *settlepb.WithdrawalItem) types.AdminWithdrawalItem {
	return types.AdminWithdrawalItem{
		WithdrawNo:      x.GetWithdrawNo(),
		MerchantId:      x.GetMerchantId(),
		ApplyAmount:     x.GetApplyAmount(),
		FeeAmount:       x.GetFeeAmount(),
		NetAmount:       x.GetNetAmount(),
		FiatDebitAmount: x.GetFiatDebitAmount(),
		Status:          x.GetStatus(),
		Currency:        "USDT",
		ReceiveAccount:  x.GetReceiveAccount(),
		ReceiveName:     x.GetReceiveName(),
		BankName:        x.GetBankName(),
		ApplyNote:       x.GetApplyNote(),
		ReviewNote:      x.GetReviewNote(),
		PayoutNote:      x.GetPayoutNote(),
		CreatedAt:       x.GetCreatedAt(),
		ReviewedAt:      x.GetReviewedAt(),
		PayoutedAt:      x.GetPayoutedAt(),
	}
}

func (a *AdminSettlement) AdminSettlementWithdrawals(req *types.AdminSettlementWithdrawalsReq) (*types.AdminSettlementWithdrawalsResp, error) {
	r, err := a.svcCtx.SettleRpc.ListWithdrawals(a.ctx, &settlepb.ListWithdrawalsReq{
		MerchantId: strings.TrimSpace(req.MerchantId),
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.AdminWithdrawalItem, 0, len(r.GetItems()))
	for _, x := range r.GetItems() {
		items = append(items, mapWithdrawal(x))
	}
	return &types.AdminSettlementWithdrawalsResp{Items: items}, nil
}

func (a *AdminSettlement) AdminCreateWithdrawal(req *types.AdminCreateWithdrawalReq) (*types.AdminCreateWithdrawalResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	if req.ApplyAmount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "apply_amount must be positive")
	}
	if req.FeeAmount < 0 {
		return nil, status.Error(codes.InvalidArgument, "fee_amount must be >= 0")
	}
	if req.ApplyAmount < req.FeeAmount {
		return nil, status.Error(codes.InvalidArgument, "apply_amount must be >= fee_amount")
	}
	ds, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
	if err != nil {
		return nil, err
	}
	rate := ds.GetFiatToUsdtRate()
	if rate <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "invalid fiat_to_usdt_rate")
	}
	mr, err := a.svcCtx.MerchantRpc.GetMerchant(a.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantId})
	if err != nil || mr.GetMerchant() == nil {
		return nil, status.Error(codes.NotFound, "merchant not found")
	}
	balanceSource := strings.ToUpper(strings.TrimSpace(req.BalanceSource))
	sourceFiatCents := mr.GetMerchant().GetAvailableBalance()
	switch balanceSource {
	case "", "AVAILABLE":
		balanceSource = "AVAILABLE"
		sourceFiatCents = mr.GetMerchant().GetAvailableBalance()
	case "PAYIN":
		sourceFiatCents = mr.GetMerchant().GetPayinBalance()
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid balance_source")
	}
	maxApplyUsdtCents := int64(math.Floor(float64(sourceFiatCents) / rate))
	if req.ApplyAmount > maxApplyUsdtCents {
		return nil, status.Error(codes.FailedPrecondition, "withdraw amount exceeds max withdrawable usdt")
	}
	fiatDebitAmount := int64(math.Ceil(float64(req.ApplyAmount) * rate))
	withdrawNo := fmt.Sprintf("WD-%s-%d", merchantId, time.Now().UnixMilli())
	reason := strings.TrimSpace(req.ApplyNote)
	if reason == "" {
		reason = "ADMIN_WITHDRAW_APPLY"
	}
	r, err := a.svcCtx.SettleRpc.CreateWithdrawal(a.ctx, &settlepb.CreateWithdrawalReq{
		WithdrawNo:      withdrawNo,
		MerchantId:      merchantId,
		ApplyAmount:     req.ApplyAmount,
		FeeAmount:       req.FeeAmount,
		FiatDebitAmount: fiatDebitAmount,
		ReceiveAccount:  strings.TrimSpace(req.ReceiveAccount),
		ReceiveName:     strings.TrimSpace(req.ReceiveName),
		BankName:        strings.TrimSpace(req.BankName),
		ApplyNote:       reason,
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminCreateWithdrawalResp{Item: mapWithdrawal(r.GetItem())}, nil
}

func (a *AdminSettlement) AdminReviewWithdrawal(req *types.AdminReviewWithdrawalReq) (*types.AdminReviewWithdrawalResp, error) {
	r, err := a.svcCtx.SettleRpc.ReviewWithdrawal(a.ctx, &settlepb.ReviewWithdrawalReq{
		WithdrawNo: strings.TrimSpace(req.WithdrawNo),
		Approved:   req.Approved,
		ReviewNote: strings.TrimSpace(req.ReviewNote),
		Operator:   strings.TrimSpace(req.Operator),
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminReviewWithdrawalResp{Item: mapWithdrawal(r.GetItem())}, nil
}

func (a *AdminSettlement) AdminPayoutWithdrawal(req *types.AdminPayoutWithdrawalReq) (*types.AdminPayoutWithdrawalResp, error) {
	r, err := a.svcCtx.SettleRpc.PayoutWithdrawal(a.ctx, &settlepb.PayoutWithdrawalReq{
		WithdrawNo: strings.TrimSpace(req.WithdrawNo),
		PayoutNote: strings.TrimSpace(req.PayoutNote),
		Operator:   strings.TrimSpace(req.Operator),
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminPayoutWithdrawalResp{Item: mapWithdrawal(r.GetItem())}, nil
}

func (a *AdminSettlement) AdminDeposit(req *types.AdminDepositReq) (*types.AdminDepositResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	mr, err := a.svcCtx.MerchantRpc.GetMerchant(a.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantId})
	if err != nil || mr.GetMerchant() == nil {
		return nil, status.Error(codes.NotFound, "merchant not found")
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "fiat"
	}
	note := strings.TrimSpace(req.Note)

	var fiatCents int64
	var reason string

	switch mode {
	case "fiat":
		if req.FiatAmountCents <= 0 {
			return nil, status.Error(codes.InvalidArgument, "fiat_amount_cents must be positive")
		}
		fiatCents = req.FiatAmountCents
		if note != "" {
			reason = fmt.Sprintf("法币存入 | %s", note)
		} else {
			reason = "法币存入"
		}
	case "usdt":
		ds, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
		if err != nil {
			return nil, err
		}
		rate := ds.GetFiatToUsdtRate()
		if rate <= 0 {
			return nil, status.Error(codes.FailedPrecondition, "invalid fiat_to_usdt_rate")
		}
		if req.UsdtAmountCents <= 0 {
			return nil, status.Error(codes.InvalidArgument, "usdt_amount_cents must be positive")
		}
		fiatCents = int64(math.Floor(float64(req.UsdtAmountCents) * rate))
		if fiatCents <= 0 {
			return nil, status.Error(codes.InvalidArgument, "converted fiat amount is zero")
		}
		if note != "" {
			reason = fmt.Sprintf("USDT存入(usdt分=%d)->法币分=%d | %s", req.UsdtAmountCents, fiatCents, note)
		} else {
			reason = fmt.Sprintf("USDT存入(usdt分=%d)->法币分=%d", req.UsdtAmountCents, fiatCents)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "mode must be fiat or usdt")
	}

	r, err := a.svcCtx.SettleRpc.DepositAvailable(a.ctx, &settlepb.DepositAvailableReq{
		MerchantId:      merchantId,
		AmountFiatCents: fiatCents,
		Reason:          reason,
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminDepositResp{
		OrderNo:           r.GetOrderNo(),
		AvailableBalance:  r.GetAvailableBalance(),
		FiatCreditedCents: fiatCents,
		Mode:              mode,
	}, nil
}
