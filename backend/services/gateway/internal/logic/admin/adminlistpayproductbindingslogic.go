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

type AdminListPayProductBindingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListPayProductBindingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPayProductBindingsLogic {
	return &AdminListPayProductBindingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListPayProductBindingsLogic) AdminListPayProductBindings(req *types.AdminListPayProductBindingsReq) (*types.AdminListPayProductBindingsResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if _, err := l.svcCtx.PayProducts.AdminGetPayProduct(l.ctx, req.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	rows, err := l.svcCtx.PayProducts.AdminListBindings(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductBindingInfo, 0, len(rows))
	for _, b := range rows {
		out = append(out, types.AdminPayProductBindingInfo{
			Id:            b.ID,
			PayProductId:  b.PayProductID,
			ChannelId:     b.ChannelID,
			ChannelName:   b.ChannelName,
			Weight:        b.Weight,
			Enabled:       b.Enabled,
		})
	}
	return &types.AdminListPayProductBindingsResp{Bindings: out}, nil
}
