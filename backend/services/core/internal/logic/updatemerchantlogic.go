package logic

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMerchantLogic {
	return &UpdateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMerchantLogic) UpdateMerchant(in *merchantpb.UpdateMerchantReq) (*merchantpb.UpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	existing, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}

	secret := strings.TrimSpace(in.GetApiSecret())
	if secret == "" {
		secret = existing.ApiSecret
	}

	statusVal := in.GetStatus()
	if statusVal == 0 {
		statusVal = existing.Status
	}

	rec := &store.Merchant{
		MerchantId:  merchantId,
		ApiSecret:   secret,
		Status:      statusVal,
		RateBps:     in.GetRateBps(),
		IpWhitelist: in.GetIpWhitelist(),
		NotifyUrl:   in.GetNotifyUrl(),
		ReturnUrl:   in.GetReturnUrl(),
	}
	if err := l.svcCtx.Merchants.UpdateByMerchantId(l.ctx, merchantId, rec); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(updated)}, nil
}
