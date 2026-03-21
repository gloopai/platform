package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListChannelsLogic {
	return &AdminListChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListChannelsLogic) AdminListChannels() (*types.AdminListChannelsResp, error) {
	items, err := l.svcCtx.Channels.List(l.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminChannelInfo, 0, len(items))
	for _, c := range items {
		out = append(out, types.AdminChannelInfo{
			Id:                 c.ID,
			Name:               c.Name,
			PayType:             c.PayType,
			GatewayUrl:         c.GatewayUrl,
			UpstreamMerchantNo: c.UpstreamMerchantNo,
			RsaPrivateKey:      c.RsaPrivateKey,
			SignSecret:         c.SignSecret,
			Weight:             c.Weight,
			MinAmount:          c.MinAmount,
			MaxAmount:          c.MaxAmount,
			Enabled:            c.Enabled,
			FuseEnabled:        c.FuseEnabled,
		})
	}
	return &types.AdminListChannelsResp{Channels: out}, nil
}

