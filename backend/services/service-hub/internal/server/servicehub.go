package server

import (
	"context"
	"errors"
	"strings"

	"github.com/gloopai/pay/common/pb/servicehub"
	"github.com/gloopai/pay/service-hub/internal/store"
	"github.com/gloopai/pay/service-hub/internal/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ServiceHubServer struct {
	servicehub.UnimplementedServiceHubServer
	svcCtx *svc.ServiceContext
}

var _ servicehub.ServiceHubServer = (*ServiceHubServer)(nil)

func NewServiceHubServer(ctx *svc.ServiceContext) *ServiceHubServer {
	return &ServiceHubServer{svcCtx: ctx}
}

func (s *ServiceHubServer) FindAdminUserByUsername(ctx context.Context, req *servicehub.FindAdminUserByUsernameReq) (*servicehub.FindAdminUserByUsernameResp, error) {
	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username required")
	}
	u, err := s.svcCtx.AdminUsers.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if u == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &servicehub.FindAdminUserByUsernameResp{
		User: &servicehub.AdminUser{
			Id:           u.ID,
			Username:     u.Username,
			PasswordHash: u.PasswordHash,
			Status:       u.Status,
		},
	}, nil
}

func (s *ServiceHubServer) ListAdminUsers(ctx context.Context, _ *servicehub.ListAdminUsersReq) (*servicehub.ListAdminUsersResp, error) {
	rows, err := s.svcCtx.AdminUsers.List(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminUserPublic, 0, len(rows))
	for _, r := range rows {
		out = append(out, &servicehub.AdminUserPublic{
			Id:       r.ID,
			Username: r.Username,
			Status:   r.Status,
		})
	}
	return &servicehub.ListAdminUsersResp{Users: out}, nil
}

func (s *ServiceHubServer) GetDisplaySettings(ctx context.Context, _ *servicehub.GetDisplaySettingsReq) (*servicehub.GetDisplaySettingsResp, error) {
	row, err := s.svcCtx.GlobalSettings.GetDisplaySettings(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.GetDisplaySettingsResp{
		CountryCode:    row.CountryCode,
		CurrencyCode:   row.CurrencyCode,
		CurrencySymbol: row.CurrencySymbol,
	}, nil
}

func (s *ServiceHubServer) UpsertDisplaySettings(ctx context.Context, req *servicehub.UpsertDisplaySettingsReq) (*servicehub.GetDisplaySettingsResp, error) {
	country := strings.ToUpper(strings.TrimSpace(req.GetCountryCode()))
	currency := strings.ToUpper(strings.TrimSpace(req.GetCurrencyCode()))
	symbol := strings.TrimSpace(req.GetCurrencySymbol())
	if country == "" || currency == "" || symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "country_code, currency_code, currency_symbol required")
	}
	if err := s.svcCtx.GlobalSettings.UpsertDisplaySettings(ctx, &store.GlobalDisplaySettings{
		CountryCode:    country,
		CurrencyCode:   currency,
		CurrencySymbol: symbol,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.GetDisplaySettingsResp{
		CountryCode:    country,
		CurrencyCode:   currency,
		CurrencySymbol: symbol,
	}, nil
}

func (s *ServiceHubServer) MarkPayoutSuccess(ctx context.Context, req *servicehub.MarkPayoutSuccessReq) (*servicehub.MarkPayoutResultResp, error) {
	orderNo := strings.TrimSpace(req.GetOrderNo())
	upstream := strings.TrimSpace(req.GetUpstreamTradeNo())
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	changed, err := s.svcCtx.PayoutOrders.MarkSuccess(ctx, orderNo, upstream)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.MarkPayoutResultResp{Changed: changed}, nil
}

func (s *ServiceHubServer) MarkPayoutFailed(ctx context.Context, req *servicehub.MarkPayoutFailedReq) (*servicehub.MarkPayoutResultResp, error) {
	orderNo := strings.TrimSpace(req.GetOrderNo())
	if orderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	changed, err := s.svcCtx.PayoutOrders.MarkFailed(ctx, orderNo)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.MarkPayoutResultResp{Changed: changed}, nil
}
