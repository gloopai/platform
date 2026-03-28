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
			MfaSecret:    u.MfaSecret,
			MfaEnabled:   u.MfaEnabled,
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
			Id:         r.ID,
			Username:   r.Username,
			Status:     r.Status,
			MfaEnabled: r.MfaEnabled,
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
		CountryCode:            row.CountryCode,
		CurrencyCode:           row.CurrencyCode,
		CurrencySymbol:         row.CurrencySymbol,
		FiatToUsdtRate:         row.FiatToUsdtRate,
		AdminMfaEnabled:        row.AdminMfaEnabled,
		MerchantNumericIdStart: row.MerchantNumericIDStart,
	}, nil
}

func (s *ServiceHubServer) UpsertDisplaySettings(ctx context.Context, req *servicehub.UpsertDisplaySettingsReq) (*servicehub.GetDisplaySettingsResp, error) {
	country := strings.ToUpper(strings.TrimSpace(req.GetCountryCode()))
	currency := strings.ToUpper(strings.TrimSpace(req.GetCurrencyCode()))
	symbol := strings.TrimSpace(req.GetCurrencySymbol())
	rate := req.GetFiatToUsdtRate()
	if country == "" || currency == "" || symbol == "" || rate <= 0 {
		return nil, status.Error(codes.InvalidArgument, "country_code, currency_code, currency_symbol, fiat_to_usdt_rate required")
	}
	start := req.GetMerchantNumericIdStart()
	if start == 0 {
		cur, err := s.svcCtx.GlobalSettings.GetDisplaySettings(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		start = cur.MerchantNumericIDStart
	}
	if start < 1 || start > 9999999999 {
		return nil, status.Error(codes.InvalidArgument, "merchant_numeric_id_start must be 1..9999999999")
	}
	if err := s.svcCtx.GlobalSettings.UpsertDisplaySettings(ctx, &store.GlobalDisplaySettings{
		CountryCode:            country,
		CurrencyCode:           currency,
		CurrencySymbol:         symbol,
		FiatToUsdtRate:         rate,
		AdminMfaEnabled:        req.GetAdminMfaEnabled(),
		MerchantNumericIDStart: start,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.GetDisplaySettingsResp{
		CountryCode:            country,
		CurrencyCode:           currency,
		CurrencySymbol:         symbol,
		FiatToUsdtRate:         rate,
		AdminMfaEnabled:        req.GetAdminMfaEnabled(),
		MerchantNumericIdStart: start,
	}, nil
}

func (s *ServiceHubServer) CreateAdminUser(ctx context.Context, req *servicehub.CreateAdminUserReq) (*servicehub.CreateAdminUserResp, error) {
	username := strings.TrimSpace(req.GetUsername())
	passwordHash := strings.TrimSpace(req.GetPasswordHash())
	if username == "" || passwordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password_hash required")
	}
	row, err := s.svcCtx.AdminUsers.Create(ctx, username, passwordHash, req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.CreateAdminUserResp{
		User: &servicehub.AdminUserPublic{
			Id:         row.ID,
			Username:   row.Username,
			Status:     row.Status,
			MfaEnabled: row.MfaEnabled,
		},
	}, nil
}

func (s *ServiceHubServer) UpdateAdminUser(ctx context.Context, req *servicehub.UpdateAdminUserReq) (*servicehub.UpdateAdminUserResp, error) {
	var passwordHash *string
	if strings.TrimSpace(req.GetPasswordHash()) != "" {
		v := strings.TrimSpace(req.GetPasswordHash())
		passwordHash = &v
	}
	var mfaSecret *string
	if req.GetMfaSecret() != "__NO_CHANGE__" {
		v := req.GetMfaSecret()
		mfaSecret = &v
	}
	var mfaEnabled *int64
	vEnabled := req.GetMfaEnabled()
	if vEnabled >= 0 {
		mfaEnabled = &vEnabled
	}

	row, err := s.svcCtx.AdminUsers.Update(ctx, req.GetId(), req.GetStatus(), passwordHash, mfaSecret, mfaEnabled)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.UpdateAdminUserResp{
		User: &servicehub.AdminUserPublic{
			Id:         row.ID,
			Username:   row.Username,
			Status:     row.Status,
			MfaEnabled: row.MfaEnabled,
		},
	}, nil
}

func (s *ServiceHubServer) DeleteAdminUser(ctx context.Context, req *servicehub.DeleteAdminUserReq) (*servicehub.DeleteAdminUserResp, error) {
	if err := s.svcCtx.AdminUsers.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.DeleteAdminUserResp{Ok: true}, nil
}

func (s *ServiceHubServer) GetAdminUserById(ctx context.Context, req *servicehub.GetAdminUserByIdReq) (*servicehub.GetAdminUserByIdResp, error) {
	row, err := s.svcCtx.AdminUsers.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.GetAdminUserByIdResp{
		User: &servicehub.AdminUser{
			Id:           row.ID,
			Username:     row.Username,
			PasswordHash: row.PasswordHash,
			Status:       row.Status,
			MfaSecret:    row.MfaSecret,
			MfaEnabled:   row.MfaEnabled,
		},
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

func (s *ServiceHubServer) GetAdminRbacMyMenus(ctx context.Context, req *servicehub.GetAdminRbacMyMenusReq) (*servicehub.GetAdminRbacMyMenusResp, error) {
	uid := req.GetAdminUserId()
	if uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "admin_user_id required")
	}
	rows, err := s.svcCtx.AdminRbac.ListMenusByUser(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminMenu, 0, len(rows))
	for _, m := range rows {
		out = append(out, &servicehub.AdminMenu{
			Id:        m.ID,
			ParentId:  m.ParentID,
			MenuKey:   m.MenuKey,
			Label:     m.Label,
			Icon:      m.Icon,
			Kind:      m.Kind,
			Path:      m.Path,
			SortOrder: m.SortOrder,
			Placement: m.Placement,
		})
	}
	return &servicehub.GetAdminRbacMyMenusResp{Menus: out}, nil
}

func (s *ServiceHubServer) ListAdminRoles(ctx context.Context, _ *servicehub.ListAdminRolesReq) (*servicehub.ListAdminRolesResp, error) {
	rows, err := s.svcCtx.AdminRbac.ListRoles(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminRole, 0, len(rows))
	for _, r := range rows {
		out = append(out, &servicehub.AdminRole{
			Id:     r.ID,
			Code:   r.Code,
			Name:   r.Name,
			Status: r.Status,
		})
	}
	return &servicehub.ListAdminRolesResp{Roles: out}, nil
}

func (s *ServiceHubServer) CreateAdminRole(ctx context.Context, req *servicehub.CreateAdminRoleReq) (*servicehub.CreateAdminRoleResp, error) {
	code := strings.TrimSpace(req.GetCode())
	name := strings.TrimSpace(req.GetName())
	statusV := req.GetStatus()
	if statusV == 0 {
		statusV = 1
	}
	r, err := s.svcCtx.AdminRbac.CreateRole(ctx, code, name, statusV)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.CreateAdminRoleResp{
		Role: &servicehub.AdminRole{Id: r.ID, Code: r.Code, Name: r.Name, Status: r.Status},
	}, nil
}

func (s *ServiceHubServer) UpdateAdminRole(ctx context.Context, req *servicehub.UpdateAdminRoleReq) (*servicehub.UpdateAdminRoleResp, error) {
	r, err := s.svcCtx.AdminRbac.UpdateRole(ctx, req.GetId(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.UpdateAdminRoleResp{
		Role: &servicehub.AdminRole{Id: r.ID, Code: r.Code, Name: r.Name, Status: r.Status},
	}, nil
}

func (s *ServiceHubServer) DeleteAdminRole(ctx context.Context, req *servicehub.DeleteAdminRoleReq) (*servicehub.DeleteAdminRoleResp, error) {
	if err := s.svcCtx.AdminRbac.DeleteRole(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.DeleteAdminRoleResp{Ok: true}, nil
}

func (s *ServiceHubServer) ListAdminMenus(ctx context.Context, _ *servicehub.ListAdminMenusReq) (*servicehub.ListAdminMenusResp, error) {
	rows, err := s.svcCtx.AdminRbac.ListMenus(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminMenu, 0, len(rows))
	for _, m := range rows {
		out = append(out, &servicehub.AdminMenu{
			Id:        m.ID,
			ParentId:  m.ParentID,
			MenuKey:   m.MenuKey,
			Label:     m.Label,
			Icon:      m.Icon,
			Kind:      m.Kind,
			Path:      m.Path,
			SortOrder: m.SortOrder,
			Placement: m.Placement,
		})
	}
	return &servicehub.ListAdminMenusResp{Menus: out}, nil
}

func (s *ServiceHubServer) CreateAdminMenu(ctx context.Context, req *servicehub.CreateAdminMenuReq) (*servicehub.CreateAdminMenuResp, error) {
	p := req.GetParentId()
	if p < 0 {
		p = 0
	}
	m, err := s.svcCtx.AdminRbac.CreateMenu(ctx, p, req.GetMenuKey(), req.GetLabel(), req.GetIcon(), req.GetKind(), req.GetPath(), req.GetSortOrder(), req.GetPlacement())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.CreateAdminMenuResp{
		Menu: &servicehub.AdminMenu{
			Id:        m.ID,
			ParentId:  m.ParentID,
			MenuKey:   m.MenuKey,
			Label:     m.Label,
			Icon:      m.Icon,
			Kind:      m.Kind,
			Path:      m.Path,
			SortOrder: m.SortOrder,
			Placement: m.Placement,
		},
	}, nil
}

func (s *ServiceHubServer) UpdateAdminMenu(ctx context.Context, req *servicehub.UpdateAdminMenuReq) (*servicehub.UpdateAdminMenuResp, error) {
	p := req.GetParentId()
	if p < 0 {
		p = 0
	}
	m, err := s.svcCtx.AdminRbac.UpdateMenu(ctx, req.GetId(), p, req.GetMenuKey(), req.GetLabel(), req.GetIcon(), req.GetKind(), req.GetPath(), req.GetSortOrder(), req.GetPlacement())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.UpdateAdminMenuResp{
		Menu: &servicehub.AdminMenu{
			Id:        m.ID,
			ParentId:  m.ParentID,
			MenuKey:   m.MenuKey,
			Label:     m.Label,
			Icon:      m.Icon,
			Kind:      m.Kind,
			Path:      m.Path,
			SortOrder: m.SortOrder,
			Placement: m.Placement,
		},
	}, nil
}

func (s *ServiceHubServer) DeleteAdminMenu(ctx context.Context, req *servicehub.DeleteAdminMenuReq) (*servicehub.DeleteAdminMenuResp, error) {
	if err := s.svcCtx.AdminRbac.DeleteMenu(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.DeleteAdminMenuResp{Ok: true}, nil
}

func (s *ServiceHubServer) GetAdminRoleMenus(ctx context.Context, req *servicehub.GetAdminRoleMenusReq) (*servicehub.GetAdminRoleMenusResp, error) {
	ids, err := s.svcCtx.AdminRbac.GetRoleMenuIDs(ctx, req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.GetAdminRoleMenusResp{MenuIds: ids}, nil
}

func (s *ServiceHubServer) SetAdminRoleMenus(ctx context.Context, req *servicehub.SetAdminRoleMenusReq) (*servicehub.SetAdminRoleMenusResp, error) {
	if err := s.svcCtx.AdminRbac.SetRoleMenuIDs(ctx, req.GetRoleId(), req.GetMenuIds()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.SetAdminRoleMenusResp{Ok: true}, nil
}

func (s *ServiceHubServer) GetAdminUserRoles(ctx context.Context, req *servicehub.GetAdminUserRolesReq) (*servicehub.GetAdminUserRolesResp, error) {
	ids, err := s.svcCtx.AdminRbac.GetUserRoleIDs(ctx, req.GetAdminUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.GetAdminUserRolesResp{RoleIds: ids}, nil
}

func (s *ServiceHubServer) SetAdminUserRoles(ctx context.Context, req *servicehub.SetAdminUserRolesReq) (*servicehub.SetAdminUserRolesResp, error) {
	if err := s.svcCtx.AdminRbac.SetUserRoleIDs(ctx, req.GetAdminUserId(), req.GetRoleIds()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.SetAdminUserRolesResp{Ok: true}, nil
}

func (s *ServiceHubServer) GetAdminRbacMyPerms(ctx context.Context, req *servicehub.GetAdminRbacMyPermsReq) (*servicehub.GetAdminRbacMyPermsResp, error) {
	uid := req.GetAdminUserId()
	if uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "admin_user_id required")
	}
	isSuper, keys, err := s.svcCtx.AdminRbac.ListPermKeysByUser(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.GetAdminRbacMyPermsResp{IsSuperAdmin: isSuper, PermKeys: keys}, nil
}

func (s *ServiceHubServer) ListAdminPermissions(ctx context.Context, _ *servicehub.ListAdminPermissionsReq) (*servicehub.ListAdminPermissionsResp, error) {
	rows, err := s.svcCtx.AdminRbacCfg.ListPermissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminPermission, 0, len(rows))
	for _, p := range rows {
		out = append(out, &servicehub.AdminPermission{
			Id:       p.ID,
			PermKey:  p.PermKey,
			Label:    p.Label,
			Category: p.Category,
			MenuKey:  p.MenuKey,
			Status:   p.Status,
		})
	}
	return &servicehub.ListAdminPermissionsResp{Permissions: out}, nil
}

func (s *ServiceHubServer) CreateAdminPermission(ctx context.Context, req *servicehub.CreateAdminPermissionReq) (*servicehub.CreateAdminPermissionResp, error) {
	p, err := s.svcCtx.AdminRbacCfg.CreatePermission(ctx, req.GetPermKey(), req.GetLabel(), req.GetCategory(), req.GetMenuKey(), req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.CreateAdminPermissionResp{
		Permission: &servicehub.AdminPermission{Id: p.ID, PermKey: p.PermKey, Label: p.Label, Category: p.Category, MenuKey: p.MenuKey, Status: p.Status},
	}, nil
}

func (s *ServiceHubServer) UpdateAdminPermission(ctx context.Context, req *servicehub.UpdateAdminPermissionReq) (*servicehub.UpdateAdminPermissionResp, error) {
	p, err := s.svcCtx.AdminRbacCfg.UpdatePermission(ctx, req.GetId(), req.GetLabel(), req.GetCategory(), req.GetMenuKey(), req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.UpdateAdminPermissionResp{
		Permission: &servicehub.AdminPermission{Id: p.ID, PermKey: p.PermKey, Label: p.Label, Category: p.Category, MenuKey: p.MenuKey, Status: p.Status},
	}, nil
}

func (s *ServiceHubServer) DeleteAdminPermission(ctx context.Context, req *servicehub.DeleteAdminPermissionReq) (*servicehub.DeleteAdminPermissionResp, error) {
	if err := s.svcCtx.AdminRbacCfg.DeletePermission(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.DeleteAdminPermissionResp{Ok: true}, nil
}

func (s *ServiceHubServer) GetAdminRolePermKeys(ctx context.Context, req *servicehub.GetAdminRolePermKeysReq) (*servicehub.GetAdminRolePermKeysResp, error) {
	keys, err := s.svcCtx.AdminRbacCfg.GetRolePermKeys(ctx, req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.GetAdminRolePermKeysResp{PermKeys: keys}, nil
}

func (s *ServiceHubServer) SetAdminRolePermKeys(ctx context.Context, req *servicehub.SetAdminRolePermKeysReq) (*servicehub.SetAdminRolePermKeysResp, error) {
	if err := s.svcCtx.AdminRbacCfg.SetRolePermKeys(ctx, req.GetRoleId(), req.GetPermKeys()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.SetAdminRolePermKeysResp{Ok: true}, nil
}

func (s *ServiceHubServer) ListAdminApiRules(ctx context.Context, _ *servicehub.ListAdminApiRulesReq) (*servicehub.ListAdminApiRulesResp, error) {
	rows, err := s.svcCtx.AdminRbacCfg.ListApiRules(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*servicehub.AdminApiRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, &servicehub.AdminApiRule{
			Id:          r.ID,
			Method:      r.Method,
			PathPattern: r.PathPattern,
			PermKey:     r.PermKey,
			Status:      r.Status,
			Remark:      r.Remark,
		})
	}
	return &servicehub.ListAdminApiRulesResp{Rules: out}, nil
}

func (s *ServiceHubServer) UpsertAdminApiRule(ctx context.Context, req *servicehub.UpsertAdminApiRuleReq) (*servicehub.UpsertAdminApiRuleResp, error) {
	r, err := s.svcCtx.AdminRbacCfg.UpsertApiRule(ctx, req.GetMethod(), req.GetPathPattern(), req.GetPermKey(), req.GetStatus(), req.GetRemark())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &servicehub.UpsertAdminApiRuleResp{
		Rule: &servicehub.AdminApiRule{Id: r.ID, Method: r.Method, PathPattern: r.PathPattern, PermKey: r.PermKey, Status: r.Status, Remark: r.Remark},
	}, nil
}

func (s *ServiceHubServer) DeleteAdminApiRule(ctx context.Context, req *servicehub.DeleteAdminApiRuleReq) (*servicehub.DeleteAdminApiRuleResp, error) {
	if err := s.svcCtx.AdminRbacCfg.DeleteApiRule(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &servicehub.DeleteAdminApiRuleResp{Ok: true}, nil
}

func (s *ServiceHubServer) PublishPortalNotification(ctx context.Context, req *servicehub.PublishPortalNotificationReq) (*servicehub.PublishPortalNotificationResp, error) {
	if s.svcCtx.NotifyPublisher == nil {
		return nil, status.Error(codes.FailedPrecondition, "notify nsq not configured")
	}
	id, err := s.svcCtx.NotifyPublisher.Publish(ctx, req)
	if err != nil {
		return nil, err
	}
	return &servicehub.PublishPortalNotificationResp{NotificationId: id}, nil
}
