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

type AdminDeletePayProductBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeletePayProductBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeletePayProductBindingLogic {
	return &AdminDeletePayProductBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeletePayProductBindingLogic) AdminDeletePayProductBinding(req *types.AdminDeletePayProductBindingReq) (*types.AdminDeletePayProductBindingResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	err := l.svcCtx.PayProducts.AdminDeleteBinding(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &types.AdminDeletePayProductBindingResp{Ok: true}, nil
}
