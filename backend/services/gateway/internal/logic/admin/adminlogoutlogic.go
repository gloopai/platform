package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLogoutLogic {
	return &AdminLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLogoutLogic) AdminLogout(token string) (*types.AdminLogoutResp, error) {
	tok := strings.TrimSpace(token)
	if tok == "" {
		return &types.AdminLogoutResp{Ok: true}, nil
	}
	if l.svcCtx.Config.AdminToken != "" && tok == l.svcCtx.Config.AdminToken {
		return &types.AdminLogoutResp{Ok: true}, nil
	}
	_ = l.svcCtx.Sessions.DeleteAdminSession(l.ctx, shared.TokenHash(tok))
	return &types.AdminLogoutResp{Ok: true}, nil
}
