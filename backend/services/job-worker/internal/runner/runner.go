package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gloopai/pay/common/jobkeys"
	"github.com/gloopai/pay/job-worker/internal/store"
	"github.com/gloopai/pay/job-worker/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type Runner struct {
	svcCtx    *svc.ServiceContext
	jobRoutes map[string]func(context.Context, store.Run, *store.Job)
}

func New(svcCtx *svc.ServiceContext) *Runner {
	r := &Runner{
		svcCtx:    svcCtx,
		jobRoutes: map[string]func(context.Context, store.Run, *store.Job){},
	}
	// 统一注册点：后续新增任务只需要在这里加一行。
	// 示例模板见 jobs_template.go
	r.registerJob(jobkeys.TestLogHeartbeat, r.execTestLogHeartbeat)
	r.registerJob(jobkeys.AdminOperationLogsCleanup, r.execAdminOperationLogsCleanup)
	r.assertJobKeysMatchJobkeys()
	return r
}

// assertJobKeysMatchJobkeys 保证 runner 的注册表与 common/jobkeys.RegisteredKeys() 完全一致，避免管理台下拉与真实 handler 漂移。
func (r *Runner) assertJobKeysMatchJobkeys() {
	list := jobkeys.RegisteredKeys()
	want := make(map[string]struct{}, len(list))
	for _, k := range list {
		want[k] = struct{}{}
	}
	for k := range r.jobRoutes {
		if _, ok := want[k]; !ok {
			logx.Errorf("runner registered job_key %q not listed in jobkeys.RegisteredKeys()", k)
			os.Exit(1)
		}
	}
	for _, k := range list {
		if _, ok := r.jobRoutes[k]; !ok {
			logx.Errorf("jobkeys.RegisteredKeys() contains %q but runner did not register it", k)
			os.Exit(1)
		}
	}
}

func (r *Runner) registerJob(jobKey string, handler func(context.Context, store.Run, *store.Job)) {
	key := strings.TrimSpace(jobKey)
	if key == "" || handler == nil {
		return
	}
	r.jobRoutes[key] = handler
}

func (r *Runner) Start() {
	workerID := strings.TrimSpace(r.svcCtx.Config.Worker.ID)
	if workerID == "" {
		host, _ := os.Hostname()
		workerID = fmt.Sprintf("job-worker-%s", host)
	}
	pollSec := r.svcCtx.Config.Worker.PollIntervalSeconds
	if pollSec <= 0 {
		pollSec = 5
	}
	maxClaim := r.svcCtx.Config.Worker.MaxClaimPerTick
	if maxClaim <= 0 {
		maxClaim = 10
	}
	maxEnqueue := r.svcCtx.Config.Worker.MaxEnqueueDuePerTick
	if maxEnqueue <= 0 {
		maxEnqueue = 50
	}
	logx.Infof("job-worker started, worker_id=%s poll=%ds", workerID, pollSec)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	hbSec := r.svcCtx.Config.Worker.HeartbeatIntervalSeconds
	if hbSec <= 0 {
		hbSec = 15
	}
	go r.runHeartbeat(ctx, workerID, time.Duration(hbSec)*time.Second)
	tk := time.NewTicker(time.Duration(pollSec) * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			logx.Info("job-worker stopping")
			return
		case <-tk.C:
			if sec := r.svcCtx.Config.Worker.StaleRunningSeconds; sec > 0 {
				n, err := r.svcCtx.Store.ResetStaleRunningRuns(ctx, time.Duration(sec)*time.Second)
				if err != nil {
					logx.Errorf("reset stale running runs failed: %v", err)
				} else if n > 0 {
					logx.Infof("reset stale running runs: %d", n)
				}
			}
			enq, enqErr := r.svcCtx.Store.EnqueueDueJobs(ctx, maxEnqueue)
			if enqErr != nil {
				logx.Errorf("enqueue due jobs failed: %v", enqErr)
			}
			runs, err := r.svcCtx.Store.ClaimQueuedRuns(ctx, workerID, maxClaim)
			if err != nil {
				logx.Errorf("claim queued runs failed: %v", err)
				continue
			}
			logx.Infof("poll tick: enqueue_scheduled=%d claimed_runs=%d", enq, len(runs))
			for _, run := range runs {
				logx.Infof("run start run_id=%d job_id=%d job_key=%s", run.ID, run.JobID, run.JobKey)
				r.executeRun(ctx, run)
			}
		}
	}
}

func (r *Runner) runHeartbeat(ctx context.Context, workerID string, every time.Duration) {
	host, _ := os.Hostname()
	send := func() {
		if err := r.svcCtx.Store.UpsertHeartbeat(ctx, workerID, host); err != nil {
			logx.Errorf("job worker heartbeat failed: %v", err)
		}
	}
	send()
	tk := time.NewTicker(every)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			send()
		}
	}
}

func (r *Runner) executeRun(ctx context.Context, run store.Run) {
	job, err := r.svcCtx.Store.GetJobByID(ctx, run.JobID)
	if err != nil {
		_ = r.svcCtx.Store.FinishRun(ctx, run.ID, run.JobID, store.JobStatusFailed, "job not found", "JOB_NOT_FOUND", err.Error(), nil)
		return
	}
	h, ok := r.jobRoutes[job.JobKey]
	if !ok {
		_ = r.svcCtx.Store.FinishRun(ctx, run.ID, run.JobID, store.JobStatusSkipped, "unknown job key", "UNKNOWN_JOB", job.JobKey, nil)
		return
	}
	h(ctx, run, job)
}

type adminOperationLogsCleanupPayload struct {
	RetentionDays int64 `json:"retention_days"`
}

// execTestLogHeartbeat 测试任务：打一条带时间戳的日志，并把时间写入 run 的 summary/output_json。
func (r *Runner) execTestLogHeartbeat(ctx context.Context, run store.Run, job *store.Job) {
	now := time.Now()
	summary := fmt.Sprintf("test heartbeat at %s", now.Format(time.RFC3339Nano))
	logx.Infof("scheduled_job test_log_heartbeat job_key=%s run_id=%d %s", job.JobKey, run.ID, summary)
	_ = r.svcCtx.Store.FinishRun(ctx, run.ID, run.JobID, store.JobStatusSuccess, summary, "", "", map[string]any{
		"logged_at": now.UTC().Format(time.RFC3339Nano),
		"run_id":    run.ID,
		"job_key":   job.JobKey,
	})
}

func (r *Runner) execAdminOperationLogsCleanup(ctx context.Context, run store.Run, job *store.Job) {
	p := adminOperationLogsCleanupPayload{RetentionDays: 30}
	if strings.TrimSpace(job.PayloadJSON) != "" {
		_ = json.Unmarshal([]byte(job.PayloadJSON), &p)
	}
	if p.RetentionDays <= 0 {
		p.RetentionDays = 30
	}
	cutoff := time.Now().Add(-time.Duration(p.RetentionDays) * 24 * time.Hour)
	deleted, err := r.svcCtx.Store.DeleteAdminOperationLogsBefore(ctx, cutoff)
	if err != nil {
		_ = r.svcCtx.Store.FinishRun(ctx, run.ID, run.JobID, store.JobStatusFailed, "cleanup admin_operation_logs failed", "DB_ERROR", err.Error(), nil)
		return
	}
	_ = r.svcCtx.Store.FinishRun(ctx, run.ID, run.JobID, store.JobStatusSuccess, fmt.Sprintf("deleted=%d retention_days=%d", deleted, p.RetentionDays), "", "", map[string]any{
		"deleted":        deleted,
		"retention_days": p.RetentionDays,
		"cutoff":         cutoff.UTC().Format(time.RFC3339),
	})
}
