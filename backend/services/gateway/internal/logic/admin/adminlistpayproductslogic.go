package logic

import (
	"context"

	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListPayProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPayProductsLogic {
	return &AdminListPayProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListPayProductsLogic) AdminListPayProducts() (*types.AdminListPayProductsResp, error) {
	rows, err := l.svcCtx.PayProducts.AdminListAllPayProducts(l.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductInfo, 0, len(rows))
	for _, p := range rows {
		out = append(out, types.AdminPayProductInfo{
			Id:        p.ID,
			Code:      p.Code,
			Name:      p.Name,
			SortOrder: p.SortOrder,
			Enabled:   p.Enabled,
		})
	}
	return &types.AdminListPayProductsResp{Products: out}, nil
}
