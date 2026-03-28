package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/gloopai/pay/channeldriver"
	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/svc"
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
	if l.svcCtx.OpenAPIMemoryReady() {
		ch, pid, err := kvcache.RoutePayinFromMemory(
			in.GetPayinType(),
			in.GetAmount(),
			l.svcCtx.PayinProductSnapshot,
			l.svcCtx.PayinProductBindingsSnapshot,
			l.svcCtx.ChannelSnapshot,
		)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return &channelpb.RouteResp{ChannelId: ch, PayinProductId: pid}, nil
	}
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

// GetSignSecret 收银台 / OpenAPI 默认走 Consul 内存（authoritative_db=false）；管理侧传 authoritative_db=true 只读库。
func (l *GetSignSecretLogic) GetSignSecret(in *channelpb.GetSignSecretReq) (*channelpb.GetSignSecretResp, error) {
	if in.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	if !in.GetAuthoritativeDb() {
		if snap, ok := l.svcCtx.ChannelSnapshot.Get(in.GetChannelId()); ok && snap != nil {
			sec := strings.TrimSpace(snap.SignSecret)
			uc := strings.TrimSpace(snap.ChannelConfig)
			if uc != "" {
				_, _, js, _ := channeldriver.ConfigFieldsFromChannelJSON(uc)
				if js != "" {
					sec = js
				}
			}
			return &channelpb.GetSignSecretResp{SignSecret: sec}, nil
		}
	}
	override := ""
	secret, err := l.svcCtx.Channels.GetSignSecret(l.ctx, in.GetChannelId(), override)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, status.Error(codes.Internal, "query channel failed")
	}
	return &channelpb.GetSignSecretResp{SignSecret: secret}, nil
}
