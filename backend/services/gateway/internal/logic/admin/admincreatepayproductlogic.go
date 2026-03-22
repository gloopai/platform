package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminCreatePayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreatePayProductLogic {
	return &AdminCreatePayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreatePayProductLogic) AdminCreatePayProduct(req *types.AdminCreatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	id, err := l.svcCtx.PayProducts.AdminCreatePayProduct(l.ctx, code, name, req.SortOrder, req.Enabled)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	p, err := l.svcCtx.PayProducts.AdminGetPayProduct(l.ctx, id)
	if err != nil {
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
