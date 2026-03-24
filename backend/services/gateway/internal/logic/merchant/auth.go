package logic

import (
	"context"
	"strings"
	"time"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MerchantAuth 商户控制台登录。
type MerchantAuth struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantAuth(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantAuth {
	return &MerchantAuth{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *MerchantAuth) MerchantLogin(req *types.MerchantLoginReq) (*types.MerchantLoginResp, error) {
	merchantId := strings.TrimSpace(req.MerchantId)
	secret := req.ApiSecret
	if merchantId == "" || secret == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id and api_secret required")
	}
	auth, err := a.svcCtx.MerchantRpc.GetAuthInfo(a.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
	if err != nil || auth.GetStatus() != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if auth.GetApiSecret() != secret {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	tok, expiresAt, err := shared.IssueMerchantJWT(a.svcCtx.Config.JwtSecret, merchantId, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &types.MerchantLoginResp{
		Token:      tok,
		ExpiresAt:  expiresAt.Unix(),
		MerchantId: merchantId,
	}, nil
}

func (a *MerchantAuth) MerchantLogout(token string) (*types.MerchantLogoutResp, error) {
	_ = strings.TrimSpace(token)
	return &types.MerchantLogoutResp{Ok: true}, nil
}
