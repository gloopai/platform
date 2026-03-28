package logic

import (
	"context"
	"errors"
	"strings"

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

// GetSignSecret 当前仅由网关收银台 / OpenAPI 下单链路调用（见 gateway checkout），此处可走 Consul 内存中的通道配置 JSON，避免每次查库。
// 管理台与其它逻辑请勿新增调用；若需权威数据请读库或 ListChannels/GetChannel。
func (l *GetSignSecretLogic) GetSignSecret(in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	if in.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	override := ""
	if snap, ok := l.svcCtx.ChannelSnapshot.Get(in.GetChannelId()); ok && snap != nil {
		if v := strings.TrimSpace(snap.ChannelConfig); v != "" {
			override = v
		}
	}
	secret, err := l.svcCtx.Channels.GetSignSecret(l.ctx, in.GetChannelId(), override)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, status.Error(codes.Internal, "query channel failed")
	}
	return &channelpb.GetSignSecretResp{SignSecret: secret}, nil
}
