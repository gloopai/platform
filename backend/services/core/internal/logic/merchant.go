package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"

	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
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

	if len(in.GetPayProductIds()) > 0 {
		if err := l.svcCtx.MerchantPayProducts.Replace(l.ctx, merchantId, in.GetPayProductIds()); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}

	created, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "merchant create failed")
		}
		return nil, err
	}
	ids, err := l.svcCtx.MerchantPayProducts.ListProductIDs(l.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(created, ids)}, nil
}

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
	ids, err := l.svcCtx.MerchantPayProducts.ListProductIDs(l.ctx, merchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, "load merchant pay products failed")
	}
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(updated, ids)}, nil
}

type GetMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMerchantLogic {
	return &GetMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMerchantLogic) GetMerchant(in *merchantpb.GetMerchantReq) (*merchantpb.GetMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	m, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}
	ids, _ := l.svcCtx.MerchantPayProducts.ListProductIDs(l.ctx, merchantId)
	return &merchantpb.GetMerchantResp{
		Merchant: toMerchantInfo(m, ids),
	}, nil
}

type GetAuthInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuthInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthInfoLogic {
	return &GetAuthInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuthInfoLogic) GetAuthInfo(in *merchantpb.GetAuthInfoReq) (*merchantpb.GetAuthInfoResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	m, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}
	return &merchantpb.GetAuthInfoResp{
		ApiSecret:   m.ApiSecret,
		Status:      m.Status,
		IpWhitelist: m.IpWhitelist,
		NotifyUrl:   m.NotifyUrl,
		ReturnUrl:   m.ReturnUrl,
		Balance:     m.Balance,
		RateBps:     m.RateBps,
	}, nil
}

type ListMerchantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantsLogic {
	return &ListMerchantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMerchantsLogic) ListMerchants(in *merchantpb.ListMerchantsReq) (*merchantpb.ListMerchantsResp, error) {
	items, err := l.svcCtx.Merchants.List(l.ctx, in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*merchantpb.MerchantInfo, 0, len(items))
	for i := range items {
		m := items[i]
		ids, err := l.svcCtx.MerchantPayProducts.ListProductIDs(l.ctx, m.MerchantId)
		if err != nil {
			return nil, status.Error(codes.Internal, "load merchant pay products failed")
		}
		out = append(out, toMerchantInfo(&m, ids))
	}
	return &merchantpb.ListMerchantsResp{Merchants: out}, nil
}

func newSecret() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func toMerchantInfo(m *store.Merchant, payProductIds []int64) *merchantpb.MerchantInfo {
	if m == nil {
		return nil
	}
	return &merchantpb.MerchantInfo{
		MerchantId:      m.MerchantId,
		ApiSecret:       m.ApiSecret,
		Status:          m.Status,
		RateBps:         m.RateBps,
		IpWhitelist:     m.IpWhitelist,
		Balance:         m.Balance,
		FrozenBalance:   m.FrozenBalance,
		WithdrawnAmount: m.WithdrawnAmount,
		NotifyUrl:       m.NotifyUrl,
		ReturnUrl:       m.ReturnUrl,
		PayProductIds:   payProductIds,
	}
}
