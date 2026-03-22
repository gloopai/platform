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

type AdminUpdateChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateChannelLogic {
	return &AdminUpdateChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateChannelLogic) AdminUpdateChannel(req *types.AdminUpsertChannelReq) (*types.AdminUpsertChannelResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
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
		PayType:             req.PayType,
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
	if err := l.svcCtx.Channels.Update(l.ctx, req.Id, ch); err != nil {
		return nil, err
	}
	updated, err := l.svcCtx.Channels.GetByID(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertChannelResp{
		Channel: types.AdminChannelInfo{
			Id:                 updated.ID,
			Name:               updated.Name,
			PayType:             updated.PayType,
			GatewayUrl:         updated.GatewayUrl,
			UpstreamMerchantNo: updated.UpstreamMerchantNo,
			RsaPrivateKey:      updated.RsaPrivateKey,
			SignSecret:         updated.SignSecret,
			Weight:             updated.Weight,
			MinAmount:          updated.MinAmount,
			MaxAmount:          updated.MaxAmount,
			Enabled:            updated.Enabled,
			FuseEnabled:        updated.FuseEnabled,
		},
	}, nil
}

