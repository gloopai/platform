package server

import (
	"context"

	"github.com/gloopai/pay/channel/internal/logic"
	channelpb "github.com/gloopai/pay/common/pb/channel"
)

func (s *ChannelServer) GetSignSecret(ctx context.Context, in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	l := logic.NewGetSignSecretLogic(ctx, s.svcCtx)
	return l.GetSignSecret(in)
}
