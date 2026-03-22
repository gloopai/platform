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

// AdminPayProducts 管理后台支付产品定义及其与通道的绑定关系。
type AdminPayProducts struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminPayProducts(ctx context.Context, svcCtx *svc.ServiceContext) *AdminPayProducts {
	return &AdminPayProducts{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (p *AdminPayProducts) AdminListPayProducts() (*types.AdminListPayProductsResp, error) {
	rows, err := p.svcCtx.PayProducts.AdminListAllPayProducts(p.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductInfo, 0, len(rows))
	for _, row := range rows {
		out = append(out, types.AdminPayProductInfo{
			Id:        row.ID,
			Code:      row.Code,
			Name:      row.Name,
			SortOrder: row.SortOrder,
			Enabled:   row.Enabled,
		})
	}
	return &types.AdminListPayProductsResp{Products: out}, nil
}

func (p *AdminPayProducts) AdminCreatePayProduct(req *types.AdminCreatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	id, err := p.svcCtx.PayProducts.AdminCreatePayProduct(p.ctx, code, name, req.SortOrder, req.Enabled)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	row, err := p.svcCtx.PayProducts.AdminGetPayProduct(p.ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertPayProductResp{
		Product: types.AdminPayProductInfo{
			Id:        row.ID,
			Code:      row.Code,
			Name:      row.Name,
			SortOrder: row.SortOrder,
			Enabled:   row.Enabled,
		},
	}, nil
}

func (p *AdminPayProducts) AdminUpdatePayProduct(req *types.AdminUpdatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
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
	err := p.svcCtx.PayProducts.AdminUpdatePayProduct(p.ctx, req.Id, code, name, req.SortOrder, req.Enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "code already exists")
		}
		return nil, err
	}
	row, err := p.svcCtx.PayProducts.AdminGetPayProduct(p.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	return &types.AdminUpsertPayProductResp{
		Product: types.AdminPayProductInfo{
			Id:        row.ID,
			Code:      row.Code,
			Name:      row.Name,
			SortOrder: row.SortOrder,
			Enabled:   row.Enabled,
		},
	}, nil
}

func (p *AdminPayProducts) AdminListPayProductBindings(req *types.AdminListPayProductBindingsReq) (*types.AdminListPayProductBindingsResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if _, err := p.svcCtx.PayProducts.AdminGetPayProduct(p.ctx, req.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	rows, err := p.svcCtx.PayProducts.AdminListBindings(p.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductBindingInfo, 0, len(rows))
	for _, b := range rows {
		out = append(out, types.AdminPayProductBindingInfo{
			Id:           b.ID,
			PayProductId: b.PayProductID,
			ChannelId:    b.ChannelID,
			ChannelName:  b.ChannelName,
			Weight:       b.Weight,
			Enabled:      b.Enabled,
		})
	}
	return &types.AdminListPayProductBindingsResp{Bindings: out}, nil
}

func (p *AdminPayProducts) AdminUpsertPayProductBinding(req *types.AdminUpsertPayProductBindingReq) (*types.AdminUpsertPayProductBindingResp, error) {
	if req.PayProductId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if req.ChannelId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	if req.Weight <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	if _, err := p.svcCtx.PayProducts.AdminGetPayProduct(p.ctx, req.PayProductId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "pay product not found")
		}
		return nil, err
	}
	ok, err := p.svcCtx.PayProducts.AdminChannelExists(p.ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "channel not found")
	}
	bid, err := p.svcCtx.PayProducts.AdminUpsertBinding(p.ctx, req.PayProductId, req.ChannelId, req.Weight, req.Enabled)
	if err != nil {
		return nil, err
	}
	b, err := p.svcCtx.PayProducts.AdminGetBindingByID(p.ctx, bid)
	if err != nil {
		return nil, err
	}
	return &types.AdminUpsertPayProductBindingResp{
		Binding: types.AdminPayProductBindingInfo{
			Id:           b.ID,
			PayProductId: b.PayProductID,
			ChannelId:    b.ChannelID,
			ChannelName:  b.ChannelName,
			Weight:       b.Weight,
			Enabled:      b.Enabled,
		},
	}, nil
}

func (p *AdminPayProducts) AdminUpdatePayProductBinding(req *types.AdminUpdatePayProductBindingReq) (*types.AdminUpdatePayProductBindingResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if req.Weight <= 0 {
		return nil, status.Error(codes.InvalidArgument, "weight must be positive")
	}
	err := p.svcCtx.PayProducts.AdminUpdateBinding(p.ctx, req.Id, req.Weight, req.Enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	b, err := p.svcCtx.PayProducts.AdminGetBindingByID(p.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &types.AdminUpdatePayProductBindingResp{
		Binding: types.AdminPayProductBindingInfo{
			Id:           b.ID,
			PayProductId: b.PayProductID,
			ChannelId:    b.ChannelID,
			ChannelName:  b.ChannelName,
			Weight:       b.Weight,
			Enabled:      b.Enabled,
		},
	}, nil
}

func (p *AdminPayProducts) AdminDeletePayProductBinding(req *types.AdminDeletePayProductBindingReq) (*types.AdminDeletePayProductBindingResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	err := p.svcCtx.PayProducts.AdminDeleteBinding(p.ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "binding not found")
		}
		return nil, err
	}
	return &types.AdminDeletePayProductBindingResp{Ok: true}, nil
}
