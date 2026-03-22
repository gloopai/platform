package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminUpsertPayProductBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpsertPayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpsertPayProductBindingLogic {
	return &AdminUpsertPayProductBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpsertPayProductBindingLogic) AdminUpsertPayProductBinding(req *types.AdminUpsertPayProductBindingReq) (*types.AdminUpsertPayProductBindingResp, error) {
	if req.PayProductId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if req.ChannelId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	if req.Weight <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	if _, err := l.svcCtx.PayProducts.AdminGetPayProduct(l.ctx, req.PayProductId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	ok, err := l.svcCtx.PayProducts.AdminChannelExists(l.ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "channel not found")
	}
	bid, err := l.svcCtx.PayProducts.AdminUpsertBinding(l.ctx, req.PayProductId, req.ChannelId, req.Weight, req.Enabled)
	if err != nil {
		return nil, err
	}
	b, err := l.svcCtx.PayProducts.AdminGetBindingByID(l.ctx, bid)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertPayProductBindingResp{
		Binding: types.AdminPayProductBindingInfo{
			Id:            b.ID,
			PayProductId:  b.PayProductID,
			ChannelId:     b.ChannelID,
			ChannelName:   b.ChannelName,
			Weight:        b.Weight,
			Enabled:       b.Enabled,
		},
	}, nil
}
