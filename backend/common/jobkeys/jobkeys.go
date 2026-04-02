// Package jobkeys 定义与 job-worker runner 中 registerJob 一致的 job_key，供管理台下拉与校验。
package jobkeys

import (
	"errors"
	"regexp"
)

// 与 runner.New 中 registerJob 的字符串保持一致（新增 handler 时在此增加常量并在 runner 注册）。
const (
	// TestLogHeartbeat 仅打时间日志，供联调/验证调度与多节点。
	TestLogHeartbeat = "test_log_heartbeat"
	// AdminOperationLogsCleanup 清理后台操作日志历史数据。
	AdminOperationLogsCleanup = "admin_operation_logs_cleanup"
)

var jobKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// RegisteredKeys 返回当前 job-worker 已实现的 handler key。
// 管理台 GET /v1/admin/jobs/keys 直接返回本列表（不读数据库）；须与 runner.New 里 registerJob 一一对应。
// 新增任务：在本包增加常量并加入本切片，同时在 runner 里 registerJob（job-worker 启动时会校验二者一致）。
func RegisteredKeys() []string {
	return []string{
		TestLogHeartbeat,
		AdminOperationLogsCleanup,
	}
}

// ValidateJobKey 校验 job_key 格式（与 DB uk_job_key、runner 路由一致）。
func ValidateJobKey(s string) error {
	if s == "" {
		return errors.New("job_key required")
	}
	if !jobKeyPattern.MatchString(s) {
		return errors.New("job_key must match ^[a-z][a-z0-9_]{0,62}$ (lowercase snake)")
	}
	return nil
}
