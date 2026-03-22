package server

import (
	"context"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/trade/internal/logic"
	"github.com/gloopai/pay/trade/internal/svc"
)

type ChannelServer struct {
	svcCtx *svc.ServiceContext
	channelpb.UnimplementedChannelServer
}

func NewChannelServer(svcCtx *svc.ServiceContext) *ChannelServer {
	return &ChannelServer{
		svcCtx: svcCtx,
	}
}

func (s *ChannelServer) Route(ctx context.Context, in *channelpb.RouteReq) (*channelpb.RouteResp, error) {
	l := logic.NewRouteLogic(ctx, s.svcCtx)
	return l.Route(in)
}

func (s *ChannelServer) GetSignSecret(ctx context.Context, in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	l := logic.NewGetSignSecretLogic(ctx, s.svcCtx)
	return l.GetSignSecret(in)
}
