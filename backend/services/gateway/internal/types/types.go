// platform-admin：管理端请求/响应类型（与 gateway.api 解耦，手写维护）。

package types

type AdminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	MfaCode  string `json:"mfa_code,optional"`
}

type AdminLoginResp struct {
	Token            string `json:"token"`
	ExpiresAt        int64  `json:"expires_at"`
	MfaSetupRequired bool   `json:"mfa_setup_required"`
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
	MfaEnabled  int64  `json:"mfa_enabled"`
	MfaPending  int64  `json:"mfa_pending"`
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

// AdminMfaConfirmSelfReq 当前登录用户确认绑定 MFA（body）。
type AdminMfaConfirmSelfReq struct {
	Code string `json:"code"`
}

// --- 定时任务（scheduled_jobs / job-worker） ---

type AdminScheduledJobsReq struct {
	Limit  int64 `form:"limit,optional"`
	Offset int64 `form:"offset,optional"`
}

// AdminScheduledJobKeysResp 当前 job-worker 已实现的 job_key 列表 + 合法格式说明（自定义 key 须先在 runner 注册）。
type AdminScheduledJobKeysResp struct {
	Keys    []string `json:"keys"`
	Pattern string   `json:"pattern"`
}

type AdminScheduledJobsResp struct {
	Jobs  []AdminScheduledJob `json:"jobs"`
	Total int64               `json:"total"`
}

type AdminScheduledJob struct {
	Id                  int64  `json:"id"`
	JobKey              string `json:"job_key"`
	Name                string `json:"name"`
	Category            string `json:"category"`
	Enabled             bool   `json:"enabled"`
	Builtin             bool   `json:"builtin"`
	ScheduleType        string `json:"schedule_type"`
	CronExpr            string `json:"cron_expr"`
	IntervalSeconds     int64  `json:"interval_seconds"`
	Timezone            string `json:"timezone"`
	PayloadJson         string `json:"payload_json"`
	ConcurrencyPolicy   string `json:"concurrency_policy"`
	MisfirePolicy       string `json:"misfire_policy"`
	MaxRetry            int64  `json:"max_retry"`
	RetryBackoffSeconds int64  `json:"retry_backoff_seconds"`
	NextRunAt           int64  `json:"next_run_at"`
	LastRunAt           int64  `json:"last_run_at"`
	LastStatus          string `json:"last_status"`
	LastError           string `json:"last_error"`
	UpdatedBy           string `json:"updated_by"`
}

type AdminCreateScheduledJobReq struct {
	JobKey              string `json:"job_key"`
	Name                string `json:"name"`
	Category            string `json:"category,optional"`
	Enabled             bool   `json:"enabled,optional"`
	ScheduleType        string `json:"schedule_type,optional"`
	CronExpr            string `json:"cron_expr,optional"`
	IntervalSeconds     int64  `json:"interval_seconds,optional"`
	Timezone            string `json:"timezone,optional"`
	PayloadJson         string `json:"payload_json,optional"`
	ConcurrencyPolicy   string `json:"concurrency_policy,optional"`
	MisfirePolicy       string `json:"misfire_policy,optional"`
	MaxRetry            int64  `json:"max_retry,optional"`
	RetryBackoffSeconds int64  `json:"retry_backoff_seconds,optional"`
}

type AdminUpdateScheduledJobReq struct {
	Id                  int64  `path:"id"`
	Name                string `json:"name"`
	Category            string `json:"category,optional"`
	ScheduleType        string `json:"schedule_type,optional"`
	CronExpr            string `json:"cron_expr,optional"`
	IntervalSeconds     int64  `json:"interval_seconds,optional"`
	Timezone            string `json:"timezone,optional"`
	PayloadJson         string `json:"payload_json,optional"`
	ConcurrencyPolicy   string `json:"concurrency_policy,optional"`
	MisfirePolicy       string `json:"misfire_policy,optional"`
	MaxRetry            int64  `json:"max_retry,optional"`
	RetryBackoffSeconds int64  `json:"retry_backoff_seconds,optional"`
	NextRunAt           int64  `json:"next_run_at,optional"`
}

type AdminToggleScheduledJobReq struct {
	Id      int64 `path:"id"`
	Enabled bool  `json:"enabled"`
}

type AdminRunScheduledJobReq struct {
	Id            int64  `path:"id"`
	CorrelationId string `json:"correlation_id,optional"`
}

type AdminSimpleOkResp struct {
	Ok bool `json:"ok"`
}

type AdminScheduledJobRunsReq struct {
	JobId       int64  `form:"job_id,optional"`
	Status      string `form:"status,optional"`
	TriggerType string `form:"trigger_type,optional"`
	Limit       int64  `form:"limit,optional"`
	Offset      int64  `form:"offset,optional"`
}

type AdminScheduledJobRunsResp struct {
	Runs  []AdminScheduledJobRun `json:"runs"`
	Total int64                  `json:"total"`
}

type AdminJobWorkerNodesResp struct {
	Nodes       []AdminJobWorkerNode `json:"nodes"`
	QueuedTotal int64                `json:"queued_total"`
}

type AdminJobWorkerNode struct {
	WorkerId        string `json:"worker_id"`
	Hostname        string `json:"hostname"`
	LastHeartbeatAt int64  `json:"last_heartbeat_at"`
	RunningTasks    int64  `json:"running_tasks"`
	SuccessLastHour int64  `json:"success_last_hour"`
}

type AdminScheduledJobRun struct {
	Id            int64  `json:"id"`
	JobId         int64  `json:"job_id"`
	JobKey        string `json:"job_key"`
	JobName       string `json:"job_name"`
	TriggerType   string `json:"trigger_type"`
	ScheduledAt   int64  `json:"scheduled_at"`
	StartedAt     int64  `json:"started_at"`
	FinishedAt    int64  `json:"finished_at"`
	DurationMs    int64  `json:"duration_ms"`
	Status        string `json:"status"`
	Attempt       int64  `json:"attempt"`
	WorkerId      string `json:"worker_id"`
	Summary       string `json:"summary"`
	ErrorCode     string `json:"error_code"`
	ErrorMessage  string `json:"error_message"`
	OutputJson    string `json:"output_json"`
	CorrelationId string `json:"correlation_id"`
}

type AdminScheduledJobRunIdReq struct {
	Id int64 `path:"id"`
}

type AdminScheduledJobRunResp struct {
	Run AdminScheduledJobRun `json:"run"`
}

// --- 操作审计（admin_operation_logs） ---

type AdminOperationLogsReq struct {
	StartSec    int64  `form:"start_sec,optional"`
	EndSec      int64  `form:"end_sec,optional"`
	AdminUserID int64  `form:"admin_user_id,optional"`
	Method      string `form:"method,optional"`
	PathKeyword string `form:"path_keyword,optional"`
	PermKey     string `form:"perm_key,optional"`
	Success     string `form:"success,optional"` // 空=不限，1=成功，0=失败
	Limit       int64  `form:"limit,optional"`
	Offset      int64  `form:"offset,optional"`
}

type AdminOperationLogRow struct {
	ID            int64  `json:"id"`
	CreatedAt     int64  `json:"created_at"`
	RequestID     string `json:"request_id"`
	AdminUserID   int64  `json:"admin_user_id"`
	AdminUsername string `json:"admin_username"`
	OperatorIP    string `json:"operator_ip"`
	UserAgent     string `json:"user_agent"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	QueryString   string `json:"query_string"`
	PermKey       string `json:"perm_key"`
	HTTPStatus    int32  `json:"http_status"`
	Success       bool   `json:"success"`
	DurationMs    int64  `json:"duration_ms"`
	ErrorMessage  string `json:"error_message"`
}

type AdminOperationLogsResp struct {
	Rows  []AdminOperationLogRow `json:"rows"`
	Total int64                  `json:"total"`
}
