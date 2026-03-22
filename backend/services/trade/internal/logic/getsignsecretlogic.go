package logic

import (
	"context"
	"database/sql"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/trade/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		return nil, status.Error(codes.Internal, "query channel failed")
	}
	return &channelpb.GetSignSecretResp{SignSecret: secret}, nil
}
