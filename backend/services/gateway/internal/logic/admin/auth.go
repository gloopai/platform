package logic

import (
	"context"
	"strings"
	"time"

	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminAuth 管理后台登录与会话。
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

	u, err := a.svcCtx.AdminUsers.FindByUsername(a.ctx, username)
	if err != nil || u == nil || u.Status != 1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	tok, err := shared.NewToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := a.svcCtx.Sessions.CreateAdminSession(a.ctx, u.ID, shared.TokenHash(tok), expiresAt); err != nil {
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
	_ = a.svcCtx.Sessions.DeleteAdminSession(a.ctx, shared.TokenHash(tok))
	return &types.AdminLogoutResp{Ok: true}, nil
}
