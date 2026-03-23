package logic

import (
	"context"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminPayinProducts 管理后台支付产品定义及其与通道的绑定关系。
type AdminPayinProducts struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminPayinProducts(ctx context.Context, svcCtx *svc.ServiceContext) *AdminPayinProducts {
	return &AdminPayinProducts{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (p *AdminPayinProducts) AdminListPayProducts() (*types.AdminListPayProductsResp, error) {
	r, err := p.svcCtx.ChannelRpc.AdminListPayProducts(p.ctx, &channelpb.AdminListPayProductsReq{})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductInfo, 0, len(r.GetProducts()))
	for _, row := range r.GetProducts() {
		out = append(out, types.AdminPayProductInfo{
			Id:        row.GetId(),
			Code:      row.GetCode(),
			Name:      row.GetName(),
			SortOrder: row.GetSortOrder(),
			Enabled:   row.GetEnabled(),
		})
	}
	return &types.AdminListPayProductsResp{Products: out}, nil
}

func (p *AdminPayinProducts) AdminCreatePayProduct(req *types.AdminCreatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	resp, err := p.svcCtx.ChannelRpc.AdminCreatePayProduct(p.ctx, &channelpb.AdminCreatePayProductReq{
		Code: code, Name: name, SortOrder: req.SortOrder, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	pr := resp.GetProduct()
	return &types.AdminUpsertPayProductResp{
		Product: types.AdminPayProductInfo{
			Id: pr.GetId(), Code: pr.GetCode(), Name: pr.GetName(), SortOrder: pr.GetSortOrder(), Enabled: pr.GetEnabled(),
		},
	}, nil
}

func (p *AdminPayinProducts) AdminUpdatePayProduct(req *types.AdminUpdatePayProductReq) (*types.AdminUpsertPayProductResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	resp, err := p.svcCtx.ChannelRpc.AdminUpdatePayProduct(p.ctx, &channelpb.AdminUpdatePayProductReq{
		Id: req.Id, Code: code, Name: name, SortOrder: req.SortOrder, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	pr := resp.GetProduct()
	return &types.AdminUpsertPayProductResp{
		Product: types.AdminPayProductInfo{
			Id: pr.GetId(), Code: pr.GetCode(), Name: pr.GetName(), SortOrder: pr.GetSortOrder(), Enabled: pr.GetEnabled(),
		},
	}, nil
}

func (p *AdminPayinProducts) AdminListPayProductBindings(req *types.AdminListPayProductBindingsReq) (*types.AdminListPayProductBindingsResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	r, err := p.svcCtx.ChannelRpc.AdminListPayProductBindings(p.ctx, &channelpb.AdminListPayProductBindingsReq{
		PayProductId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayProductBindingInfo, 0, len(r.GetBindings()))
	for _, b := range r.GetBindings() {
		out = append(out, types.AdminPayProductBindingInfo{
			Id:           b.GetId(),
			PayProductId: b.GetPayProductId(),
			ChannelId:    b.GetChannelId(),
			ChannelName:  b.GetChannelName(),
			Weight:       b.GetWeight(),
			Enabled:      b.GetEnabled(),
		})
	}
	return &types.AdminListPayProductBindingsResp{Bindings: out}, nil
}

func (p *AdminPayinProducts) AdminUpsertPayProductBinding(req *types.AdminUpsertPayProductBindingReq) (*types.AdminUpsertPayProductBindingResp, error) {
	resp, err := p.svcCtx.ChannelRpc.AdminUpsertPayProductBinding(p.ctx, &channelpb.AdminUpsertPayProductBindingReq{
		PayProductId: req.PayProductId,
		ChannelId:    req.ChannelId,
		Weight:       req.Weight,
		Enabled:      req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	b := resp.GetBinding()
	bi := types.AdminPayProductBindingInfo{
		Id: b.GetId(), PayProductId: b.GetPayProductId(), ChannelId: b.GetChannelId(), ChannelName: b.GetChannelName(),
		Weight: b.GetWeight(), Enabled: b.GetEnabled(),
	}
	return &types.AdminUpsertPayProductBindingResp{Binding: bi}, nil
}

func (p *AdminPayinProducts) AdminUpdatePayProductBinding(req *types.AdminUpdatePayProductBindingReq) (*types.AdminUpdatePayProductBindingResp, error) {
	resp, err := p.svcCtx.ChannelRpc.AdminUpdatePayProductBinding(p.ctx, &channelpb.AdminUpdatePayProductBindingReq{
		Id: req.Id, Weight: req.Weight, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	b := resp.GetBinding()
	bi := types.AdminPayProductBindingInfo{
		Id: b.GetId(), PayProductId: b.GetPayProductId(), ChannelId: b.GetChannelId(), ChannelName: b.GetChannelName(),
		Weight: b.GetWeight(), Enabled: b.GetEnabled(),
	}
	return &types.AdminUpdatePayProductBindingResp{Binding: bi}, nil
}

func (p *AdminPayinProducts) AdminDeletePayProductBinding(req *types.AdminDeletePayProductBindingReq) (*types.AdminDeletePayProductBindingResp, error) {
	_, err := p.svcCtx.ChannelRpc.AdminDeletePayProductBinding(p.ctx, &channelpb.AdminDeletePayProductBindingReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.AdminDeletePayProductBindingResp{Ok: true}, nil
}

func (p *AdminPayinProducts) AdminListPayoutProducts() (*types.AdminListPayoutProductsResp, error) {
	r, err := p.svcCtx.ChannelRpc.AdminListPayoutProducts(p.ctx, &channelpb.AdminListPayoutProductsReq{})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayoutProductInfo, 0, len(r.GetProducts()))
	for _, row := range r.GetProducts() {
		out = append(out, types.AdminPayoutProductInfo{
			Id: row.GetId(), Code: row.GetCode(), Name: row.GetName(), SortOrder: row.GetSortOrder(), Enabled: row.GetEnabled(),
		})
	}
	return &types.AdminListPayoutProductsResp{Products: out}, nil
}

func (p *AdminPayinProducts) AdminCreatePayoutProduct(req *types.AdminCreatePayoutProductReq) (*types.AdminUpsertPayoutProductResp, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	resp, err := p.svcCtx.ChannelRpc.AdminCreatePayoutProduct(p.ctx, &channelpb.AdminCreatePayoutProductReq{
		Code: code, Name: name, SortOrder: req.SortOrder, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	pr := resp.GetProduct()
	return &types.AdminUpsertPayoutProductResp{
		Product: types.AdminPayoutProductInfo{
			Id: pr.GetId(), Code: pr.GetCode(), Name: pr.GetName(), SortOrder: pr.GetSortOrder(), Enabled: pr.GetEnabled(),
		},
	}, nil
}

func (p *AdminPayinProducts) AdminUpdatePayoutProduct(req *types.AdminUpdatePayoutProductReq) (*types.AdminUpsertPayoutProductResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name required")
	}
	resp, err := p.svcCtx.ChannelRpc.AdminUpdatePayoutProduct(p.ctx, &channelpb.AdminUpdatePayoutProductReq{
		Id: req.Id, Code: code, Name: name, SortOrder: req.SortOrder, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	pr := resp.GetProduct()
	return &types.AdminUpsertPayoutProductResp{
		Product: types.AdminPayoutProductInfo{
			Id: pr.GetId(), Code: pr.GetCode(), Name: pr.GetName(), SortOrder: pr.GetSortOrder(), Enabled: pr.GetEnabled(),
		},
	}, nil
}

func (p *AdminPayinProducts) AdminListPayoutProductBindings(req *types.AdminListPayoutProductBindingsReq) (*types.AdminListPayoutProductBindingsResp, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	r, err := p.svcCtx.ChannelRpc.AdminListPayoutProductBindings(p.ctx, &channelpb.AdminListPayoutProductBindingsReq{
		PayoutProductId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminPayoutProductBindingInfo, 0, len(r.GetBindings()))
	for _, b := range r.GetBindings() {
		out = append(out, types.AdminPayoutProductBindingInfo{
			Id: b.GetId(), PayoutProductId: b.GetPayoutProductId(), ChannelId: b.GetChannelId(), ChannelName: b.GetChannelName(),
			Weight: b.GetWeight(), Enabled: b.GetEnabled(),
		})
	}
	return &types.AdminListPayoutProductBindingsResp{Bindings: out}, nil
}

func (p *AdminPayinProducts) AdminUpsertPayoutProductBinding(req *types.AdminUpsertPayoutProductBindingReq) (*types.AdminUpsertPayoutProductBindingResp, error) {
	resp, err := p.svcCtx.ChannelRpc.AdminUpsertPayoutProductBinding(p.ctx, &channelpb.AdminUpsertPayoutProductBindingReq{
		PayoutProductId: req.PayoutProductId,
		ChannelId:       req.ChannelId,
		Weight:          req.Weight,
		Enabled:         req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	b := resp.GetBinding()
	bi := types.AdminPayoutProductBindingInfo{
		Id: b.GetId(), PayoutProductId: b.GetPayoutProductId(), ChannelId: b.GetChannelId(), ChannelName: b.GetChannelName(),
		Weight: b.GetWeight(), Enabled: b.GetEnabled(),
	}
	return &types.AdminUpsertPayoutProductBindingResp{Binding: bi}, nil
}

func (p *AdminPayinProducts) AdminUpdatePayoutProductBinding(req *types.AdminUpdatePayoutProductBindingReq) (*types.AdminUpdatePayoutProductBindingResp, error) {
	resp, err := p.svcCtx.ChannelRpc.AdminUpdatePayoutProductBinding(p.ctx, &channelpb.AdminUpdatePayoutProductBindingReq{
		Id: req.Id, Weight: req.Weight, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	b := resp.GetBinding()
	bi := types.AdminPayoutProductBindingInfo{
		Id: b.GetId(), PayoutProductId: b.GetPayoutProductId(), ChannelId: b.GetChannelId(), ChannelName: b.GetChannelName(),
		Weight: b.GetWeight(), Enabled: b.GetEnabled(),
	}
	return &types.AdminUpdatePayoutProductBindingResp{Binding: bi}, nil
}

func (p *AdminPayinProducts) AdminDeletePayoutProductBinding(req *types.AdminDeletePayoutProductBindingReq) (*types.AdminDeletePayoutProductBindingResp, error) {
	_, err := p.svcCtx.ChannelRpc.AdminDeletePayoutProductBinding(p.ctx, &channelpb.AdminDeletePayoutProductBindingReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.AdminDeletePayoutProductBindingResp{Ok: true}, nil
}
