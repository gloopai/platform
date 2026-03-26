package logic

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/middleware"
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

func (a *MerchantAuth) MerchantChangePassword(req *types.MerchantChangePasswordReq) (*types.MerchantChangePasswordResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(a.ctx))
	if merchantId == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant not authenticated")
	}
	oldPassword := req.OldPassword
	newPassword := req.NewPassword
	if oldPassword == "" || newPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "old_password and new_password required")
	}
	if oldPassword == newPassword {
		return nil, status.Error(codes.InvalidArgument, "new password must be different")
	}
	if err := validateMerchantPassword(newPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	auth, err := a.svcCtx.MerchantRpc.GetAuthInfo(a.ctx, &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
	if err != nil || auth.GetStatus() != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(auth.GetPasswordHash()), []byte(oldPassword)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid old password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "password hash failed")
	}
	current, err := a.svcCtx.MerchantRpc.GetMerchant(a.ctx, &merchantclient.GetMerchantReq{MerchantId: merchantId})
	if err != nil {
		return nil, err
	}
	m := current.GetMerchant()
	if m == nil {
		return nil, status.Error(codes.NotFound, "merchant not found")
	}
	if _, err := a.svcCtx.MerchantRpc.UpdateMerchant(a.ctx, &merchantclient.UpdateMerchantReq{
		MerchantId:           merchantId,
		AppSecret:            m.GetAppSecret(),
		Status:               m.GetStatus(),
		DefaultPayinRateBps:  m.GetDefaultPayinRateBps(),
		DefaultPayoutRateBps: m.GetDefaultPayoutRateBps(),
		NotifyUrl:            m.GetNotifyUrl(),
		ReturnUrl:            m.GetReturnUrl(),
		IpWhitelist:          m.GetIpWhitelist(),
		PasswordHash:         string(hash),
	}); err != nil {
		return nil, err
	}
	return &types.MerchantChangePasswordResp{Ok: true}, nil
}

func validateMerchantPassword(password string) error {
	if len(password) < 8 || len(password) > 64 {
		return fmt.Errorf("password length must be between 8 and 64")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("password must contain letters and digits")
	}
	return nil
}
