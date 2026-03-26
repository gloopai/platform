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
	"golang.org/x/crypto/bcrypt"
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
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := req.Password
	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	auth, err := a.svcCtx.MerchantRpc.GetAuthInfo(a.ctx, &merchantclient.GetAuthInfoReq{Email: email})
	if err != nil || auth.GetStatus() != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(auth.GetPasswordHash()), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	merchantId := auth.GetMerchantId()
	if merchantId == "" {
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
		AppId:      auth.GetAppId(),
	}, nil
}

func (a *MerchantAuth) MerchantLogout(token string) (*types.MerchantLogoutResp, error) {
	_ = strings.TrimSpace(token)
	return &types.MerchantLogoutResp{Ok: true}, nil
}
