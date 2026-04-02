package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gloopai/pay/job-worker/internal/store"
)

// ------------------------------
// 新任务开发模板（复制本文件中的示例）
// ------------------------------
//
// 1) 在 common/jobkeys/jobkeys.go 增加常量并加入 RegisteredKeys()，在 New(...) 里注册（runner.go）：
//    r.registerJob(jobkeys.YourJobKey, r.execYourJob)
//    管理台新建任务时从下拉选择同一字符串（或「自定义」填写与常量一致的值）。
//    规范：所有定时任务都必须接入 scheduled_jobs / job-worker 体系，不要在业务服务内单独起 ticker/cron。
//
// 2) 在此文件新增 payload + handler：
//    - payload 用于解析 scheduled_jobs.payload_json
//    - handler 内写核心逻辑，最后调用 FinishRun 写日志
//
// 3) 按需：在 seed_demo.sql 增加默认行，或在管理台新建任务并选用已注册 job_key
//
// 注意：
// - 任务要幂等（重复执行不应产生脏数据）
// - 失败要写清 error_code / error_message
// - summary 尽量写可读统计，方便运营排查

// YourJobPayload 示例：你的任务参数结构
type YourJobPayload struct {
	BatchSize int64 `json:"batch_size"`
	DryRun    bool  `json:"dry_run"`
}

// execYourJob 示例：最小任务处理模板
func (r *Runner) execYourJob(ctx context.Context, run store.Run, job *store.Job) {
	// 1) 默认参数
	p := YourJobPayload{
		BatchSize: 200,
		DryRun:    false,
	}

	// 2) 合并 payload_json（可选）
	if strings.TrimSpace(job.PayloadJSON) != "" {
		_ = json.Unmarshal([]byte(job.PayloadJSON), &p)
	}
	if p.BatchSize <= 0 {
		p.BatchSize = 200
	}

	// 3) 执行你的业务逻辑（示例占位）
	processed := int64(0)
	affected := int64(0)
	_ = processed
	_ = affected

	// 4) 记录成功日志（示例）
	_ = r.svcCtx.Store.FinishRun(
		ctx,
		run.ID,
		run.JobID,
		store.JobStatusSuccess,
		fmt.Sprintf("processed=%d affected=%d dry_run=%v", processed, affected, p.DryRun),
		"",
		"",
		map[string]any{
			"processed": processed,
			"affected":  affected,
			"dry_run":   p.DryRun,
		},
	)
}
