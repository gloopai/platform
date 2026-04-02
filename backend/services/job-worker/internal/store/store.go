package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Job struct {
	ID              int64
	JobKey          string
	ScheduleType    string
	CronExpr        string
	IntervalSeconds int64
	Timezone        string
	PayloadJSON     string
}

type Run struct {
	ID      int64
	JobID   int64
	JobKey  string
	Attempt int64
}

const (
	JobStatusSuccess = "success"
	JobStatusFailed  = "failed"
	JobStatusSkipped = "skipped"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// UpsertHeartbeat 由 job-worker 周期调用，供管理台列出节点与聚合负载。
func (s *Store) UpsertHeartbeat(ctx context.Context, workerID, hostname string) error {
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

func (s *Store) EnqueueDueJobs(ctx context.Context, limit int64) (int64, error) {
	if limit <= 0 {
		limit = 100
	}
	type due struct{ ID int64 }
	var rows []due
	// 仅当该任务当前没有 queued/running 时才入队，避免 next_run_at 在完成前未推进导致每轮轮询重复 INSERT、队列堆积。
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
	var n int64
	for _, r := range rows {
		if err := s.db.WithContext(ctx).Exec(
			`INSERT INTO scheduled_job_runs (job_id, trigger_type, scheduled_at, status, attempt)
			 VALUES (?, 'scheduler', NOW(), 'queued', 1)`, r.ID,
		).Error; err == nil {
			n++
		}
	}
	return n, nil
}

// ResetStaleRunningRuns 将长时间未结束的 running 标记为失败，避免 worker 崩溃后 forbid 策略永久卡死同任务队列。
func (s *Store) ResetStaleRunningRuns(ctx context.Context, maxAge time.Duration) (int64, error) {
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

func (s *Store) ClaimQueuedRuns(ctx context.Context, workerID string, limit int64) ([]Run, error) {
	if strings.TrimSpace(workerID) == "" {
		return nil, errors.New("worker_id required")
	}
	if limit <= 0 {
		limit = 20
	}
	var out []Run
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
			out = []Run{}
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
		var rows []Run
		if err := tx.Raw(
			`SELECT r.id, r.job_id, j.job_key, r.attempt
			   FROM scheduled_job_runs r
			   JOIN scheduled_jobs j ON j.id = r.job_id
			  WHERE r.id IN ? ORDER BY r.id ASC`, idsOnly,
		).Scan(&rows).Error; err != nil {
			return err
		}
		out = rows
		return nil
	})
	return out, err
}

func (s *Store) GetJobByID(ctx context.Context, id int64) (*Job, error) {
	var out Job
	if err := s.db.WithContext(ctx).Raw(
		`SELECT id, job_key, schedule_type, cron_expr, interval_seconds, timezone, payload_json
		   FROM scheduled_jobs
		  WHERE id = ? LIMIT 1`, id,
	).Scan(&out).Error; err != nil {
		return nil, err
	}
	if out.ID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &out, nil
}

func (s *Store) DeleteAdminOperationLogsBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	res := s.db.WithContext(ctx).Exec(
		`DELETE FROM admin_operation_logs WHERE created_at < ?`,
		cutoff,
	)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *Store) FinishRun(ctx context.Context, runID, jobID int64, status, summary, errCode, errMessage string, output any) error {
	outJSON := ""
	if output != nil {
		b, _ := json.Marshal(output)
		outJSON = string(b)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var startAt time.Time
		tx.Raw(`SELECT COALESCE(started_at, NOW()) FROM scheduled_job_runs WHERE id = ? LIMIT 1`, runID).Scan(&startAt)
		dur := time.Since(startAt).Milliseconds()
		if dur < 0 {
			dur = 0
		}
		if err := tx.Exec(
			`UPDATE scheduled_job_runs
			    SET status = ?, finished_at = NOW(), duration_ms = ?, summary = ?, error_code = ?, error_message = ?, output_json = ?
			  WHERE id = ?`,
			status, dur, truncate(summary, 255), truncate(errCode, 64), truncate(errMessage, 255), outJSON, runID,
		).Error; err != nil {
			return err
		}
		job, err := s.GetJobByID(ctx, jobID)
		if err != nil {
			return err
		}
		nextRunAt, err := calcNextRunAt(job, time.Now())
		if err != nil {
			nextRunAt = time.Now().Add(time.Minute)
		}
		return tx.Exec(
			`UPDATE scheduled_jobs
			    SET last_run_at = NOW(), last_status = ?, last_error = ?,
			        next_run_at = ?
			  WHERE id = ?`,
			status, truncate(errMessage, 255), nextRunAt, jobID,
		).Error
	})
}

func calcNextRunAt(job *Job, now time.Time) (time.Time, error) {
	if job == nil {
		return time.Time{}, errors.New("job required")
	}
	loc := time.Local
	if z := strings.TrimSpace(job.Timezone); z != "" {
		if l, err := time.LoadLocation(z); err == nil {
			loc = l
		}
	}
	cur := now.In(loc)
	st := strings.TrimSpace(job.ScheduleType)
	if st == "" {
		st = "fixed_interval"
	}
	switch st {
	case "fixed_interval":
		sec := job.IntervalSeconds
		if sec <= 0 {
			sec = 60
		}
		return cur.Add(time.Duration(sec) * time.Second).In(time.Local), nil
	case "hourly":
		minute, err := parseMinute(job.CronExpr)
		if err != nil {
			return time.Time{}, err
		}
		next := time.Date(cur.Year(), cur.Month(), cur.Day(), cur.Hour(), minute, 0, 0, loc)
		if !next.After(cur) {
			next = next.Add(time.Hour)
		}
		return next.In(time.Local), nil
	case "daily":
		h, m, err := parseHourMinute(job.CronExpr)
		if err != nil {
			return time.Time{}, err
		}
		next := time.Date(cur.Year(), cur.Month(), cur.Day(), h, m, 0, 0, loc)
		if !next.After(cur) {
			next = next.Add(24 * time.Hour)
		}
		return next.In(time.Local), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported schedule_type: %s", st)
	}
}

func parseMinute(v string) (int, error) {
	v = strings.TrimSpace(v)
	var m int
	_, err := fmt.Sscanf(v, "%d", &m)
	if err != nil || m < 0 || m > 59 {
		return 0, errors.New("cron_expr minute must be 0..59")
	}
	return m, nil
}

func parseHourMinute(v string) (int, int, error) {
	v = strings.TrimSpace(v)
	var h, m int
	_, err := fmt.Sscanf(v, "%d:%d", &h, &m)
	if err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, errors.New("cron_expr must be HH:MM")
	}
	return h, m, nil
}

func truncate(v string, n int) string {
	v = strings.TrimSpace(v)
	if n <= 0 || len(v) <= n {
		return v
	}
	return v[:n]
}
