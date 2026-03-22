package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminUpdateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateMerchantLogic {
	return &AdminUpdateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateMerchantLogic) AdminUpdateMerchant(req *types.AdminUpdateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := ""
	if req.ResetSecret {
		tok, err := newToken()
		if err != nil {
			return nil, err
		}
		secret = tok
	}
	r, err := l.svcCtx.MerchantRpc.UpdateMerchant(l.ctx, &merchantclient.UpdateMerchantReq{
		MerchantId:  merchantId,
		ApiSecret:   secret,
		Status:      req.Status,
		RateBps:     req.RateBps,
		NotifyUrl:   req.NotifyUrl,
		ReturnUrl:   req.ReturnUrl,
		IpWhitelist: req.IpWhitelist,
	})
	if err != nil {
		return nil, err
	}
	updated := r.GetMerchant()

	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:  updated.GetMerchantId(),
			ApiSecret:   updated.GetApiSecret(),
			Status:      updated.GetStatus(),
			RateBps:     updated.GetRateBps(),
			NotifyUrl:   updated.GetNotifyUrl(),
			ReturnUrl:   updated.GetReturnUrl(),
			IpWhitelist: updated.GetIpWhitelist(),
			Balance:     updated.GetBalance(),
		},
	}, nil
}
