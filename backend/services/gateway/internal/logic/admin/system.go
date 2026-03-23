package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminSystem 系统管理（MVP：管理员账号只读列表）。
type AdminSystem struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSystem(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSystem {
	return &AdminSystem{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminSystem) ListAdminUsers() (*types.AdminUsersResp, error) {
	rows, err := a.svcCtx.AdminUsers.List(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminUserRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.AdminUserRow{
			ID:       r.ID,
			Username: r.Username,
			Status:   r.Status,
		})
	}
	return &types.AdminUsersResp{Users: out}, nil
}
