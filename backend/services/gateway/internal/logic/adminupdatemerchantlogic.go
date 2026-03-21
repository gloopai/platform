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

	current, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}

	upd := &store.Merchant{
		MerchantId:  merchantId,
		ApiSecret:   current.ApiSecret,
		Status:      current.Status,
		RateBps:     current.RateBps,
		NotifyUrl:   current.NotifyUrl,
		ReturnUrl:   current.ReturnUrl,
		IpWhitelist: current.IpWhitelist,
		Balance:     current.Balance,
	}

	if req.Status == 0 || req.Status == 1 {
		upd.Status = req.Status
	}
	if req.RateBps != 0 {
		upd.RateBps = req.RateBps
	}
	if req.NotifyUrl != "" {
		upd.NotifyUrl = req.NotifyUrl
	}
	if req.ReturnUrl != "" {
		upd.ReturnUrl = req.ReturnUrl
	}
	if req.IpWhitelist != "" {
		upd.IpWhitelist = req.IpWhitelist
	}

	if err := l.svcCtx.Merchants.Update(l.ctx, merchantId, upd); err != nil {
		return nil, err
	}

	if req.ResetSecret {
		tok, err := newToken()
		if err != nil {
			return nil, err
		}
		if err := l.svcCtx.Merchants.UpdateSecret(l.ctx, merchantId, tok); err != nil {
			return nil, err
		}
	}

	updated, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}

	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:  updated.MerchantId,
			ApiSecret:   updated.ApiSecret,
			Status:      updated.Status,
			RateBps:     updated.RateBps,
			NotifyUrl:   updated.NotifyUrl,
			ReturnUrl:   updated.ReturnUrl,
			IpWhitelist: updated.IpWhitelist,
			Balance:     updated.Balance,
		},
	}, nil
}

