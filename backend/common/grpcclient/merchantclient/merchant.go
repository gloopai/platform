package merchantclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/merchant"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	GetMerchantReq     = merchant.GetMerchantReq
	GetMerchantResp    = merchant.GetMerchantResp
	GetAuthInfoReq     = merchant.GetAuthInfoReq
	GetAuthInfoResp    = merchant.GetAuthInfoResp
	ListMerchantsReq   = merchant.ListMerchantsReq
	ListMerchantsResp  = merchant.ListMerchantsResp
	CreateMerchantReq  = merchant.CreateMerchantReq
	UpdateMerchantReq  = merchant.UpdateMerchantReq
	UpsertMerchantResp = merchant.UpsertMerchantResp
	MerchantInfo       = merchant.MerchantInfo

	Merchant interface {
		GetMerchant(ctx context.Context, in *GetMerchantReq, opts ...grpc.CallOption) (*GetMerchantResp, error)
		GetAuthInfo(ctx context.Context, in *GetAuthInfoReq, opts ...grpc.CallOption) (*GetAuthInfoResp, error)
		ListMerchants(ctx context.Context, in *ListMerchantsReq, opts ...grpc.CallOption) (*ListMerchantsResp, error)
		CreateMerchant(ctx context.Context, in *CreateMerchantReq, opts ...grpc.CallOption) (*UpsertMerchantResp, error)
		UpdateMerchant(ctx context.Context, in *UpdateMerchantReq, opts ...grpc.CallOption) (*UpsertMerchantResp, error)
		ReplaceMerchantPayProducts(ctx context.Context, in *merchant.ReplaceMerchantPayProductsReq, opts ...grpc.CallOption) (*merchant.ReplaceMerchantPayProductsResp, error)
		ListMerchantPayProductIds(ctx context.Context, in *merchant.ListMerchantPayProductIdsReq, opts ...grpc.CallOption) (*merchant.ListMerchantPayProductIdsResp, error)
	}

	defaultMerchant struct {
		cli zrpc.Client
	}
)

func NewMerchant(cli zrpc.Client) Merchant {
	return &defaultMerchant{cli: cli}
}

func (m *defaultMerchant) client() merchant.MerchantClient {
	return merchant.NewMerchantClient(m.cli.Conn())
}

func (m *defaultMerchant) GetMerchant(ctx context.Context, in *GetMerchantReq, opts ...grpc.CallOption) (*GetMerchantResp, error) {
	return m.client().GetMerchant(ctx, in, opts...)
}

func (m *defaultMerchant) GetAuthInfo(ctx context.Context, in *GetAuthInfoReq, opts ...grpc.CallOption) (*GetAuthInfoResp, error) {
	return m.client().GetAuthInfo(ctx, in, opts...)
}

func (m *defaultMerchant) ListMerchants(ctx context.Context, in *ListMerchantsReq, opts ...grpc.CallOption) (*ListMerchantsResp, error) {
	return m.client().ListMerchants(ctx, in, opts...)
}

func (m *defaultMerchant) CreateMerchant(ctx context.Context, in *CreateMerchantReq, opts ...grpc.CallOption) (*UpsertMerchantResp, error) {
	return m.client().CreateMerchant(ctx, in, opts...)
}

func (m *defaultMerchant) UpdateMerchant(ctx context.Context, in *UpdateMerchantReq, opts ...grpc.CallOption) (*UpsertMerchantResp, error) {
	return m.client().UpdateMerchant(ctx, in, opts...)
}

func (m *defaultMerchant) ReplaceMerchantPayProducts(ctx context.Context, in *merchant.ReplaceMerchantPayProductsReq, opts ...grpc.CallOption) (*merchant.ReplaceMerchantPayProductsResp, error) {
	return m.client().ReplaceMerchantPayProducts(ctx, in, opts...)
}

func (m *defaultMerchant) ListMerchantPayProductIds(ctx context.Context, in *merchant.ListMerchantPayProductIdsReq, opts ...grpc.CallOption) (*merchant.ListMerchantPayProductIdsResp, error) {
	return m.client().ListMerchantPayProductIds(ctx, in, opts...)
}
