package logic

import (
	"context"
	"fmt"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	settlepb "github.com/gloopai/pay/common/pb/settle"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
	cg := make([]types.MerchantPayinGrant, 0, len(m.GetPayinGrants()))
	for _, g := range m.GetPayinGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayinGrant{PayinProductId: g.GetPayinProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		cg = append(cg, row)
	}
	pg := make([]types.MerchantPayoutGrant, 0, len(m.GetPayoutGrants()))
	for _, g := range m.GetPayoutGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayoutGrant{PayoutProductId: g.GetPayoutProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		row.FeeMode = g.GetFeeMode()
		if row.FeeMode <= 0 {
			row.FeeMode = 1
		}
		row.FeeFixedAmount = g.GetFeeFixedAmount()
		pg = append(pg, row)
	}
	return types.AdminMerchantInfo{
		MerchantId:           m.GetMerchantId(),
		AppId:                m.GetAppId(),
		Email:                m.GetEmail(),
		AppSecret:            m.GetAppSecret(),
		Status:               m.GetStatus(),
		DefaultPayinRateBps:  m.GetDefaultPayinRateBps(),
		DefaultPayoutRateBps: m.GetDefaultPayoutRateBps(),
		NotifyUrl:            m.GetNotifyUrl(),
		ReturnUrl:            m.GetReturnUrl(),
		IpWhitelist:          m.GetIpWhitelist(),
		PayinBalance:         m.GetPayinBalance(),
		AvailableBalance:     m.GetAvailableBalance(),
		PayinProductIds:      m.GetPayinProductIds(),
		PayoutProductIds:     m.GetPayoutProductIds(),
		PayinGrants:          cg,
		PayoutGrants:         pg,
	}
}

func (m *AdminMerchants) AdminTransferPayinToPayout(req *types.AdminTransferPayinToPayoutReq) (*types.AdminTransferPayinToPayoutResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	r, err := m.svcCtx.SettleRpc.TransferPayinToPayout(m.ctx, &settlepb.TransferPayinToPayoutReq{
		MerchantId: merchantId,
		Amount:     req.Amount,
		Reason:     strings.TrimSpace(req.Reason),
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminTransferPayinToPayoutResp{
		Ok:               r.GetChanged(),
		PayinBalance:     r.GetPayinBalance(),
		AvailableBalance: r.GetAvailableBalance(),
	}, nil
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
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}

	appID, err := newAppID()
	if err != nil {
		return nil, err
	}
	password, err := newMerchantPassword()
	if err != nil {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	r, err := m.svcCtx.MerchantRpc.CreateMerchant(m.ctx, &merchantclient.CreateMerchantReq{
		MerchantId:           merchantId,
		AppId:                appID,
		Email:                email,
		PasswordHash:         string(passwordHash),
		Status:               1,
		DefaultPayinRateBps:  req.DefaultPayinRateBps,
		DefaultPayoutRateBps: req.DefaultPayoutRateBps,
		NotifyUrl:            req.NotifyUrl,
		ReturnUrl:            req.ReturnUrl,
		IpWhitelist:          req.IpWhitelist,
		PayinProductIds:      req.PayinProductIds,
		PayoutProductIds:     req.PayoutProductIds,
	})
	if err != nil {
		return nil, err
	}
	created := r.GetMerchant()
	return &types.AdminUpsertMerchantResp{
		Merchant:          toAdminMerchantInfo(created),
		GeneratedPassword: password,
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
	passwordHash := ""
	generatedPassword := ""
	if req.ResetPassword {
		pwd, err := newMerchantPassword()
		if err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = string(hash)
		generatedPassword = pwd
	}
	r, err := m.svcCtx.MerchantRpc.UpdateMerchant(m.ctx, &merchantclient.UpdateMerchantReq{
		MerchantId:           merchantId,
		AppSecret:            secret,
		PasswordHash:         passwordHash,
		Status:               req.Status,
		DefaultPayinRateBps:  req.DefaultPayinRateBps,
		DefaultPayoutRateBps: req.DefaultPayoutRateBps,
		NotifyUrl:            req.NotifyUrl,
		ReturnUrl:            req.ReturnUrl,
		IpWhitelist:          req.IpWhitelist,
	})
	if err != nil {
		return nil, err
	}
	updated := r.GetMerchant()

	if req.PayinGrants != nil {
		pbGrants := make([]*merchantpb.MerchantPayinGrant, 0, len(req.PayinGrants))
		for _, g := range req.PayinGrants {
			row := &merchantpb.MerchantPayinGrant{PayinProductId: g.PayinProductId}
			if g.MerchantRateBps != nil {
				v := *g.MerchantRateBps
				row.MerchantRateBps = &v
			}
			pbGrants = append(pbGrants, row)
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayinProducts(m.ctx, &merchantpb.ReplaceMerchantPayinProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	} else if req.PayinProductIds != nil {
		var pbGrants []*merchantpb.MerchantPayinGrant
		for _, id := range req.PayinProductIds {
			if id <= 0 {
				continue
			}
			pbGrants = append(pbGrants, &merchantpb.MerchantPayinGrant{PayinProductId: id})
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayinProducts(m.ctx, &merchantpb.ReplaceMerchantPayinProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}

	if req.PayoutGrants != nil {
		pbGrants := make([]*merchantpb.MerchantPayoutGrant, 0, len(req.PayoutGrants))
		for _, g := range req.PayoutGrants {
			row := &merchantpb.MerchantPayoutGrant{
				PayoutProductId: g.PayoutProductId,
				FeeMode:         g.FeeMode,
				FeeFixedAmount:  g.FeeFixedAmount,
			}
			if row.FeeMode <= 0 {
				row.FeeMode = 1
			}
			if g.MerchantRateBps != nil {
				v := *g.MerchantRateBps
				row.MerchantRateBps = &v
			}
			pbGrants = append(pbGrants, row)
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayoutProducts(m.ctx, &merchantpb.ReplaceMerchantPayoutProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant payout products failed")
		}
	} else if req.PayoutProductIds != nil {
		var pbGrants []*merchantpb.MerchantPayoutGrant
		for _, id := range req.PayoutProductIds {
			if id <= 0 {
				continue
			}
			pbGrants = append(pbGrants, &merchantpb.MerchantPayoutGrant{PayoutProductId: id})
		}
		if _, err := m.svcCtx.MerchantRpc.ReplaceMerchantPayoutProducts(m.ctx, &merchantpb.ReplaceMerchantPayoutProductsReq{
			MerchantId: merchantId,
			Grants:     pbGrants,
		}); err != nil {
			return nil, status.Error(codes.Internal, "save merchant payout products failed")
		}
	}

	gr, err := m.svcCtx.MerchantRpc.ListMerchantPayinProductIds(m.ctx, &merchantpb.ListMerchantPayinProductIdsReq{MerchantId: merchantId})
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}
	pr, err := m.svcCtx.MerchantRpc.ListMerchantPayoutProductIds(m.ctx, &merchantpb.ListMerchantPayoutProductIdsReq{MerchantId: merchantId})
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant payout products failed")
	}
	mi := toAdminMerchantInfo(updated)
	// 覆盖为最新白名单
	var payIds []int64
	for _, g := range gr.GetGrants() {
		if g != nil {
			payIds = append(payIds, g.GetPayinProductId())
		}
	}
	var payoutIds []int64
	for _, g := range pr.GetGrants() {
		if g != nil {
			payoutIds = append(payoutIds, g.GetPayoutProductId())
		}
	}
	mi.PayinProductIds = payIds
	mi.PayoutProductIds = payoutIds
	mi.PayinGrants = nil
	mi.PayoutGrants = nil
	for _, g := range gr.GetGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayinGrant{PayinProductId: g.GetPayinProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		mi.PayinGrants = append(mi.PayinGrants, row)
	}
	for _, g := range pr.GetGrants() {
		if g == nil {
			continue
		}
		row := types.MerchantPayoutGrant{PayoutProductId: g.GetPayoutProductId()}
		if g.MerchantRateBps != nil {
			v := *g.MerchantRateBps
			row.MerchantRateBps = &v
		}
		row.FeeMode = g.GetFeeMode()
		if row.FeeMode <= 0 {
			row.FeeMode = 1
		}
		row.FeeFixedAmount = g.GetFeeFixedAmount()
		mi.PayoutGrants = append(mi.PayoutGrants, row)
	}
	return &types.AdminUpsertMerchantResp{Merchant: mi, GeneratedPassword: generatedPassword}, nil
}

func newAppID() (string, error) {
	tok, err := shared.NewToken()
	if err != nil {
		return "", err
	}
	if len(tok) > 16 {
		tok = tok[:16]
	}
	return fmt.Sprintf("app_%s", tok), nil
}

func newMerchantPassword() (string, error) {
	tok, err := shared.NewToken()
	if err != nil {
		return "", err
	}
	if len(tok) > 12 {
		tok = tok[:12]
	}
	return fmt.Sprintf("M#%s", tok), nil
}
