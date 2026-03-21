package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/merchant/merchantclient"
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
	r, err := l.svcCtx.MerchantRpc.CreateMerchant(l.ctx, &merchantclient.CreateMerchantReq{
		MerchantId:  merchantId,
		ApiSecret:   secret,
		Status:      1,
		RateBps:     req.RateBps,
		NotifyUrl:   req.NotifyUrl,
		ReturnUrl:   req.ReturnUrl,
		IpWhitelist: req.IpWhitelist,
	})
	if err != nil {
		return nil, err
	}
	created := r.GetMerchant()
	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:  created.GetMerchantId(),
			ApiSecret:   created.GetApiSecret(),
			Status:      created.GetStatus(),
			RateBps:     created.GetRateBps(),
			NotifyUrl:   created.GetNotifyUrl(),
			ReturnUrl:   created.GetReturnUrl(),
			IpWhitelist: created.GetIpWhitelist(),
			Balance:     created.GetBalance(),
		},
	}, nil
}
