package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminCreateChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateChannelLogic {
	return &AdminCreateChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateChannelLogic) AdminCreateChannel(req *types.AdminUpsertChannelReq) (*types.AdminUpsertChannelResp, error) {
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

	ch := &store.Channel{
		Name:               req.Name,
		PayType:            req.PayType,
		GatewayUrl:         req.GatewayUrl,
		UpstreamMerchantNo: req.UpstreamMerchantNo,
		RsaPrivateKey:      req.RsaPrivateKey,
		SignSecret:         req.SignSecret,
		Weight:             req.Weight,
		MinAmount:          req.MinAmount,
		MaxAmount:          req.MaxAmount,
		Enabled:            req.Enabled,
		FuseEnabled:        req.FuseEnabled,
	}

	id, err := l.svcCtx.Channels.Create(l.ctx, ch)
	if err != nil {
		return nil, err
	}
	created, err := l.svcCtx.Channels.GetByID(l.ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{
		Channel: types.AdminChannelInfo{
			Id:                 created.ID,
			Name:               created.Name,
			PayType:            created.PayType,
			GatewayUrl:         created.GatewayUrl,
			UpstreamMerchantNo: created.UpstreamMerchantNo,
			RsaPrivateKey:      created.RsaPrivateKey,
			SignSecret:         created.SignSecret,
			Weight:             created.Weight,
			MinAmount:          created.MinAmount,
			MaxAmount:          created.MaxAmount,
			Enabled:            created.Enabled,
			FuseEnabled:        created.FuseEnabled,
		},
	}, nil
}
