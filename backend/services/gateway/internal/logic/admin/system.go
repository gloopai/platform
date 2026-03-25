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

// AdminSystem 系统管理（MVP：管理员账号只读列表）。
type AdminSystem struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSystem(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSystem {
	return &AdminSystem{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (a *AdminSystem) ListAdminUsers() (*types.AdminUsersResp, error) {
	rows, err := a.svcCtx.ServiceHub.ListAdminUsers(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminUserRow, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		out = append(out, types.AdminUserRow{
			ID:       r.GetId(),
			Username: r.GetUsername(),
			Status:   r.GetStatus(),
		})
	}
	return &types.AdminUsersResp{Users: out}, nil
}

func (a *AdminSystem) GetDisplaySettings(req *types.AdminDisplaySettingsReq) (*types.AdminDisplaySettingsResp, error) {
	row, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
	if err != nil {
		return nil, err
	}
	return &types.AdminDisplaySettingsResp{
		CountryCode:    row.GetCountryCode(),
		CurrencyCode:   row.GetCurrencyCode(),
		CurrencySymbol: row.GetCurrencySymbol(),
	}, nil
}

func (a *AdminSystem) UpdateDisplaySettings(req *types.AdminDisplaySettingsUpdateReq) (*types.AdminDisplaySettingsResp, error) {
	country := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	currency := strings.ToUpper(strings.TrimSpace(req.CurrencyCode))
	symbol := strings.TrimSpace(req.CurrencySymbol)
	if country == "" || currency == "" || symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "country_code, currency_code, currency_symbol required")
	}
	if err := a.svcCtx.ServiceHub.UpsertDisplaySettings(a.ctx, country, currency, symbol); err != nil {
		return nil, err
	}
	return &types.AdminDisplaySettingsResp{
		CountryCode:    country,
		CurrencyCode:   currency,
		CurrencySymbol: symbol,
	}, nil
}
