package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminMerchants 管理后台商户（入驻商户）与商户侧支付产品白名单。
type AdminMerchants struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminMerchants(ctx context.Context, svcCtx *svc.ServiceContext) *AdminMerchants {
	return &AdminMerchants{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (m *AdminMerchants) AdminListMerchants() (*types.AdminListMerchantsResp, error) {
	r, err := m.svcCtx.MerchantRpc.ListMerchants(m.ctx, &merchantclient.ListMerchantsReq{Limit: 200})
	if err != nil {
		return nil, err
	}
	items := r.GetMerchants()
	out := make([]types.AdminMerchantInfo, 0, len(items))
	for _, row := range items {
		ids, err := m.svcCtx.MerchantPayProducts.ListProductIDs(m.ctx, row.GetMerchantId())
		if err != nil {
			return nil, status.Error(codes.Internal, "load merchant pay products failed")
		}
		out = append(out, types.AdminMerchantInfo{
			MerchantId:    row.GetMerchantId(),
			ApiSecret:     row.GetApiSecret(),
			Status:        row.GetStatus(),
			RateBps:       row.GetRateBps(),
			NotifyUrl:     row.GetNotifyUrl(),
			ReturnUrl:     row.GetReturnUrl(),
			IpWhitelist:   row.GetIpWhitelist(),
			Balance:       row.GetBalance(),
			PayProductIds: ids,
		})
	}
	return &types.AdminListMerchantsResp{Merchants: out}, nil
}

func (m *AdminMerchants) AdminCreateMerchant(req *types.AdminCreateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := strings.TrimSpace(req.ApiSecret)
	r, err := m.svcCtx.MerchantRpc.CreateMerchant(m.ctx, &merchantclient.CreateMerchantReq{
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

	if err := m.svcCtx.MerchantPayProducts.Replace(m.ctx, merchantId, req.PayProductIds); err != nil {
		return nil, status.Error(codes.Internal, "save merchant pay products failed")
	}

	ids, err := m.svcCtx.MerchantPayProducts.ListProductIDs(m.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}

	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:    created.GetMerchantId(),
			ApiSecret:     created.GetApiSecret(),
			Status:        created.GetStatus(),
			RateBps:       created.GetRateBps(),
			NotifyUrl:     created.GetNotifyUrl(),
			ReturnUrl:     created.GetReturnUrl(),
			IpWhitelist:   created.GetIpWhitelist(),
			Balance:       created.GetBalance(),
			PayProductIds: ids,
		},
	}, nil
}

func (m *AdminMerchants) AdminUpdateMerchant(req *types.AdminUpdateMerchantReq) (*types.AdminUpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	secret := ""
	if req.ResetSecret {
		tok, err := shared.NewToken()
		if err != nil {
			return nil, err
		}
		secret = tok
	}
	r, err := m.svcCtx.MerchantRpc.UpdateMerchant(m.ctx, &merchantclient.UpdateMerchantReq{
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

	if req.PayProductIds != nil {
		if err := m.svcCtx.MerchantPayProducts.Replace(m.ctx, merchantId, req.PayProductIds); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}

	ids, err := m.svcCtx.MerchantPayProducts.ListProductIDs(m.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}

	return &types.AdminUpsertMerchantResp{
		Merchant: types.AdminMerchantInfo{
			MerchantId:    updated.GetMerchantId(),
			ApiSecret:     updated.GetApiSecret(),
			Status:        updated.GetStatus(),
			RateBps:       updated.GetRateBps(),
			NotifyUrl:     updated.GetNotifyUrl(),
			ReturnUrl:     updated.GetReturnUrl(),
			IpWhitelist:   updated.GetIpWhitelist(),
			Balance:       updated.GetBalance(),
			PayProductIds: ids,
		},
	}, nil
}
