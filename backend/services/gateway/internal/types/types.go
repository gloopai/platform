// platform-admin：管理端请求/响应类型（与 gateway.api 解耦，手写维护）。

package types

type AdminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	MfaCode  string `json:"mfa_code,optional"`
}

type AdminLoginResp struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type AdminLogoutResp struct {
	Ok bool `json:"ok"`
}

type AdminUserRow struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Status     int64  `json:"status"`
	MfaEnabled int64  `json:"mfa_enabled"`
}

type AdminUsersResp struct {
	Users []AdminUserRow `json:"users"`
}

type AdminMeResp struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type AdminDisplaySettingsReq struct{}

type AdminDisplaySettingsUpdateReq struct {
	CountryCode            string  `json:"country_code"`
	CurrencyCode           string  `json:"currency_code"`
	CurrencySymbol         string  `json:"currency_symbol"`
	FiatToUsdtRate         float64 `json:"fiat_to_usdt_rate"`
	AdminMfaEnabled        int64   `json:"admin_mfa_enabled,optional"`
	MerchantNumericIdStart int64   `json:"merchant_numeric_id_start,optional"`
}

type AdminDisplaySettingsResp struct {
	CountryCode            string  `json:"country_code"`
	CurrencyCode           string  `json:"currency_code"`
	CurrencySymbol         string  `json:"currency_symbol"`
	FiatToUsdtRate         float64 `json:"fiat_to_usdt_rate"`
	AdminMfaEnabled        int64   `json:"admin_mfa_enabled"`
	MerchantNumericIdStart int64   `json:"merchant_numeric_id_start"`
}

type AdminCreateUserReq struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Status   int64   `json:"status,optional"`
	RoleIds  []int64 `json:"role_ids,optional"`
}

type AdminUpdateUserReq struct {
	Id      int64   `path:"id"`
	Status  int64   `json:"status"`
	RoleIds []int64 `json:"role_ids,optional"`
}

type AdminResetUserPasswordReq struct {
	Id       int64  `path:"id"`
	Password string `json:"password"`
}

type AdminDeleteUserReq struct {
	Id int64 `path:"id"`
}

type AdminMfaSetupReq struct {
	Id int64 `path:"id"`
}

type AdminMfaSetupResp struct {
	Secret     string `json:"secret"`
	OtpAuthUrl string `json:"otpauth_url"`
	QrDataUrl  string `json:"qr_data_url"`
}

type AdminMfaConfirmReq struct {
	Id   int64  `path:"id"`
	Code string `json:"code"`
}

type AdminMfaDisableReq struct {
	Id int64 `path:"id"`
}
