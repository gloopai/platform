package logic

import (
	"context"
	"strings"
	"time"

	"github.com/gloopai/platform/common/jwtutil"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/gateway/internal/types"
	"github.com/pquerna/otp/totp"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminAuth 管理后台登录。
type AdminAuth struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminAuth(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAuth {
	return &AdminAuth{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminAuth) AdminLogin(req *types.AdminLoginReq) (*types.AdminLoginResp, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	u, err := a.svcCtx.ServiceHub.FindAdminUserByUsername(a.ctx, username)
	if err != nil || u == nil || u.GetStatus() != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.GetPasswordHash()), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	ds, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
	if err != nil {
		return nil, err
	}
	if ds.GetAdminMfaEnabled() == 1 && u.GetMfaEnabled() == 1 {
		code := strings.TrimSpace(req.MfaCode)
		if code == "" {
			return nil, status.Error(codes.Unauthenticated, "mfa code required")
		}
		if ok := totp.Validate(code, u.GetMfaSecret()); !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid mfa code")
		}
	}

	tok, expiresAt, err := jwtutil.IssueAdminJWT(a.svcCtx.Config.JwtSecret, u.GetId(), 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &types.AdminLoginResp{
		Token:     tok,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (a *AdminAuth) AdminLogout(token string) (*types.AdminLogoutResp, error) {
	tok := strings.TrimSpace(token)
	if tok == "" {
		return &types.AdminLogoutResp{Ok: true}, nil
	}
	if a.svcCtx.Config.AdminToken != "" && tok == a.svcCtx.Config.AdminToken {
		return &types.AdminLogoutResp{Ok: true}, nil
	}
	return &types.AdminLogoutResp{Ok: true}, nil
}
