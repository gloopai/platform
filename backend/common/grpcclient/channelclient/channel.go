package channelclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/channel"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	RouteReq          = channel.RouteReq
	RouteResp         = channel.RouteResp
	GetSignSecretReq  = channel.GetSignSecretReq
	GetSignSecretResp = channel.GetSignSecretResp

	Channel interface {
		Route(ctx context.Context, in *RouteReq, opts ...grpc.CallOption) (*RouteResp, error)
		GetSignSecret(ctx context.Context, in *GetSignSecretReq, opts ...grpc.CallOption) (*GetSignSecretResp, error)

		ListChannels(ctx context.Context, in *channel.ListChannelsReq, opts ...grpc.CallOption) (*channel.ListChannelsResp, error)
		CreateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error)
		UpdateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error)
		GetRoutingSummary(ctx context.Context, in *channel.GetRoutingSummaryReq, opts ...grpc.CallOption) (*channel.GetRoutingSummaryResp, error)

		ListTerminalPayProducts(ctx context.Context, in *channel.ListTerminalPayProductsReq, opts ...grpc.CallOption) (*channel.ListTerminalPayProductsResp, error)
		MerchantHasPayProductCode(ctx context.Context, in *channel.MerchantHasPayProductCodeReq, opts ...grpc.CallOption) (*channel.MerchantHasPayProductCodeResp, error)
		ResolveLockedChannelForMerchant(ctx context.Context, in *channel.ResolveLockedChannelForMerchantReq, opts ...grpc.CallOption) (*channel.ResolveLockedChannelForMerchantResp, error)
		GetPayProductDisplayName(ctx context.Context, in *channel.GetPayProductDisplayNameReq, opts ...grpc.CallOption) (*channel.GetPayProductDisplayNameResp, error)

		AdminListPayProducts(ctx context.Context, in *channel.AdminListPayProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayProductsResp, error)
		AdminCreatePayProduct(ctx context.Context, in *channel.AdminCreatePayProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductResp, error)
		AdminUpdatePayProduct(ctx context.Context, in *channel.AdminUpdatePayProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductResp, error)
		AdminListPayProductBindings(ctx context.Context, in *channel.AdminListPayProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayProductBindingsResp, error)
		AdminUpsertPayProductBinding(ctx context.Context, in *channel.AdminUpsertPayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductBindingResp, error)
		AdminUpdatePayProductBinding(ctx context.Context, in *channel.AdminUpdatePayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayProductBindingResp, error)
		AdminDeletePayProductBinding(ctx context.Context, in *channel.AdminDeletePayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayProductBindingResp, error)

		AdminListPayoutProducts(ctx context.Context, in *channel.AdminListPayoutProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayoutProductsResp, error)
		AdminCreatePayoutProduct(ctx context.Context, in *channel.AdminCreatePayoutProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductResp, error)
		AdminUpdatePayoutProduct(ctx context.Context, in *channel.AdminUpdatePayoutProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductResp, error)
		AdminListPayoutProductBindings(ctx context.Context, in *channel.AdminListPayoutProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayoutProductBindingsResp, error)
		AdminUpsertPayoutProductBinding(ctx context.Context, in *channel.AdminUpsertPayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductBindingResp, error)
		AdminUpdatePayoutProductBinding(ctx context.Context, in *channel.AdminUpdatePayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayoutProductBindingResp, error)
		AdminDeletePayoutProductBinding(ctx context.Context, in *channel.AdminDeletePayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayoutProductBindingResp, error)
	}

	defaultChannel struct {
		cli zrpc.Client
	}
)

func NewChannel(cli zrpc.Client) Channel {
	return &defaultChannel{cli: cli}
}

func (m *defaultChannel) client() channel.ChannelClient {
	return channel.NewChannelClient(m.cli.Conn())
}

func (m *defaultChannel) Route(ctx context.Context, in *RouteReq, opts ...grpc.CallOption) (*RouteResp, error) {
	return m.client().Route(ctx, in, opts...)
}

func (m *defaultChannel) GetSignSecret(ctx context.Context, in *GetSignSecretReq, opts ...grpc.CallOption) (*GetSignSecretResp, error) {
	return m.client().GetSignSecret(ctx, in, opts...)
}

func (m *defaultChannel) ListChannels(ctx context.Context, in *channel.ListChannelsReq, opts ...grpc.CallOption) (*channel.ListChannelsResp, error) {
	return m.client().ListChannels(ctx, in, opts...)
}

func (m *defaultChannel) CreateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error) {
	return m.client().CreateChannel(ctx, in, opts...)
}

func (m *defaultChannel) UpdateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error) {
	return m.client().UpdateChannel(ctx, in, opts...)
}

func (m *defaultChannel) GetRoutingSummary(ctx context.Context, in *channel.GetRoutingSummaryReq, opts ...grpc.CallOption) (*channel.GetRoutingSummaryResp, error) {
	return m.client().GetRoutingSummary(ctx, in, opts...)
}

func (m *defaultChannel) ListTerminalPayProducts(ctx context.Context, in *channel.ListTerminalPayProductsReq, opts ...grpc.CallOption) (*channel.ListTerminalPayProductsResp, error) {
	return m.client().ListTerminalPayProducts(ctx, in, opts...)
}

func (m *defaultChannel) MerchantHasPayProductCode(ctx context.Context, in *channel.MerchantHasPayProductCodeReq, opts ...grpc.CallOption) (*channel.MerchantHasPayProductCodeResp, error) {
	return m.client().MerchantHasPayProductCode(ctx, in, opts...)
}

func (m *defaultChannel) ResolveLockedChannelForMerchant(ctx context.Context, in *channel.ResolveLockedChannelForMerchantReq, opts ...grpc.CallOption) (*channel.ResolveLockedChannelForMerchantResp, error) {
	return m.client().ResolveLockedChannelForMerchant(ctx, in, opts...)
}

func (m *defaultChannel) GetPayProductDisplayName(ctx context.Context, in *channel.GetPayProductDisplayNameReq, opts ...grpc.CallOption) (*channel.GetPayProductDisplayNameResp, error) {
	return m.client().GetPayProductDisplayName(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayProducts(ctx context.Context, in *channel.AdminListPayProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayProductsResp, error) {
	return m.client().AdminListPayProducts(ctx, in, opts...)
}

func (m *defaultChannel) AdminCreatePayProduct(ctx context.Context, in *channel.AdminCreatePayProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductResp, error) {
	return m.client().AdminCreatePayProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayProduct(ctx context.Context, in *channel.AdminUpdatePayProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductResp, error) {
	return m.client().AdminUpdatePayProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayProductBindings(ctx context.Context, in *channel.AdminListPayProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayProductBindingsResp, error) {
	return m.client().AdminListPayProductBindings(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpsertPayProductBinding(ctx context.Context, in *channel.AdminUpsertPayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayProductBindingResp, error) {
	return m.client().AdminUpsertPayProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayProductBinding(ctx context.Context, in *channel.AdminUpdatePayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayProductBindingResp, error) {
	return m.client().AdminUpdatePayProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminDeletePayProductBinding(ctx context.Context, in *channel.AdminDeletePayProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayProductBindingResp, error) {
	return m.client().AdminDeletePayProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayoutProducts(ctx context.Context, in *channel.AdminListPayoutProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayoutProductsResp, error) {
	return m.client().AdminListPayoutProducts(ctx, in, opts...)
}

func (m *defaultChannel) AdminCreatePayoutProduct(ctx context.Context, in *channel.AdminCreatePayoutProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductResp, error) {
	return m.client().AdminCreatePayoutProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayoutProduct(ctx context.Context, in *channel.AdminUpdatePayoutProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductResp, error) {
	return m.client().AdminUpdatePayoutProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayoutProductBindings(ctx context.Context, in *channel.AdminListPayoutProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayoutProductBindingsResp, error) {
	return m.client().AdminListPayoutProductBindings(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpsertPayoutProductBinding(ctx context.Context, in *channel.AdminUpsertPayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayoutProductBindingResp, error) {
	return m.client().AdminUpsertPayoutProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayoutProductBinding(ctx context.Context, in *channel.AdminUpdatePayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayoutProductBindingResp, error) {
	return m.client().AdminUpdatePayoutProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminDeletePayoutProductBinding(ctx context.Context, in *channel.AdminDeletePayoutProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayoutProductBindingResp, error) {
	return m.client().AdminDeletePayoutProductBinding(ctx, in, opts...)
}
