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

func NewAdminPayProducts(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminPayProducts {
	return adminlogic.NewAdminPayProducts(ctx, svcCtx)
}

func NewAdminRouting(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminRouting {
	return adminlogic.NewAdminRouting(ctx, svcCtx)
}

func NewAdminStats(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminStats {
	return adminlogic.NewAdminStats(ctx, svcCtx)
}
