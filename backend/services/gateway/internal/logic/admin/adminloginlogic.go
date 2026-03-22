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

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (*types.AdminLoginResp, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	u, err := l.svcCtx.AdminUsers.FindByUsername(l.ctx, username)
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
	if err := l.svcCtx.Sessions.CreateAdminSession(l.ctx, u.ID, shared.TokenHash(tok), expiresAt); err != nil {
		return nil, err
	}
	return &types.AdminLoginResp{
		Token:     tok,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}
