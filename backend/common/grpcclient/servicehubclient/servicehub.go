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
	PublishPortalNotificationReq  = servicehub.PublishPortalNotificationReq
	PublishPortalNotificationResp = servicehub.PublishPortalNotificationResp
	AdminUser                   = servicehub.AdminUser
	AdminUserPublic             = servicehub.AdminUserPublic

	AdminMenu       = servicehub.AdminMenu
	AdminRole       = servicehub.AdminRole
	AdminPermission = servicehub.AdminPermission
	AdminApiRule    = servicehub.AdminApiRule
)

// ServiceHub 平台支撑数据 RPC（admin_users / global_settings / payout_orders 辅助）
type ServiceHub interface {
	FindAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]*AdminUserPublic, error)
	GetDisplaySettings(ctx context.Context) (*GetDisplaySettingsResp, error)
	UpsertDisplaySettings(ctx context.Context, country, currency, symbol string, fiatToUsdtRate float64, adminMfaEnabled int64, merchantNumericIDStart int64) error
	MarkPayoutSuccess(ctx context.Context, orderNo, upstreamTradeNo string) (bool, error)
	MarkPayoutFailed(ctx context.Context, orderNo string) (bool, error)

	CreateAdminUser(ctx context.Context, username, passwordHash string, status int64) (*AdminUserPublic, error)
	UpdateAdminUser(ctx context.Context, id int64, status int64, passwordHash, mfaSecret *string, mfaEnabled *int64) (*AdminUserPublic, error)
	DeleteAdminUser(ctx context.Context, id int64) (bool, error)
	GetAdminUserById(ctx context.Context, id int64) (*AdminUser, error)

	// RBAC
	GetAdminRbacMyMenus(ctx context.Context, adminUserID int64) ([]*AdminMenu, error)
	ListAdminRoles(ctx context.Context) ([]*AdminRole, error)
	CreateAdminRole(ctx context.Context, code, name string, status int64) (*AdminRole, error)
	UpdateAdminRole(ctx context.Context, id int64, name string, status int64) (*AdminRole, error)
	DeleteAdminRole(ctx context.Context, id int64) (bool, error)
	ListAdminMenus(ctx context.Context) ([]*AdminMenu, error)
	CreateAdminMenu(ctx context.Context, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error)
	UpdateAdminMenu(ctx context.Context, id int64, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error)
	DeleteAdminMenu(ctx context.Context, id int64) (bool, error)
	GetAdminRoleMenus(ctx context.Context, roleID int64) ([]int64, error)
	SetAdminRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) (bool, error)
	GetAdminUserRoles(ctx context.Context, adminUserID int64) ([]int64, error)
	SetAdminUserRoles(ctx context.Context, adminUserID int64, roleIDs []int64) (bool, error)

	// permissions
	GetAdminRbacMyPerms(ctx context.Context, adminUserID int64) (isSuper bool, permKeys []string, err error)

	// config
	ListAdminPermissions(ctx context.Context) ([]*AdminPermission, error)
	CreateAdminPermission(ctx context.Context, permKey, label, category, menuKey string, status int64) (*AdminPermission, error)
	UpdateAdminPermission(ctx context.Context, id int64, label, category, menuKey string, status int64) (*AdminPermission, error)
	DeleteAdminPermission(ctx context.Context, id int64) (bool, error)
	GetAdminRolePermKeys(ctx context.Context, roleID int64) ([]string, error)
	SetAdminRolePermKeys(ctx context.Context, roleID int64, permKeys []string) (bool, error)
	ListAdminApiRules(ctx context.Context) ([]*AdminApiRule, error)
	UpsertAdminApiRule(ctx context.Context, method, pathPattern, permKey, remark string, status int64) (*AdminApiRule, error)
	DeleteAdminApiRule(ctx context.Context, id int64) (bool, error)

	// Notifications (Redis fan-out; SSE served by service-hub HTTP)
	PublishPortalNotification(ctx context.Context, req *PublishPortalNotificationReq) (*PublishPortalNotificationResp, error)
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

func (d *defaultClient) UpsertDisplaySettings(ctx context.Context, country, currency, symbol string, fiatToUsdtRate float64, adminMfaEnabled int64, merchantNumericIDStart int64) error {
	_, err := d.cli.UpsertDisplaySettings(ctx, &servicehub.UpsertDisplaySettingsReq{
		CountryCode:            country,
		CurrencyCode:           currency,
		CurrencySymbol:         symbol,
		FiatToUsdtRate:         fiatToUsdtRate,
		AdminMfaEnabled:        adminMfaEnabled,
		MerchantNumericIdStart: merchantNumericIDStart,
	})
	return err
}

func (d *defaultClient) CreateAdminUser(ctx context.Context, username, passwordHash string, status int64) (*AdminUserPublic, error) {
	r, err := d.cli.CreateAdminUser(ctx, &servicehub.CreateAdminUserReq{
		Username: username, PasswordHash: passwordHash, Status: status,
	})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.User, nil
}

func (d *defaultClient) UpdateAdminUser(ctx context.Context, id int64, status int64, passwordHash, mfaSecret *string, mfaEnabled *int64) (*AdminUserPublic, error) {
	req := &servicehub.UpdateAdminUserReq{Id: id, Status: status, MfaEnabled: -1, MfaSecret: "__NO_CHANGE__"}
	if passwordHash != nil {
		req.PasswordHash = *passwordHash
	}
	if mfaSecret != nil {
		req.MfaSecret = *mfaSecret
	}
	if mfaEnabled != nil {
		req.MfaEnabled = *mfaEnabled
	}
	r, err := d.cli.UpdateAdminUser(ctx, req)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.User, nil
}

func (d *defaultClient) DeleteAdminUser(ctx context.Context, id int64) (bool, error) {
	r, err := d.cli.DeleteAdminUser(ctx, &servicehub.DeleteAdminUserReq{Id: id})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) GetAdminUserById(ctx context.Context, id int64) (*AdminUser, error) {
	r, err := d.cli.GetAdminUserById(ctx, &servicehub.GetAdminUserByIdReq{Id: id})
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

func (d *defaultClient) GetAdminRbacMyMenus(ctx context.Context, adminUserID int64) ([]*AdminMenu, error) {
	r, err := d.cli.GetAdminRbacMyMenus(ctx, &servicehub.GetAdminRbacMyMenusReq{AdminUserId: adminUserID})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Menus, nil
}

func (d *defaultClient) ListAdminRoles(ctx context.Context) ([]*AdminRole, error) {
	r, err := d.cli.ListAdminRoles(ctx, &servicehub.ListAdminRolesReq{})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Roles, nil
}

func (d *defaultClient) CreateAdminRole(ctx context.Context, code, name string, status int64) (*AdminRole, error) {
	r, err := d.cli.CreateAdminRole(ctx, &servicehub.CreateAdminRoleReq{Code: code, Name: name, Status: status})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Role, nil
}

func (d *defaultClient) UpdateAdminRole(ctx context.Context, id int64, name string, status int64) (*AdminRole, error) {
	r, err := d.cli.UpdateAdminRole(ctx, &servicehub.UpdateAdminRoleReq{Id: id, Name: name, Status: status})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Role, nil
}

func (d *defaultClient) DeleteAdminRole(ctx context.Context, id int64) (bool, error) {
	r, err := d.cli.DeleteAdminRole(ctx, &servicehub.DeleteAdminRoleReq{Id: id})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) ListAdminMenus(ctx context.Context) ([]*AdminMenu, error) {
	r, err := d.cli.ListAdminMenus(ctx, &servicehub.ListAdminMenusReq{})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Menus, nil
}

func (d *defaultClient) CreateAdminMenu(ctx context.Context, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error) {
	r, err := d.cli.CreateAdminMenu(ctx, &servicehub.CreateAdminMenuReq{
		ParentId: parentID, MenuKey: menuKey, Label: label, Icon: icon, Kind: kind, Path: path, SortOrder: sortOrder, Placement: placement,
	})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Menu, nil
}

func (d *defaultClient) UpdateAdminMenu(ctx context.Context, id int64, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error) {
	r, err := d.cli.UpdateAdminMenu(ctx, &servicehub.UpdateAdminMenuReq{
		Id: id, ParentId: parentID, MenuKey: menuKey, Label: label, Icon: icon, Kind: kind, Path: path, SortOrder: sortOrder, Placement: placement,
	})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Menu, nil
}

func (d *defaultClient) DeleteAdminMenu(ctx context.Context, id int64) (bool, error) {
	r, err := d.cli.DeleteAdminMenu(ctx, &servicehub.DeleteAdminMenuReq{Id: id})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) GetAdminRoleMenus(ctx context.Context, roleID int64) ([]int64, error) {
	r, err := d.cli.GetAdminRoleMenus(ctx, &servicehub.GetAdminRoleMenusReq{RoleId: roleID})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.MenuIds, nil
}

func (d *defaultClient) SetAdminRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) (bool, error) {
	r, err := d.cli.SetAdminRoleMenus(ctx, &servicehub.SetAdminRoleMenusReq{RoleId: roleID, MenuIds: menuIDs})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) GetAdminUserRoles(ctx context.Context, adminUserID int64) ([]int64, error) {
	r, err := d.cli.GetAdminUserRoles(ctx, &servicehub.GetAdminUserRolesReq{AdminUserId: adminUserID})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.RoleIds, nil
}

func (d *defaultClient) SetAdminUserRoles(ctx context.Context, adminUserID int64, roleIDs []int64) (bool, error) {
	r, err := d.cli.SetAdminUserRoles(ctx, &servicehub.SetAdminUserRolesReq{AdminUserId: adminUserID, RoleIds: roleIDs})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) GetAdminRbacMyPerms(ctx context.Context, adminUserID int64) (bool, []string, error) {
	r, err := d.cli.GetAdminRbacMyPerms(ctx, &servicehub.GetAdminRbacMyPermsReq{AdminUserId: adminUserID})
	if err != nil {
		return false, nil, err
	}
	if r == nil {
		return false, nil, nil
	}
	return r.IsSuperAdmin, r.PermKeys, nil
}

func (d *defaultClient) ListAdminPermissions(ctx context.Context) ([]*AdminPermission, error) {
	r, err := d.cli.ListAdminPermissions(ctx, &servicehub.ListAdminPermissionsReq{})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Permissions, nil
}

func (d *defaultClient) CreateAdminPermission(ctx context.Context, permKey, label, category, menuKey string, status int64) (*AdminPermission, error) {
	r, err := d.cli.CreateAdminPermission(ctx, &servicehub.CreateAdminPermissionReq{PermKey: permKey, Label: label, Category: category, MenuKey: menuKey, Status: status})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Permission, nil
}

func (d *defaultClient) UpdateAdminPermission(ctx context.Context, id int64, label, category, menuKey string, status int64) (*AdminPermission, error) {
	r, err := d.cli.UpdateAdminPermission(ctx, &servicehub.UpdateAdminPermissionReq{Id: id, Label: label, Category: category, MenuKey: menuKey, Status: status})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Permission, nil
}

func (d *defaultClient) DeleteAdminPermission(ctx context.Context, id int64) (bool, error) {
	r, err := d.cli.DeleteAdminPermission(ctx, &servicehub.DeleteAdminPermissionReq{Id: id})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) GetAdminRolePermKeys(ctx context.Context, roleID int64) ([]string, error) {
	r, err := d.cli.GetAdminRolePermKeys(ctx, &servicehub.GetAdminRolePermKeysReq{RoleId: roleID})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.PermKeys, nil
}

func (d *defaultClient) SetAdminRolePermKeys(ctx context.Context, roleID int64, permKeys []string) (bool, error) {
	r, err := d.cli.SetAdminRolePermKeys(ctx, &servicehub.SetAdminRolePermKeysReq{RoleId: roleID, PermKeys: permKeys})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) ListAdminApiRules(ctx context.Context) ([]*AdminApiRule, error) {
	r, err := d.cli.ListAdminApiRules(ctx, &servicehub.ListAdminApiRulesReq{})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Rules, nil
}

func (d *defaultClient) UpsertAdminApiRule(ctx context.Context, method, pathPattern, permKey, remark string, status int64) (*AdminApiRule, error) {
	r, err := d.cli.UpsertAdminApiRule(ctx, &servicehub.UpsertAdminApiRuleReq{Method: method, PathPattern: pathPattern, PermKey: permKey, Status: status, Remark: remark})
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r.Rule, nil
}

func (d *defaultClient) DeleteAdminApiRule(ctx context.Context, id int64) (bool, error) {
	r, err := d.cli.DeleteAdminApiRule(ctx, &servicehub.DeleteAdminApiRuleReq{Id: id})
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return r.Ok, nil
}

func (d *defaultClient) PublishPortalNotification(ctx context.Context, req *servicehub.PublishPortalNotificationReq) (*PublishPortalNotificationResp, error) {
	return d.cli.PublishPortalNotification(ctx, req)
}
