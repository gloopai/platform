package logic

import (
	"context"

	adminlogic "github.com/gloopai/platform/gateway/internal/logic/admin"
	"github.com/gloopai/platform/gateway/internal/svc"
)

func NewAdminAuth(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminAuth {
	return adminlogic.NewAdminAuth(ctx, svcCtx)
}

func NewAdminSystem(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminSystem {
	return adminlogic.NewAdminSystem(ctx, svcCtx)
}

func NewAdminRbac(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRbac {
	return adminlogic.NewAdminRbac(ctx, svcCtx)
}

func NewAdminJobs(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminJobs {
	return adminlogic.NewAdminJobs(ctx, svcCtx)
}
