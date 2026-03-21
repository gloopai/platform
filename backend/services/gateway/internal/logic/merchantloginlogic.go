package logic

import (
	"context"
	"strings"
	"time"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/gloopai/pay/merchant/merchantclient"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantLoginLogic {
	return &MerchantLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantLoginLogic) MerchantLogin(req *types.MerchantLoginReq) (*types.MerchantLoginResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	secret := req.ApiSecret
	if merchantId == "" || secret == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and api_secret required")
	}
	auth, err := l.svcCtx.MerchantRpc.GetAuthInfo(l.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
	if err != nil || auth.GetStatus() != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if auth.GetApiSecret() != secret {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	tok, err := newToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := l.svcCtx.Sessions.CreateMerchantSession(l.ctx, merchantId, tokenHash(tok), expiresAt); err != nil {
		return nil, err
	}
	return &types.MerchantLoginResp{
		Token:      tok,
		ExpiresAt:  expiresAt.Unix(),
		MerchantId: merchantId,
	}, nil
}
