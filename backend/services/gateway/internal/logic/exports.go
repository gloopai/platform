package logic

import (
	"context"

	adminlogic "github.com/gloopai/pay/gateway/internal/logic/admin"
	checkoutlogic "github.com/gloopai/pay/gateway/internal/logic/checkout"
	merchantlogic "github.com/gloopai/pay/gateway/internal/logic/merchant"
	"github.com/gloopai/pay/gateway/internal/svc"
)

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *checkoutlogic.CreateOrderLogic {
	return checkoutlogic.NewCreateOrderLogic(ctx, svcCtx)
}

func NewQueryOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *checkoutlogic.QueryOrderLogic {
	return checkoutlogic.NewQueryOrderLogic(ctx, svcCtx)
}

func NewTerminalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *checkoutlogic.TerminalOrderLogic {
	return checkoutlogic.NewTerminalOrderLogic(ctx, svcCtx)
}

func NewUpstreamNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *checkoutlogic.UpstreamNotifyLogic {
	return checkoutlogic.NewUpstreamNotifyLogic(ctx, svcCtx)
}

func NewMerchantSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantSummaryLogic {
	return merchantlogic.NewMerchantSummaryLogic(ctx, svcCtx)
}

func NewMerchantOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantOrdersLogic {
	return merchantlogic.NewMerchantOrdersLogic(ctx, svcCtx)
}

func NewMerchantFundLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantFundLogsLogic {
	return merchantlogic.NewMerchantFundLogsLogic(ctx, svcCtx)
}

func NewMerchantOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantOrderDetailLogic {
	return merchantlogic.NewMerchantOrderDetailLogic(ctx, svcCtx)
}

func NewMerchantRetryNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantRetryNotifyLogic {
	return merchantlogic.NewMerchantRetryNotifyLogic(ctx, svcCtx)
}

func NewMerchantLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantLoginLogic {
	return merchantlogic.NewMerchantLoginLogic(ctx, svcCtx)
}

func NewMerchantLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *merchantlogic.MerchantLogoutLogic {
	return merchantlogic.NewMerchantLogoutLogic(ctx, svcCtx)
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminLoginLogic {
	return adminlogic.NewAdminLoginLogic(ctx, svcCtx)
}

func NewAdminLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminLogoutLogic {
	return adminlogic.NewAdminLogoutLogic(ctx, svcCtx)
}

func NewAdminListChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminListChannelsLogic {
	return adminlogic.NewAdminListChannelsLogic(ctx, svcCtx)
}

func NewAdminCreateChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminCreateChannelLogic {
	return adminlogic.NewAdminCreateChannelLogic(ctx, svcCtx)
}

func NewAdminUpdateChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminUpdateChannelLogic {
	return adminlogic.NewAdminUpdateChannelLogic(ctx, svcCtx)
}

func NewAdminListMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminListMerchantsLogic {
	return adminlogic.NewAdminListMerchantsLogic(ctx, svcCtx)
}

func NewAdminCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminCreateMerchantLogic {
	return adminlogic.NewAdminCreateMerchantLogic(ctx, svcCtx)
}

func NewAdminUpdateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminUpdateMerchantLogic {
	return adminlogic.NewAdminUpdateMerchantLogic(ctx, svcCtx)
}

func NewAdminListPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminListPayProductsLogic {
	return adminlogic.NewAdminListPayProductsLogic(ctx, svcCtx)
}

func NewAdminCreatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminCreatePayProductLogic {
	return adminlogic.NewAdminCreatePayProductLogic(ctx, svcCtx)
}

func NewAdminUpdatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminUpdatePayProductLogic {
	return adminlogic.NewAdminUpdatePayProductLogic(ctx, svcCtx)
}

func NewAdminListPayProductBindingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminListPayProductBindingsLogic {
	return adminlogic.NewAdminListPayProductBindingsLogic(ctx, svcCtx)
}

func NewAdminUpsertPayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminUpsertPayProductBindingLogic {
	return adminlogic.NewAdminUpsertPayProductBindingLogic(ctx, svcCtx)
}

func NewAdminUpdatePayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminUpdatePayProductBindingLogic {
	return adminlogic.NewAdminUpdatePayProductBindingLogic(ctx, svcCtx)
}

func NewAdminDeletePayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *adminlogic.AdminDeletePayProductBindingLogic {
	return adminlogic.NewAdminDeletePayProductBindingLogic(ctx, svcCtx)
}
