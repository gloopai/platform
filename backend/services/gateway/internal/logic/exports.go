package logic

import (
	"context"

	adminlogic "github.com/gloopai/pay/gateway/internal/logic/admin"
	"github.com/gloopai/pay/gateway/internal/svc"
)

func NewAdminAuth(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminAuth {
	return adminlogic.NewAdminAuth(ctx, svcCtx)
}

func NewAdminSystem(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminSystem {
	return adminlogic.NewAdminSystem(ctx, svcCtx)
}

func NewAdminOps(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminOps {
	return adminlogic.NewAdminOps(ctx, svcCtx)
}

func NewAdminRbac(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRbac {
	return adminlogic.NewAdminRbac(ctx, svcCtx)
}
