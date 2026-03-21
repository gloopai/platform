package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminCreateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateMerchantLogic {
	return &AdminCreateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateMerchantLogic) AdminCreateMerchant(req *types.AdminCreateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := strings.TrimSpace(req.ApiSecret)
	if secret == "" {
		tok, err := newToken()
		if err != nil {
			return nil, err
		}
		secret = tok
	}

	m := &store.Merchant{
		MerchantId:  merchantId,
		ApiSecret:   secret,
		Status:      1,
		RateBps:     req.RateBps,
		NotifyUrl:   req.NotifyUrl,
		ReturnUrl:   req.ReturnUrl,
		IpWhitelist: req.IpWhitelist,
		Balance:     0,
	}
	if err := l.svcCtx.Merchants.Create(l.ctx, m); err != nil {
		return nil, err
	}
	created, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:  created.MerchantId,
			ApiSecret:   created.ApiSecret,
			Status:      created.Status,
			RateBps:     created.RateBps,
			NotifyUrl:   created.NotifyUrl,
			ReturnUrl:   created.ReturnUrl,
			IpWhitelist: created.IpWhitelist,
			Balance:     created.Balance,
		},
	}, nil
}

