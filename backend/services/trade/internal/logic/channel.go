package logic

import (
	"context"
	"errors"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
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
	channelId, payProductID, err := l.svcCtx.Channels.Route(l.ctx, in.GetPayinType(), in.GetAmount())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &channelpb.RouteResp{ChannelId: channelId, PayinProductId: payProductID}, nil
}

type GetSignSecretLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSignSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSignSecretLogic {
	return &GetSignSecretLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSignSecretLogic) GetSignSecret(in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	if in.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	secret, err := l.svcCtx.Channels.GetSignSecret(l.ctx, in.GetChannelId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, status.Error(codes.Internal, "query channel failed")
	}
	return &channelpb.GetSignSecretResp{SignSecret: secret}, nil
}
