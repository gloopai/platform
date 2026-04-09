package logic

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gloopai/platform/service-hub/hubclient"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/gateway/internal/types"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
			ID:         r.GetId(),
			Username:   r.GetUsername(),
			Status:     r.GetStatus(),
			MfaEnabled: r.GetMfaEnabled(),
		})
	}
	return &types.AdminUsersResp{Users: out}, nil
}

func (a *AdminSystem) GetAdminMe(adminID int64) (*types.AdminMeResp, error) {
	if adminID <= 0 {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	u, err := a.svcCtx.ServiceHub.GetAdminUserById(a.ctx, adminID)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(u.GetUsername())
	email := ""
	if strings.Contains(username, "@") {
		email = username
	}
	display := email
	if display == "" {
		display = username
	}
	if display == "" {
		display = "管理员"
	}
	roleName := ""
	roleIDs, rerr := a.svcCtx.ServiceHub.GetAdminUserRoles(a.ctx, adminID)
	if rerr == nil && len(roleIDs) > 0 {
		allRoles, lerr := a.svcCtx.ServiceHub.ListAdminRoles(a.ctx)
		if lerr == nil {
			roleMap := make(map[int64]string, len(allRoles))
			for _, rr := range allRoles {
				if rr == nil {
					continue
				}
				roleMap[rr.GetId()] = strings.TrimSpace(rr.GetName())
			}
			names := make([]string, 0, len(roleIDs))
			for _, rid := range roleIDs {
				if n := strings.TrimSpace(roleMap[rid]); n != "" {
					names = append(names, n)
				}
			}
			if len(names) > 0 {
				roleName = strings.Join(names, "、")
			}
		}
	}
	if roleName == "" {
		roleName = "管理员"
	}
	mfaPending := int64(0)
	if u.GetMfaEnabled() != 1 && strings.TrimSpace(u.GetMfaSecret()) != "" {
		mfaPending = 1
	}
	return &types.AdminMeResp{
		ID:          u.GetId(),
		Username:    username,
		Email:       email,
		DisplayName: display,
		Role:        roleName,
		MfaEnabled:  u.GetMfaEnabled(),
		MfaPending:  mfaPending,
	}, nil
}

func (a *AdminSystem) GetDisplaySettings(req *types.AdminDisplaySettingsReq) (*types.AdminDisplaySettingsResp, error) {
	row, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
	if err != nil {
		return nil, err
	}
	start := row.GetMerchantNumericIdStart()
	if start < 1 {
		start = 5000000000
	}
	return &types.AdminDisplaySettingsResp{
		CountryCode:            row.GetCountryCode(),
		CurrencyCode:           row.GetCurrencyCode(),
		CurrencySymbol:         row.GetCurrencySymbol(),
		FiatToUsdtRate:         row.GetFiatToUsdtRate(),
		AdminMfaEnabled:        row.GetAdminMfaEnabled(),
		MerchantNumericIdStart: start,
		SystemName:             strings.TrimSpace(row.GetSystemName()),
	}, nil
}

func (a *AdminSystem) UpdateDisplaySettings(req *types.AdminDisplaySettingsUpdateReq) (*types.AdminDisplaySettingsResp, error) {
	country := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	currency := strings.ToUpper(strings.TrimSpace(req.CurrencyCode))
	symbol := strings.TrimSpace(req.CurrencySymbol)
	rate := req.FiatToUsdtRate
	if country == "" || currency == "" || symbol == "" || rate <= 0 {
		return nil, status.Error(codes.InvalidArgument, "country_code, currency_code, currency_symbol, fiat_to_usdt_rate required")
	}
	start := req.MerchantNumericIdStart
	if start == 0 {
		cur, gerr := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
		if gerr != nil {
			return nil, gerr
		}
		start = cur.GetMerchantNumericIdStart()
		if start < 1 {
			start = 5000000000
		}
	}
	if start < 1 || start > 9999999999 {
		return nil, status.Error(codes.InvalidArgument, "merchant_numeric_id_start must be 1..9999999999")
	}
	if err := a.svcCtx.ServiceHub.UpsertDisplaySettings(a.ctx, country, currency, symbol, rate, req.AdminMfaEnabled, start, req.SystemName); err != nil {
		return nil, err
	}
	return a.GetDisplaySettings(&types.AdminDisplaySettingsReq{})
}

func (a *AdminSystem) ListAdminOperationLogs(req *types.AdminOperationLogsReq) (*types.AdminOperationLogsResp, error) {
	now := time.Now()
	startSec := req.StartSec
	endSec := req.EndSec
	if startSec == 0 && endSec == 0 {
		endSec = now.Unix()
		startSec = now.Add(-24 * time.Hour).Unix()
	} else if endSec == 0 {
		endSec = now.Unix()
	} else if startSec == 0 {
		startSec = time.Unix(endSec, 0).Add(-24 * time.Hour).Unix()
	}
	grpcReq := &hubclient.ListAdminOperationLogsReq{
		StartSec:    startSec,
		EndSec:      endSec,
		AdminUserId: req.AdminUserID,
		Method:      strings.ToUpper(strings.TrimSpace(req.Method)),
		PathKeyword: strings.TrimSpace(req.PathKeyword),
		PermKey:     strings.TrimSpace(req.PermKey),
		Limit:       req.Limit,
		Offset:      req.Offset,
	}
	switch strings.TrimSpace(req.Success) {
	case "1":
		v := true
		grpcReq.Success = &v
	case "0":
		v := false
		grpcReq.Success = &v
	}
	rows, total, err := a.svcCtx.ServiceHub.ListAdminOperationLogs(a.ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminOperationLogRow, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		out = append(out, types.AdminOperationLogRow{
			ID:            r.GetId(),
			CreatedAt:     r.GetCreatedAt(),
			RequestID:     r.GetRequestId(),
			AdminUserID:   r.GetAdminUserId(),
			AdminUsername: r.GetAdminUsername(),
			OperatorIP:    r.GetOperatorIp(),
			UserAgent:     r.GetUserAgent(),
			Method:        r.GetMethod(),
			Path:          r.GetPath(),
			QueryString:   r.GetQueryString(),
			PermKey:       r.GetPermKey(),
			HTTPStatus:    r.GetHttpStatus(),
			Success:       r.GetSuccess(),
			DurationMs:    r.GetDurationMs(),
			ErrorMessage:  r.GetErrorMessage(),
		})
	}
	return &types.AdminOperationLogsResp{Rows: out, Total: total}, nil
}

func (a *AdminSystem) CreateAdminUser(req *types.AdminCreateUserReq) (*types.AdminUsersResp, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	statusV := req.Status
	if statusV == 0 {
		statusV = 1
	}
	u, err := a.svcCtx.ServiceHub.CreateAdminUser(a.ctx, username, string(hash), statusV)
	if err != nil {
		return nil, err
	}
	if len(req.RoleIds) > 0 && u != nil {
		_, _ = a.svcCtx.ServiceHub.SetAdminUserRoles(a.ctx, u.GetId(), req.RoleIds)
	}
	return a.ListAdminUsers()
}

func (a *AdminSystem) UpdateAdminUser(req *types.AdminUpdateUserReq) (*types.AdminUsersResp, error) {
	_, err := a.svcCtx.ServiceHub.UpdateAdminUser(a.ctx, req.Id, req.Status, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if req.RoleIds != nil {
		if _, err := a.svcCtx.ServiceHub.SetAdminUserRoles(a.ctx, req.Id, req.RoleIds); err != nil {
			return nil, err
		}
	}
	return a.ListAdminUsers()
}

func (a *AdminSystem) DeleteAdminUser(req *types.AdminDeleteUserReq) (map[string]any, error) {
	ok, err := a.svcCtx.ServiceHub.DeleteAdminUser(a.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": ok}, nil
}

func (a *AdminSystem) ResetAdminUserPassword(req *types.AdminResetUserPasswordReq) (map[string]any, error) {
	password := strings.TrimSpace(req.Password)
	if len(password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "password too short")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user, err := a.svcCtx.ServiceHub.GetAdminUserById(a.ctx, req.Id)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if _, err := a.svcCtx.ServiceHub.UpdateAdminUser(a.ctx, req.Id, user.GetStatus(), ptrStr(string(hash)), nil, nil); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (a *AdminSystem) SetupAdminUserMfa(req *types.AdminMfaSetupReq) (*types.AdminMfaSetupResp, error) {
	user, err := a.svcCtx.ServiceHub.GetAdminUserById(a.ctx, req.Id)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	issuer := adminMfaIssuer(a)
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.GetUsername(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	secret := key.Secret()
	if _, err := a.svcCtx.ServiceHub.UpdateAdminUser(a.ctx, req.Id, user.GetStatus(), nil, ptrStr(secret), ptrI64(0)); err != nil {
		return nil, err
	}
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.AdminMfaSetupResp{
		Secret:     secret,
		OtpAuthUrl: key.URL(),
		QrDataUrl:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	}, nil
}

func (a *AdminSystem) ConfirmAdminUserMfa(req *types.AdminMfaConfirmReq) (map[string]any, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	user, err := a.svcCtx.ServiceHub.GetAdminUserById(a.ctx, req.Id)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	secret := strings.TrimSpace(user.GetMfaSecret())
	if secret == "" {
		return nil, status.Error(codes.FailedPrecondition, "mfa secret not initialized")
	}
	if !totp.Validate(code, secret) {
		return nil, status.Error(codes.InvalidArgument, "invalid mfa code")
	}
	if _, err := a.svcCtx.ServiceHub.UpdateAdminUser(a.ctx, req.Id, user.GetStatus(), nil, ptrStr(secret), ptrI64(1)); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// SetupAdminSelfMfa 当前登录用户发起 MFA 绑定（生成密钥与二维码）。
func (a *AdminSystem) SetupAdminSelfMfa(adminID int64) (*types.AdminMfaSetupResp, error) {
	return a.SetupAdminUserMfa(&types.AdminMfaSetupReq{Id: adminID})
}

// ConfirmAdminSelfMfa 当前登录用户确认 MFA 绑定。
func (a *AdminSystem) ConfirmAdminSelfMfa(adminID int64, req *types.AdminMfaConfirmSelfReq) (map[string]any, error) {
	code := strings.TrimSpace(req.Code)
	return a.ConfirmAdminUserMfa(&types.AdminMfaConfirmReq{Id: adminID, Code: code})
}

func (a *AdminSystem) DisableAdminUserMfa(req *types.AdminMfaDisableReq) (map[string]any, error) {
	user, err := a.svcCtx.ServiceHub.GetAdminUserById(a.ctx, req.Id)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if _, err := a.svcCtx.ServiceHub.UpdateAdminUser(a.ctx, req.Id, user.GetStatus(), nil, ptrStr(""), ptrI64(0)); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func ptrStr(v string) *string { return &v }
func ptrI64(v int64) *int64   { return &v }

// 与未配置「系统名称」时的历史 MFA issuer 一致
const defaultAdminSystemName = "Pay Platform Admin"

func adminMfaIssuer(a *AdminSystem) string {
	ds, err := a.svcCtx.ServiceHub.GetDisplaySettings(a.ctx)
	if err != nil || ds == nil {
		return defaultAdminSystemName
	}
	s := strings.TrimSpace(ds.GetSystemName())
	if s == "" {
		return defaultAdminSystemName
	}
	return s
}
