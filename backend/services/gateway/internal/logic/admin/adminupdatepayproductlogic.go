package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminUpdatePayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdatePayProductLogic {
	return &AdminUpdatePayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdatePayProductLogic) AdminUpdatePayProduct(req *types.AdminUpdatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	err := l.svcCtx.PayProducts.AdminUpdatePayProduct(l.ctx, req.Id, code, name, req.SortOrder, req.Enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := l.svcCtx.PayProducts.AdminGetPayProduct(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	return &types.AdminUpsertPayProductResp{
		Product: types.AdminPayProductInfo{
			Id:        p.ID,
			Code:      p.Code,
			Name:      p.Name,
			SortOrder: p.SortOrder,
			Enabled:   p.Enabled,
		},
	}, nil
}
