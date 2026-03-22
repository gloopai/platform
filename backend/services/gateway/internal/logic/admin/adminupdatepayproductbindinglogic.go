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

type AdminUpdatePayProductBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdatePayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdatePayProductBindingLogic {
	return &AdminUpdatePayProductBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdatePayProductBindingLogic) AdminUpdatePayProductBinding(req *types.AdminUpdatePayProductBindingReq) (*types.AdminUpdatePayProductBindingResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if req.Weight <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	err := l.svcCtx.PayProducts.AdminUpdateBinding(l.ctx, req.Id, req.Weight, req.Enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	b, err := l.svcCtx.PayProducts.AdminGetBindingByID(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &types.AdminUpdatePayProductBindingResp{
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
