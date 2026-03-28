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

		GetChannel(ctx context.Context, in *channel.GetChannelReq, opts ...grpc.CallOption) (*channel.GetChannelResp, error)
		ListChannels(ctx context.Context, in *channel.ListChannelsReq, opts ...grpc.CallOption) (*channel.ListChannelsResp, error)
		CreateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error)
		UpdateChannel(ctx context.Context, in *channel.UpsertChannelReq, opts ...grpc.CallOption) (*channel.UpsertChannelResp, error)
		GetRoutingSummary(ctx context.Context, in *channel.GetRoutingSummaryReq, opts ...grpc.CallOption) (*channel.GetRoutingSummaryResp, error)

		ListTerminalPayinProducts(ctx context.Context, in *channel.ListTerminalPayinProductsReq, opts ...grpc.CallOption) (*channel.ListTerminalPayinProductsResp, error)
		MerchantHasPayinProductCode(ctx context.Context, in *channel.MerchantHasPayinProductCodeReq, opts ...grpc.CallOption) (*channel.MerchantHasPayinProductCodeResp, error)
		ResolveLockedChannelForMerchant(ctx context.Context, in *channel.ResolveLockedChannelForMerchantReq, opts ...grpc.CallOption) (*channel.ResolveLockedChannelForMerchantResp, error)
		GetPayinProductDisplayName(ctx context.Context, in *channel.GetPayinProductDisplayNameReq, opts ...grpc.CallOption) (*channel.GetPayinProductDisplayNameResp, error)

		AdminListPayinProducts(ctx context.Context, in *channel.AdminListPayinProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayinProductsResp, error)
		AdminCreatePayinProduct(ctx context.Context, in *channel.AdminCreatePayinProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductResp, error)
		AdminUpdatePayinProduct(ctx context.Context, in *channel.AdminUpdatePayinProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductResp, error)
		AdminListPayinProductBindings(ctx context.Context, in *channel.AdminListPayinProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayinProductBindingsResp, error)
		AdminUpsertPayinProductBinding(ctx context.Context, in *channel.AdminUpsertPayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductBindingResp, error)
		AdminUpdatePayinProductBinding(ctx context.Context, in *channel.AdminUpdatePayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayinProductBindingResp, error)
		AdminDeletePayinProductBinding(ctx context.Context, in *channel.AdminDeletePayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayinProductBindingResp, error)

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

func (m *defaultChannel) GetChannel(ctx context.Context, in *channel.GetChannelReq, opts ...grpc.CallOption) (*channel.GetChannelResp, error) {
	return m.client().GetChannel(ctx, in, opts...)
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

func (m *defaultChannel) ListTerminalPayinProducts(ctx context.Context, in *channel.ListTerminalPayinProductsReq, opts ...grpc.CallOption) (*channel.ListTerminalPayinProductsResp, error) {
	return m.client().ListTerminalPayinProducts(ctx, in, opts...)
}

func (m *defaultChannel) MerchantHasPayinProductCode(ctx context.Context, in *channel.MerchantHasPayinProductCodeReq, opts ...grpc.CallOption) (*channel.MerchantHasPayinProductCodeResp, error) {
	return m.client().MerchantHasPayinProductCode(ctx, in, opts...)
}

func (m *defaultChannel) ResolveLockedChannelForMerchant(ctx context.Context, in *channel.ResolveLockedChannelForMerchantReq, opts ...grpc.CallOption) (*channel.ResolveLockedChannelForMerchantResp, error) {
	return m.client().ResolveLockedChannelForMerchant(ctx, in, opts...)
}

func (m *defaultChannel) GetPayinProductDisplayName(ctx context.Context, in *channel.GetPayinProductDisplayNameReq, opts ...grpc.CallOption) (*channel.GetPayinProductDisplayNameResp, error) {
	return m.client().GetPayinProductDisplayName(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayinProducts(ctx context.Context, in *channel.AdminListPayinProductsReq, opts ...grpc.CallOption) (*channel.AdminListPayinProductsResp, error) {
	return m.client().AdminListPayinProducts(ctx, in, opts...)
}

func (m *defaultChannel) AdminCreatePayinProduct(ctx context.Context, in *channel.AdminCreatePayinProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductResp, error) {
	return m.client().AdminCreatePayinProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayinProduct(ctx context.Context, in *channel.AdminUpdatePayinProductReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductResp, error) {
	return m.client().AdminUpdatePayinProduct(ctx, in, opts...)
}

func (m *defaultChannel) AdminListPayinProductBindings(ctx context.Context, in *channel.AdminListPayinProductBindingsReq, opts ...grpc.CallOption) (*channel.AdminListPayinProductBindingsResp, error) {
	return m.client().AdminListPayinProductBindings(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpsertPayinProductBinding(ctx context.Context, in *channel.AdminUpsertPayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpsertPayinProductBindingResp, error) {
	return m.client().AdminUpsertPayinProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminUpdatePayinProductBinding(ctx context.Context, in *channel.AdminUpdatePayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminUpdatePayinProductBindingResp, error) {
	return m.client().AdminUpdatePayinProductBinding(ctx, in, opts...)
}

func (m *defaultChannel) AdminDeletePayinProductBinding(ctx context.Context, in *channel.AdminDeletePayinProductBindingReq, opts ...grpc.CallOption) (*channel.AdminDeletePayinProductBindingResp, error) {
	return m.client().AdminDeletePayinProductBinding(ctx, in, opts...)
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
