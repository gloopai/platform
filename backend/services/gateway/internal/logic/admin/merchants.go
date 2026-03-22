package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
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

func toAdminMerchantInfo(m *merchantpb.MerchantInfo) types.AdminMerchantInfo {
	if m == nil {
		return types.AdminMerchantInfo{}
	}
	return types.AdminMerchantInfo{
		MerchantId:    m.GetMerchantId(),
		ApiSecret:     m.GetApiSecret(),
		Status:        m.GetStatus(),
		RateBps:       m.GetRateBps(),
		NotifyUrl:     m.GetNotifyUrl(),
		ReturnUrl:     m.GetReturnUrl(),
		IpWhitelist:   m.GetIpWhitelist(),
		Balance:       m.GetBalance(),
		PayProductIds: m.GetPayProductIds(),
	}
}

func (m *AdminMerchants) AdminListMerchants() (*types.AdminListMerchantsResp, error) {
	r, err := m.svcCtx.MerchantRpc.ListMerchants(m.ctx, &merchantclient.ListMerchantsReq{Limit: 200})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminMerchantInfo, 0, len(r.GetMerchants()))
	for _, row := range r.GetMerchants() {
		out = append(out, toAdminMerchantInfo(row))
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
		MerchantId:     merchantId,
		ApiSecret:      secret,
		Status:         1,
		RateBps:        req.RateBps,
		NotifyUrl:      req.NotifyUrl,
		ReturnUrl:      req.ReturnUrl,
		IpWhitelist:    req.IpWhitelist,
		PayProductIds:  req.PayProductIds,
	})
	if err != nil {
		return nil, err
	}
	created := r.GetMerchant()
	return &types.AdminUpsertMerchantResp{
		Merchant: toAdminMerchantInfo(created),
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
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayProducts(m.ctx, &merchantpb.ReplaceMerchantPayProductsReq{
			MerchantId:    merchantId,
			PayProductIds: req.PayProductIds,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}

	ids, err := m.svcCtx.MerchantRpc.ListMerchantPayProductIds(m.ctx, &merchantpb.ListMerchantPayProductIdsReq{MerchantId: merchantId})
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}
	// 用最新白名单 ID 覆盖展示
	_ = ids
	mi := toAdminMerchantInfo(updated)
	mi.PayProductIds = ids.GetPayProductIds()
	return &types.AdminUpsertMerchantResp{Merchant: mi}, nil
}
