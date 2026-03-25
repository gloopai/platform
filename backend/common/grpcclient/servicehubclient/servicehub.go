package servicehubclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/servicehub"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	FindAdminUserByUsernameReq  = servicehub.FindAdminUserByUsernameReq
	FindAdminUserByUsernameResp = servicehub.FindAdminUserByUsernameResp
	ListAdminUsersReq           = servicehub.ListAdminUsersReq
	ListAdminUsersResp          = servicehub.ListAdminUsersResp
	GetDisplaySettingsReq       = servicehub.GetDisplaySettingsReq
	GetDisplaySettingsResp      = servicehub.GetDisplaySettingsResp
	UpsertDisplaySettingsReq    = servicehub.UpsertDisplaySettingsReq
	MarkPayoutSuccessReq        = servicehub.MarkPayoutSuccessReq
	MarkPayoutFailedReq         = servicehub.MarkPayoutFailedReq
	MarkPayoutResultResp        = servicehub.MarkPayoutResultResp
	AdminUser                   = servicehub.AdminUser
	AdminUserPublic             = servicehub.AdminUserPublic
)

// ServiceHub 平台支撑数据 RPC（admin_users / global_settings / payout_orders 辅助）
type ServiceHub interface {
	FindAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]*AdminUserPublic, error)
	GetDisplaySettings(ctx context.Context) (*GetDisplaySettingsResp, error)
	UpsertDisplaySettings(ctx context.Context, country, currency, symbol string) error
	MarkPayoutSuccess(ctx context.Context, orderNo, upstreamTradeNo string) (bool, error)
	MarkPayoutFailed(ctx context.Context, orderNo string) (bool, error)
}

type defaultClient struct {
	cli servicehub.ServiceHubClient
}

func New(cli zrpc.Client) ServiceHub {
	return &defaultClient{cli: servicehub.NewServiceHubClient(cli.Conn())}
}

func (d *defaultClient) FindAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error) {
	r, err := d.cli.FindAdminUserByUsername(ctx, &servicehub.FindAdminUserByUsernameReq{Username: username})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if r == nil || r.User == nil {
		return nil, nil
	}
	return r.User, nil
}

func (d *defaultClient) ListAdminUsers(ctx context.Context) ([]*AdminUserPublic, error) {
	r, err := d.cli.ListAdminUsers(ctx, &servicehub.ListAdminUsersReq{})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Users, nil
}

func (d *defaultClient) GetDisplaySettings(ctx context.Context) (*GetDisplaySettingsResp, error) {
	return d.cli.GetDisplaySettings(ctx, &servicehub.GetDisplaySettingsReq{})
}

func (d *defaultClient) UpsertDisplaySettings(ctx context.Context, country, currency, symbol string) error {
	_, err := d.cli.UpsertDisplaySettings(ctx, &servicehub.UpsertDisplaySettingsReq{
		CountryCode:    country,
		CurrencyCode:   currency,
		CurrencySymbol: symbol,
	})
	return err
}

func (d *defaultClient) MarkPayoutSuccess(ctx context.Context, orderNo, upstreamTradeNo string) (bool, error) {
	r, err := d.cli.MarkPayoutSuccess(ctx, &servicehub.MarkPayoutSuccessReq{
		OrderNo:         orderNo,
		UpstreamTradeNo: upstreamTradeNo,
	})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Changed, nil
}

func (d *defaultClient) MarkPayoutFailed(ctx context.Context, orderNo string) (bool, error) {
	r, err := d.cli.MarkPayoutFailed(ctx, &servicehub.MarkPayoutFailedReq{OrderNo: orderNo})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Changed, nil
}
