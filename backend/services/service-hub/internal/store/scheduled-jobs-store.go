package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gloopai/platform/common/jobkeys"
	"gorm.io/gorm"
)

const (
	JobStatusQueued  = "queued"
	JobStatusRunning = "running"
	JobStatusSuccess = "success"
	JobStatusFailed  = "failed"
	JobStatusSkipped = "skipped"
	JobStatusTimeout = "timeout"
)

type ScheduledJob struct {
	ID                  int64
	JobKey              string
	Name                string
	Category            string
	Enabled             int64
	Builtin             int64
	ScheduleType        string
	CronExpr            string
	IntervalSeconds     int64
	Timezone            string
	PayloadJSON         string
	ConcurrencyPolicy   string
	MisfirePolicy       string
	MaxRetry            int64
	RetryBackoffSeconds int64
	NextRunAt           sql.NullTime
	LastRunAt           sql.NullTime
	LastStatus          string
	LastError           string
	UpdatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ScheduledJobRun struct {
	ID            int64
	JobID         int64
	TriggerType   string
	ScheduledAt   sql.NullTime
	StartedAt     sql.NullTime
	FinishedAt    sql.NullTime
	DurationMs    int64
	Status        string
	Attempt       int64
	WorkerID      string
	Summary       string
	ErrorCode     string
	ErrorMessage  string
	OutputJSON    string
	CorrelationID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ScheduledJobRunWithJob struct {
	ScheduledJobRun
	JobKey string
	Name   string
}

type ScheduledJobsStore struct {
	db *gorm.DB
}

func NewScheduledJobsStore(db *gorm.DB) *ScheduledJobsStore {
	return &ScheduledJobsStore{db: db}
}

func (s *ScheduledJobsStore) ListJobs(ctx context.Context, limit, offset int64) ([]ScheduledJob, int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Table("scheduled_jobs").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	var out []ScheduledJob
	if err := s.db.WithContext(ctx).
		Table("scheduled_jobs").
		Select(`id, job_key, name, category, enabled, builtin, schedule_type, cron_expr,
			interval_seconds, timezone, payload_json, concurrency_policy, misfire_policy,
			max_retry, retry_backoff_seconds, next_run_at, last_run_at, last_status, last_error,
			updated_by, created_at, updated_at`).
		Order("builtin DESC, id ASC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *ScheduledJobsStore) GetJobByID(ctx context.Context, id int64) (*ScheduledJob, error) {
	if id <= 0 {
		return nil, errors.New("id required")
	}
	var out ScheduledJob
	if err := s.db.WithContext(ctx).
		Table("scheduled_jobs").
		Select(`id, job_key, name, category, enabled, builtin, schedule_type, cron_expr,
			interval_seconds, timezone, payload_json, concurrency_policy, misfire_policy,
			max_retry, retry_backoff_seconds, next_run_at, last_run_at, last_status, last_error,
			updated_by, created_at, updated_at`).
		Where("id = ?", id).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ScheduledJobsStore) CreateJob(ctx context.Context, in *ScheduledJob) (*ScheduledJob, error) {
	if in == nil {
		return nil, errors.New("job required")
	}
	in.JobKey = strings.TrimSpace(in.JobKey)
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, errors.New("name required")
	}
	if err := jobkeys.ValidateJobKey(in.JobKey); err != nil {
		return nil, err
	}
	var dup int64
	if err := s.db.WithContext(ctx).Table("scheduled_jobs").Where("job_key = ?", in.JobKey).Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, errors.New("job_key already exists")
	}
	if in.ScheduleType == "" {
		in.ScheduleType = "fixed_interval"
	}
	if in.ScheduleType != "fixed_interval" && in.ScheduleType != "hourly" && in.ScheduleType != "daily" {
		return nil, errors.New("schedule_type must be fixed_interval/hourly/daily")
	}
	if in.IntervalSeconds <= 0 {
		in.IntervalSeconds = 60
	}
	if in.MaxRetry < 0 {
		in.MaxRetry = 0
	}
	if in.RetryBackoffSeconds < 0 {
		in.RetryBackoffSeconds = 0
	}
	if in.Timezone == "" {
		in.Timezone = "Asia/Shanghai"
	}
	if strings.TrimSpace(in.Category) == "" {
		in.Category = "custom"
	}
	if strings.TrimSpace(in.PayloadJSON) == "" {
		in.PayloadJSON = "{}"
	}
	if in.Enabled != 0 && in.Enabled != 1 {
		in.Enabled = 1
	}
	now := time.Now()
	if !in.NextRunAt.Valid {
		next, err := calcServiceHubNextRunAt(in.ScheduleType, in.CronExpr, in.IntervalSeconds, in.Timezone, now)
		if err != nil {
			return nil, err
		}
		in.NextRunAt = sql.NullTime{Time: next, Valid: true}
	}
	row := map[string]any{
		"job_key":               in.JobKey,
		"name":                  in.Name,
		"category":              strings.TrimSpace(in.Category),
		"enabled":               in.Enabled,
		"builtin":               in.Builtin,
		"schedule_type":         in.ScheduleType,
		"cron_expr":             strings.TrimSpace(in.CronExpr),
		"interval_seconds":      in.IntervalSeconds,
		"timezone":              in.Timezone,
		"payload_json":          strings.TrimSpace(in.PayloadJSON),
		"concurrency_policy":    defaultIfEmpty(strings.TrimSpace(in.ConcurrencyPolicy), "forbid"),
		"misfire_policy":        defaultIfEmpty(strings.TrimSpace(in.MisfirePolicy), "run_once"),
		"max_retry":             in.MaxRetry,
		"retry_backoff_seconds": in.RetryBackoffSeconds,
		"next_run_at":           nullableTimeAny(in.NextRunAt),
		"updated_by":            strings.TrimSpace(in.UpdatedBy),
	}
	if err := s.db.WithContext(ctx).Table("scheduled_jobs").Create(row).Error; err != nil {
		return nil, err
	}
	return s.GetJobByJobKey(ctx, in.JobKey)
}

func (s *ScheduledJobsStore) GetJobByJobKey(ctx context.Context, jobKey string) (*ScheduledJob, error) {
	jobKey = strings.TrimSpace(jobKey)
	if jobKey == "" {
		return nil, errors.New("job_key required")
	}
	var out ScheduledJob
	if err := s.db.WithContext(ctx).
		Table("scheduled_jobs").
		Select(`id, job_key, name, category, enabled, builtin, schedule_type, cron_expr,
			interval_seconds, timezone, payload_json, concurrency_policy, misfire_policy,
			max_retry, retry_backoff_seconds, next_run_at, last_run_at, last_status, last_error,
			updated_by, created_at, updated_at`).
		Where("job_key = ?", jobKey).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ScheduledJobsStore) UpdateJob(ctx context.Context, in *ScheduledJob) (*ScheduledJob, error) {
	if in == nil || in.ID <= 0 {
		return nil, errors.New("id required")
	}
	upd := map[string]any{
		"name":                  strings.TrimSpace(in.Name),
		"category":              strings.TrimSpace(in.Category),
		"schedule_type":         defaultIfEmpty(strings.TrimSpace(in.ScheduleType), "fixed_interval"),
		"cron_expr":             strings.TrimSpace(in.CronExpr),
		"interval_seconds":      maxInt64(in.IntervalSeconds, 1),
		"timezone":              defaultIfEmpty(strings.TrimSpace(in.Timezone), "Asia/Shanghai"),
		"payload_json":          strings.TrimSpace(in.PayloadJSON),
		"concurrency_policy":    defaultIfEmpty(strings.TrimSpace(in.ConcurrencyPolicy), "forbid"),
		"misfire_policy":        defaultIfEmpty(strings.TrimSpace(in.MisfirePolicy), "run_once"),
		"max_retry":             maxInt64(in.MaxRetry, 0),
		"retry_backoff_seconds": maxInt64(in.RetryBackoffSeconds, 0),
		"updated_by":            strings.TrimSpace(in.UpdatedBy),
	}
	if in.NextRunAt.Valid {
		upd["next_run_at"] = in.NextRunAt.Time
	} else {
		next, err := calcServiceHubNextRunAt(in.ScheduleType, in.CronExpr, in.IntervalSeconds, in.Timezone, time.Now())
		if err != nil {
			return nil, err
		}
		upd["next_run_at"] = next
	}
	if err := s.db.WithContext(ctx).
		Table("scheduled_jobs").
		Where("id = ?", in.ID).
		Updates(upd).Error; err != nil {
		return nil, err
	}
	return s.GetJobByID(ctx, in.ID)
}

func (s *ScheduledJobsStore) ToggleJob(ctx context.Context, id int64, enabled bool, updatedBy string) (*ScheduledJob, error) {
	if id <= 0 {
		return nil, errors.New("id required")
	}
	nextRun := "next_run_at"
	upd := map[string]any{
		"enabled":    boolToInt64(enabled),
		"updated_by": strings.TrimSpace(updatedBy),
	}
	if enabled {
		upd[nextRun] = gorm.Expr("IFNULL(next_run_at, DATE_ADD(NOW(), INTERVAL interval_seconds SECOND))")
	}
	if err := s.db.WithContext(ctx).Table("scheduled_jobs").Where("id = ?", id).Updates(upd).Error; err != nil {
		return nil, err
	}
	return s.GetJobByID(ctx, id)
}

func (s *ScheduledJobsStore) QueueJobRun(ctx context.Context, jobID int64, triggerType, correlationID string, attempt int64) (int64, error) {
	if jobID <= 0 {
		return 0, errors.New("job_id required")
	}
	if attempt <= 0 {
		attempt = 1
	}
	if strings.TrimSpace(triggerType) == "" {
		triggerType = "manual"
	}
	res := s.db.WithContext(ctx).Exec(
		`INSERT INTO scheduled_job_runs
		(job_id, trigger_type, scheduled_at, status, attempt, correlation_id)
		VALUES (?, ?, NOW(), 'queued', ?, ?)`,
		jobID, strings.TrimSpace(triggerType), attempt, strings.TrimSpace(correlationID),
	)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *ScheduledJobsStore) ListRuns(ctx context.Context, jobID int64, status, triggerType string, limit, offset int64) ([]ScheduledJobRunWithJob, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := s.db.WithContext(ctx).Table("scheduled_job_runs r").
		Joins("JOIN scheduled_jobs j ON j.id = r.job_id")
	if jobID > 0 {
		q = q.Where("r.job_id = ?", jobID)
	}
	if v := strings.TrimSpace(status); v != "" {
		q = q.Where("r.status = ?", v)
	}
	if v := strings.TrimSpace(triggerType); v != "" {
		q = q.Where("r.trigger_type = ?", v)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []ScheduledJobRunWithJob
	if err := q.Select(`r.id, r.job_id, r.trigger_type, r.scheduled_at, r.started_at, r.finished_at,
		r.duration_ms, r.status, r.attempt, r.worker_id, r.summary, r.error_code, r.error_message,
		r.output_json, r.correlation_id, r.created_at, r.updated_at, j.job_key, j.name`).
		Order("r.id DESC").
		Limit(int(limit)).
		Offset(int(offset)).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// JobWorkerNode 管理台展示用：心跳 + 负载近似（running / 近 1h 成功）。
type JobWorkerNode struct {
	WorkerID        string         `gorm:"column:worker_id"`
	Hostname        string         `gorm:"column:hostname"`
	LastHeartbeatAt sql.NullTime   `gorm:"column:last_heartbeat_at"`
	RunningTasks    int64          `gorm:"column:running_tasks"`
	SuccessLastHour int64          `gorm:"column:success_last_hour"`
}

// UpsertHeartbeat 由 job-worker 经 gRPC 调用，供管理台列出节点。
func (s *ScheduledJobsStore) UpsertHeartbeat(ctx context.Context, workerID, hostname string) error {
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return errors.New("worker_id required")
	}
	return s.db.WithContext(ctx).Exec(
		`INSERT INTO job_worker_heartbeats (worker_id, hostname, last_heartbeat_at) VALUES (?, ?, NOW())
		 ON DUPLICATE KEY UPDATE hostname = VALUES(hostname), last_heartbeat_at = VALUES(last_heartbeat_at)`,
		workerID, strings.TrimSpace(hostname),
	).Error
}

// ResetStaleRunningRuns 将长时间未结束的 running 标记为失败。
func (s *ScheduledJobsStore) ResetStaleRunningRuns(ctx context.Context, maxAge time.Duration) (int64, error) {
	if maxAge <= 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-maxAge)
	res := s.db.WithContext(ctx).Exec(
		`UPDATE scheduled_job_runs
		    SET status = 'failed',
		        finished_at = NOW(),
		        duration_ms = GREATEST(0, TIMESTAMPDIFF(MICROSECOND, started_at, NOW()) DIV 1000),
		        summary = 'stale: worker lost before finish',
		        error_code = 'STALE_RUN',
		        error_message = 'running exceeded stale timeout; reset so queue can proceed'
		  WHERE status = 'running'
		    AND started_at IS NOT NULL
		    AND started_at < ?`, cutoff,
	)
	return res.RowsAffected, res.Error
}

func (s *ScheduledJobsStore) ListJobWorkerNodes(ctx context.Context) ([]JobWorkerNode, int64, error) {
	var queuedTotal int64
	if err := s.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM scheduled_job_runs WHERE status = 'queued'`,
	).Scan(&queuedTotal).Error; err != nil {
		return nil, 0, err
	}
	var rows []JobWorkerNode
	err := s.db.WithContext(ctx).Raw(
		`SELECT h.worker_id AS worker_id, h.hostname AS hostname, h.last_heartbeat_at AS last_heartbeat_at,
			(SELECT COUNT(*) FROM scheduled_job_runs r WHERE r.worker_id = h.worker_id AND r.status = 'running') AS running_tasks,
			(SELECT COUNT(*) FROM scheduled_job_runs r WHERE r.worker_id = h.worker_id AND r.status = 'success' AND r.finished_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)) AS success_last_hour
		   FROM job_worker_heartbeats h
		  ORDER BY h.last_heartbeat_at DESC`,
	).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, queuedTotal, nil
}

func (s *ScheduledJobsStore) GetRunByID(ctx context.Context, id int64) (*ScheduledJobRunWithJob, error) {
	if id <= 0 {
		return nil, errors.New("id required")
	}
	var out ScheduledJobRunWithJob
	if err := s.db.WithContext(ctx).Table("scheduled_job_runs r").
		Joins("JOIN scheduled_jobs j ON j.id = r.job_id").
		Select(`r.id, r.job_id, r.trigger_type, r.scheduled_at, r.started_at, r.finished_at,
		r.duration_ms, r.status, r.attempt, r.worker_id, r.summary, r.error_code, r.error_message,
		r.output_json, r.correlation_id, r.created_at, r.updated_at, j.job_key, j.name`).
		Where("r.id = ?", id).Limit(1).Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ScheduledJobsStore) RetryRun(ctx context.Context, runID int64) (int64, error) {
	r, err := s.GetRunByID(ctx, runID)
	if err != nil {
		return 0, err
	}
	attempt := r.Attempt + 1
	if attempt <= 0 {
		attempt = 1
	}
	return s.QueueJobRun(ctx, r.JobID, "retry", fmt.Sprintf("retry_of_%d", r.ID), attempt)
}

func (s *ScheduledJobsStore) EnqueueDueJobs(ctx context.Context, limit int64) (int64, error) {
	if limit <= 0 {
		limit = 100
	}
	type due struct {
		ID int64
	}
	var rows []due
	if err := s.db.WithContext(ctx).Raw(
		`SELECT j.id FROM scheduled_jobs j
		 WHERE j.enabled = 1
		   AND j.next_run_at IS NOT NULL
		   AND j.next_run_at <= NOW()
		   AND NOT EXISTS (
		     SELECT 1 FROM scheduled_job_runs r
		      WHERE r.job_id = j.id AND r.status IN ('queued', 'running')
		   )
		 ORDER BY j.next_run_at ASC
		 LIMIT ?`, limit,
	).Scan(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	var count int64
	for _, r := range rows {
		if _, err := s.QueueJobRun(ctx, r.ID, "scheduler", "", 1); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *ScheduledJobsStore) ClaimQueuedRuns(ctx context.Context, workerID string, limit int64) ([]ScheduledJobRunWithJob, error) {
	if strings.TrimSpace(workerID) == "" {
		return nil, errors.New("worker_id required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	var out []ScheduledJobRunWithJob
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type rowID struct{ ID int64 }
		var ids []rowID
		if err := tx.Raw(
			`SELECT r.id
			   FROM scheduled_job_runs r
			   JOIN scheduled_jobs j ON j.id = r.job_id
			  WHERE r.status = 'queued'
			    AND (
			      j.enabled = 1
			      OR r.trigger_type IN ('manual', 'retry')
			    )
			    AND (
			      j.concurrency_policy <> 'forbid'
			      OR NOT EXISTS (
			        SELECT 1 FROM scheduled_job_runs rr
			         WHERE rr.job_id = r.job_id AND rr.status = 'running'
			      )
			    )
			  ORDER BY CASE WHEN r.trigger_type IN ('manual', 'retry') THEN 0 ELSE 1 END, r.id ASC
			  LIMIT ?
			  FOR UPDATE`, limit,
		).Scan(&ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			out = []ScheduledJobRunWithJob{}
			return nil
		}
		idsOnly := make([]int64, 0, len(ids))
		for _, x := range ids {
			idsOnly = append(idsOnly, x.ID)
		}
		if err := tx.Exec(
			`UPDATE scheduled_job_runs
			    SET status = 'running', started_at = NOW(), worker_id = ?
			  WHERE id IN ? AND status = 'queued'`, workerID, idsOnly,
		).Error; err != nil {
			return err
		}
		var rows []ScheduledJobRunWithJob
		if err := tx.Table("scheduled_job_runs r").
			Joins("JOIN scheduled_jobs j ON j.id = r.job_id").
			Select(`r.id, r.job_id, r.trigger_type, r.scheduled_at, r.started_at, r.finished_at,
				r.duration_ms, r.status, r.attempt, r.worker_id, r.summary, r.error_code,
				r.error_message, r.output_json, r.correlation_id, r.created_at, r.updated_at,
				j.job_key, j.name`).
			Where("r.id IN ?", idsOnly).
			Order("r.id ASC").
			Find(&rows).Error; err != nil {
			return err
		}
		out = rows
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *ScheduledJobsStore) FinishRun(ctx context.Context, runID int64, status, summary, errCode, errMessage string, output any) error {
	if runID <= 0 {
		return errors.New("run_id required")
	}
	if status == "" {
		status = JobStatusSuccess
	}
	outJSON := ""
	if output != nil {
		b, _ := json.Marshal(output)
		outJSON = string(b)
	}
	return s.finishRunWithOutputJSON(ctx, runID, status, summary, errCode, errMessage, outJSON)
}

// FinishRunWithOutputJSON 供 gRPC 传入已序列化的 output_json（避免 double-json）。
func (s *ScheduledJobsStore) FinishRunWithOutputJSON(ctx context.Context, runID int64, status, summary, errCode, errMessage, outputJSON string) error {
	if runID <= 0 {
		return errors.New("run_id required")
	}
	if status == "" {
		status = JobStatusSuccess
	}
	return s.finishRunWithOutputJSON(ctx, runID, status, summary, errCode, errMessage, strings.TrimSpace(outputJSON))
}

func (s *ScheduledJobsStore) finishRunWithOutputJSON(ctx context.Context, runID int64, status, summary, errCode, errMessage, outJSON string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run ScheduledJobRun
		if err := tx.Table("scheduled_job_runs").
			Select("id, job_id, started_at").
			Where("id = ?", runID).Limit(1).Take(&run).Error; err != nil {
			return err
		}
		duration := int64(0)
		if run.StartedAt.Valid {
			duration = time.Since(run.StartedAt.Time).Milliseconds()
			if duration < 0 {
				duration = 0
			}
		}
		if err := tx.Exec(
			`UPDATE scheduled_job_runs
			    SET status = ?, finished_at = NOW(), duration_ms = ?, summary = ?, error_code = ?, error_message = ?, output_json = ?
			  WHERE id = ?`,
			status, duration, truncate(summary, 255), truncate(errCode, 64), truncate(errMessage, 255), outJSON, runID,
		).Error; err != nil {
			return err
		}
		job, err := s.getScheduledJobForRuntime(tx, run.JobID)
		if err != nil {
			return err
		}
		nextRunAt, nerr := calcServiceHubNextRunAt(job.ScheduleType, job.CronExpr, job.IntervalSeconds, job.Timezone, time.Now())
		if nerr != nil {
			nextRunAt = time.Now().Add(time.Minute)
		}
		return tx.Exec(
			`UPDATE scheduled_jobs
			    SET last_run_at = NOW(), last_status = ?, last_error = ?, next_run_at = ?
			  WHERE id = ?`,
			status, truncate(errMessage, 255), nextRunAt, run.JobID,
		).Error
	})
}

func (s *ScheduledJobsStore) getScheduledJobForRuntime(tx *gorm.DB, jobID int64) (*ScheduledJob, error) {
	if jobID <= 0 {
		return nil, errors.New("job_id required")
	}
	var job ScheduledJob
	if err := tx.Table("scheduled_jobs").
		Select(`id, job_key, schedule_type, cron_expr, interval_seconds, timezone`).
		Where("id = ?", jobID).Limit(1).Take(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func defaultIfEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func maxInt64(v, min int64) int64 {
	if v < min {
		return min
	}
	return v
}

func nullableTimeAny(v sql.NullTime) any {
	if !v.Valid {
		return nil
	}
	return v.Time
}

func truncate(v string, n int) string {
	v = strings.TrimSpace(v)
	if n <= 0 || len(v) <= n {
		return v
	}
	return v[:n]
}

func calcServiceHubNextRunAt(scheduleType, cronExpr string, intervalSec int64, timezone string, now time.Time) (time.Time, error) {
	loc := time.Local
	if z := strings.TrimSpace(timezone); z != "" {
		if l, err := time.LoadLocation(z); err == nil {
			loc = l
		}
	}
	cur := now.In(loc)
	st := strings.TrimSpace(scheduleType)
	if st == "" {
		st = "fixed_interval"
	}
	switch st {
	case "fixed_interval":
		if intervalSec <= 0 {
			intervalSec = 60
		}
		return cur.Add(time.Duration(intervalSec) * time.Second).In(time.Local), nil
	case "hourly":
		var m int
		if _, err := fmt.Sscanf(strings.TrimSpace(cronExpr), "%d", &m); err != nil || m < 0 || m > 59 {
			return time.Time{}, errors.New("hourly cron_expr must be minute(0..59)")
		}
		next := time.Date(cur.Year(), cur.Month(), cur.Day(), cur.Hour(), m, 0, 0, loc)
		if !next.After(cur) {
			next = next.Add(time.Hour)
		}
		return next.In(time.Local), nil
	case "daily":
		var h, m int
		if _, err := fmt.Sscanf(strings.TrimSpace(cronExpr), "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
			return time.Time{}, errors.New("daily cron_expr must be HH:MM")
		}
		next := time.Date(cur.Year(), cur.Month(), cur.Day(), h, m, 0, 0, loc)
		if !next.After(cur) {
			next = next.Add(24 * time.Hour)
		}
		return next.In(time.Local), nil
	default:
		return time.Time{}, errors.New("schedule_type must be fixed_interval/hourly/daily")
	}
}
