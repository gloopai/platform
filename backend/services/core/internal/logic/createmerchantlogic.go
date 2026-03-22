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

type CreateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMerchantLogic {
	return &CreateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMerchantLogic) CreateMerchant(in *merchantpb.CreateMerchantReq) (*merchantpb.UpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := strings.TrimSpace(in.GetApiSecret())
	if secret == "" {
		v, err := newSecret()
		if err != nil {
			return nil, err
		}
		secret = v
	}

	statusVal := in.GetStatus()
	if statusVal == 0 {
		statusVal = 1
	}

	rec := &store.Merchant{
		MerchantId:  merchantId,
		ApiSecret:   secret,
		Status:      statusVal,
		RateBps:     in.GetRateBps(),
		IpWhitelist: in.GetIpWhitelist(),
		NotifyUrl:   in.GetNotifyUrl(),
		ReturnUrl:   in.GetReturnUrl(),
		Balance:     0,
	}
	if err := l.svcCtx.Merchants.Create(l.ctx, rec); err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "merchant already exists")
		}
		return nil, err
	}

	created, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "merchant create failed")
		}
		return nil, err
	}
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(created)}, nil
}
