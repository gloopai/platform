package server

import (
	"context"

	"github.com/gloopai/pay/channel/channel/channel"
	"github.com/gloopai/pay/channel/internal/logic"
)

func (s *ChannelServer) GetSignSecret(ctx context.Context, in *channel.GetSignSecretReq) (*channel.GetSignSecretResp, error) {
	l := logic.NewGetSignSecretLogic(ctx, s.svcCtx)
	return l.GetSignSecret(in)
}
