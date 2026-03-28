package logic

import (
	"context"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminChannels 管理后台支付通道配置。
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

func toAdminChannelInfo(ch *channelpb.ChannelRow) types.AdminChannelInfo {
	if ch == nil {
		return types.AdminChannelInfo{}
	}
	return types.AdminChannelInfo{
		Id:                     ch.GetId(),
		Name:                   ch.GetName(),
		DriverKey:              ch.GetDriverKey(),
		ChannelConfig:          ch.GetChannelConfig(),
		Weight:                 ch.GetWeight(),
		MinAmount:              ch.GetMinAmount(),
		MaxAmount:              ch.GetMaxAmount(),
		SupportsPayin:          ch.GetSupportsPayin(),
		SupportsPayout:         ch.GetSupportsPayout(),
		ChannelPayinRateBps:    ch.GetChannelPayinRateBps(),
		ChannelPayoutRateBps:   ch.GetChannelPayoutRateBps(),
		ChannelPayoutFeeMode:   ch.GetChannelPayoutFeeMode(),
		ChannelPayoutFixedFee:  ch.GetChannelPayoutFixedFee(),
		Enabled:                ch.GetEnabled(),
		FuseEnabled:            ch.GetFuseEnabled(),
	}
}

func (c *AdminChannels) AdminListChannels() (*types.AdminListChannelsResp, error) {
	r, err := c.svcCtx.ChannelRpc.ListChannels(c.ctx, &channelpb.ListChannelsReq{})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminChannelInfo, 0, len(r.GetChannels()))
	for _, ch := range r.GetChannels() {
		out = append(out, toAdminChannelInfo(ch))
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

	resp, err := c.svcCtx.ChannelRpc.CreateChannel(c.ctx, &channelpb.UpsertChannelReq{
		Name:                   req.Name,
		DriverKey:              req.DriverKey,
		ChannelConfig:          req.ChannelConfig,
		Weight:                 req.Weight,
		MinAmount:              req.MinAmount,
		MaxAmount:              req.MaxAmount,
		SupportsPayin:          req.SupportsPayin,
		SupportsPayout:         req.SupportsPayout,
		ChannelPayinRateBps:    req.ChannelPayinRateBps,
		ChannelPayoutRateBps:   req.ChannelPayoutRateBps,
		ChannelPayoutFeeMode:   req.ChannelPayoutFeeMode,
		ChannelPayoutFixedFee:  req.ChannelPayoutFixedFee,
		Enabled:                req.Enabled,
		FuseEnabled:            req.FuseEnabled,
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{Channel: toAdminChannelInfo(resp.GetChannel())}, nil
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

	resp, err := c.svcCtx.ChannelRpc.UpdateChannel(c.ctx, &channelpb.UpsertChannelReq{
		Id:                     req.Id,
		Name:                   req.Name,
		DriverKey:              req.DriverKey,
		ChannelConfig:          req.ChannelConfig,
		Weight:                 req.Weight,
		MinAmount:              req.MinAmount,
		MaxAmount:              req.MaxAmount,
		SupportsPayin:          req.SupportsPayin,
		SupportsPayout:         req.SupportsPayout,
		ChannelPayinRateBps:    req.ChannelPayinRateBps,
		ChannelPayoutRateBps:   req.ChannelPayoutRateBps,
		ChannelPayoutFeeMode:   req.ChannelPayoutFeeMode,
		ChannelPayoutFixedFee:  req.ChannelPayoutFixedFee,
		Enabled:                req.Enabled,
		FuseEnabled:            req.FuseEnabled,
	})
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{Channel: toAdminChannelInfo(resp.GetChannel())}, nil
}
