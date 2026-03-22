package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminChannels 管理后台支付通道（上游）配置。
type AdminChannels struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminChannels(ctx context.Context, svcCtx *svc.ServiceContext) *AdminChannels {
	return &AdminChannels{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *AdminChannels) AdminListChannels() (*types.AdminListChannelsResp, error) {
	items, err := c.svcCtx.Channels.List(c.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminChannelInfo, 0, len(items))
	for _, ch := range items {
		out = append(out, types.AdminChannelInfo{
			Id:                 ch.ID,
			Name:               ch.Name,
			PayType:            ch.PayType,
			GatewayUrl:         ch.GatewayUrl,
			UpstreamMerchantNo: ch.UpstreamMerchantNo,
			RsaPrivateKey:      ch.RsaPrivateKey,
			SignSecret:         ch.SignSecret,
			Weight:             ch.Weight,
			MinAmount:          ch.MinAmount,
			MaxAmount:          ch.MaxAmount,
			Enabled:            ch.Enabled,
			FuseEnabled:        ch.FuseEnabled,
		})
	}
	return &types.AdminListChannelsResp{Channels: out}, nil
}

func (c *AdminChannels) AdminCreateChannel(req *types.AdminUpsertChannelReq) (*types.AdminUpsertChannelResp, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	if req.Weight < 0 || req.Weight > 100 {
		return nil, status.Error(codes.InvalidArgument, "weight must be 0-100")
	}
	if req.MinAmount < 0 || req.MaxAmount < 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be >= 0")
	}
	if req.MaxAmount > 0 && req.MinAmount > req.MaxAmount {
		return nil, status.Error(codes.InvalidArgument, "min_amount must be <= max_amount")
	}

	ch := &store.Channel{
		Name:               req.Name,
		PayType:            req.PayType,
		GatewayUrl:         req.GatewayUrl,
		UpstreamMerchantNo: req.UpstreamMerchantNo,
		RsaPrivateKey:      req.RsaPrivateKey,
		SignSecret:         req.SignSecret,
		Weight:             req.Weight,
		MinAmount:          req.MinAmount,
		MaxAmount:          req.MaxAmount,
		Enabled:            req.Enabled,
		FuseEnabled:        req.FuseEnabled,
	}

	id, err := c.svcCtx.Channels.Create(c.ctx, ch)
	if err != nil {
		return nil, err
	}
	created, err := c.svcCtx.Channels.GetByID(c.ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{
		Channel: types.AdminChannelInfo{
			Id:                 created.ID,
			Name:               created.Name,
			PayType:            created.PayType,
			GatewayUrl:         created.GatewayUrl,
			UpstreamMerchantNo: created.UpstreamMerchantNo,
			RsaPrivateKey:      created.RsaPrivateKey,
			SignSecret:         created.SignSecret,
			Weight:             created.Weight,
			MinAmount:          created.MinAmount,
			MaxAmount:          created.MaxAmount,
			Enabled:            created.Enabled,
			FuseEnabled:        created.FuseEnabled,
		},
	}, nil
}

func (c *AdminChannels) AdminUpdateChannel(req *types.AdminUpsertChannelReq) (*types.AdminUpsertChannelResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	if req.Weight < 0 || req.Weight > 100 {
		return nil, status.Error(codes.InvalidArgument, "weight must be 0-100")
	}
	if req.MinAmount < 0 || req.MaxAmount < 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be >= 0")
	}
	if req.MaxAmount > 0 && req.MinAmount > req.MaxAmount {
		return nil, status.Error(codes.InvalidArgument, "min_amount must be <= max_amount")
	}

	ch := &store.Channel{
		Name:               req.Name,
		PayType:            req.PayType,
		GatewayUrl:         req.GatewayUrl,
		UpstreamMerchantNo: req.UpstreamMerchantNo,
		RsaPrivateKey:      req.RsaPrivateKey,
		SignSecret:         req.SignSecret,
		Weight:             req.Weight,
		MinAmount:          req.MinAmount,
		MaxAmount:          req.MaxAmount,
		Enabled:            req.Enabled,
		FuseEnabled:        req.FuseEnabled,
	}
	if err := c.svcCtx.Channels.Update(c.ctx, req.Id, ch); err != nil {
		return nil, err
	}
	updated, err := c.svcCtx.Channels.GetByID(c.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{
		Channel: types.AdminChannelInfo{
			Id:                 updated.ID,
			Name:               updated.Name,
			PayType:            updated.PayType,
			GatewayUrl:         updated.GatewayUrl,
			UpstreamMerchantNo: updated.UpstreamMerchantNo,
			RsaPrivateKey:      updated.RsaPrivateKey,
			SignSecret:         updated.SignSecret,
			Weight:             updated.Weight,
			MinAmount:          updated.MinAmount,
			MaxAmount:          updated.MaxAmount,
			Enabled:            updated.Enabled,
			FuseEnabled:        updated.FuseEnabled,
		},
	}, nil
}
