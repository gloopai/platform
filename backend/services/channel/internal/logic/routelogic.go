package logic

import (
	"context"

	"github.com/gloopai/pay/channel/internal/svc"
	channelpb "github.com/gloopai/pay/common/pb/channel"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RouteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRouteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RouteLogic {
	return &RouteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RouteLogic) Route(in *channelpb.RouteReq) (*channelpb.RouteResp, error) {
	channelId, err := l.svcCtx.Store.Route(l.ctx, in.GetPayType(), in.GetAmount())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &channelpb.RouteResp{ChannelId: channelId}, nil
}
