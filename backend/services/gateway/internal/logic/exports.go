package logic

import (
	"context"

	adminlogic "github.com/gloopai/pay/gateway/internal/logic/admin"
	checkoutlogic "github.com/gloopai/pay/gateway/internal/logic/checkout"
	merchantlogic "github.com/gloopai/pay/gateway/internal/logic/merchant"
	"github.com/gloopai/pay/gateway/internal/svc"
)

func NewCheckout(ctx context.Context, svcCtx *svc.ServiceContext) *checkoutlogic.Checkout {
	return checkoutlogic.NewCheckout(ctx, svcCtx)
}

func NewMerchantAuth(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantAuth {
	return merchantlogic.NewMerchantAuth(ctx, svcCtx)
}

func NewMerchantConsole(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantConsole {
	return merchantlogic.NewMerchantConsole(ctx, svcCtx)
}

func NewAdminAuth(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminAuth {
	return adminlogic.NewAdminAuth(ctx, svcCtx)
}

func NewAdminChannels(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminChannels {
	return adminlogic.NewAdminChannels(ctx, svcCtx)
}

func NewAdminMerchants(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminMerchants {
	return adminlogic.NewAdminMerchants(ctx, svcCtx)
}

func NewAdminPayinProducts(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminPayinProducts {
	return adminlogic.NewAdminPayinProducts(ctx, svcCtx)
}

func NewAdminRouting(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRouting {
	return adminlogic.NewAdminRouting(ctx, svcCtx)
}

func NewAdminStats(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminStats {
	return adminlogic.NewAdminStats(ctx, svcCtx)
}

func NewAdminOrders(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminOrders {
	return adminlogic.NewAdminOrders(ctx, svcCtx)
}

func NewAdminSystem(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminSystem {
	return adminlogic.NewAdminSystem(ctx, svcCtx)
}

func NewAdminSettlement(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminSettlement {
	return adminlogic.NewAdminSettlement(ctx, svcCtx)
}

func NewAdminRefunds(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRefunds {
	return adminlogic.NewAdminRefunds(ctx, svcCtx)
}

func NewAdminOps(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminOps {
	return adminlogic.NewAdminOps(ctx, svcCtx)
}

func NewAdminRbac(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRbac {
	return adminlogic.NewAdminRbac(ctx, svcCtx)
}
