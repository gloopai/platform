// Package jobworkerclient is the gRPC client for JobWorkerRuntime (registered on pay core with jobworkergrpc.Register).
// Worker binaries should only depend on this package + zrpc, not reimplement RPC calls.
package jobworkerclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	jobworkerpb "github.com/gloopai/platform/common/pb/jobworker"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrJobNotFound is returned when GetScheduledJob finds no row.
var ErrJobNotFound = errors.New("scheduled job not found")

// Run is a claimed scheduled_job_run row (execution plane).
type Run struct {
	ID      int64
	JobID   int64
	JobKey  string
	Attempt int64
}

// Job is the scheduled_jobs row fields needed by workers.
type Job struct {
	ID              int64
	JobKey          string
	ScheduleType    string
	CronExpr        string
	IntervalSeconds int64
	Timezone        string
	PayloadJSON     string
}

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Runtime is the worker-side API for enqueue/claim/heartbeat/finish (JobWorkerRuntime gRPC).
type Runtime interface {
	UpsertHeartbeat(ctx context.Context, workerID, hostname string) error
	ResetStaleRunningRuns(ctx context.Context, maxAge time.Duration) (int64, error)
	EnqueueDueJobs(ctx context.Context, limit int64) (int64, error)
	ClaimQueuedRuns(ctx context.Context, workerID string, limit int64) ([]Run, error)
	GetJobByID(ctx context.Context, id int64) (*Job, error)
	FinishRun(ctx context.Context, runID, jobID int64, stStatus, summary, errCode, errMessage string, output any) error
}

type defaultClient struct {
	cli jobworkerpb.JobWorkerRuntimeClient
}

// New returns a [Runtime] using the given zrpc client (e.g. pay core).
func New(cli zrpc.Client) Runtime {
	return &defaultClient{cli: jobworkerpb.NewJobWorkerRuntimeClient(cli.Conn())}
}

func (d *defaultClient) UpsertHeartbeat(ctx context.Context, workerID, hostname string) error {
	_, err := d.cli.Heartbeat(ctx, &jobworkerpb.HeartbeatReq{WorkerId: workerID, Hostname: hostname})
	return err
}

func (d *defaultClient) ResetStaleRunningRuns(ctx context.Context, maxAge time.Duration) (int64, error) {
	if maxAge <= 0 {
		return 0, nil
	}
	sec := int64(maxAge / time.Second)
	if sec <= 0 {
		sec = 1
	}
	resp, err := d.cli.ResetStaleRunningRuns(ctx, &jobworkerpb.ResetStaleRunningRunsReq{MaxAgeSeconds: sec})
	if err != nil {
		return 0, err
	}
	return resp.GetRowsAffected(), nil
}

func (d *defaultClient) EnqueueDueJobs(ctx context.Context, limit int64) (int64, error) {
	resp, err := d.cli.EnqueueDueJobs(ctx, &jobworkerpb.EnqueueDueJobsReq{Limit: limit})
	if err != nil {
		return 0, err
	}
	return resp.GetEnqueued(), nil
}

func (d *defaultClient) ClaimQueuedRuns(ctx context.Context, workerID string, limit int64) ([]Run, error) {
	resp, err := d.cli.ClaimQueuedRuns(ctx, &jobworkerpb.ClaimQueuedRunsReq{WorkerId: workerID, Limit: limit})
	if err != nil {
		return nil, err
	}
	var out []Run
	for _, r := range resp.GetRuns() {
		out = append(out, Run{
			ID:      r.GetRunId(),
			JobID:   r.GetJobId(),
			JobKey:  r.GetJobKey(),
			Attempt: r.GetAttempt(),
		})
	}
	return out, nil
}

func (d *defaultClient) GetJobByID(ctx context.Context, id int64) (*Job, error) {
	resp, err := d.cli.GetScheduledJob(ctx, &jobworkerpb.GetScheduledJobReq{JobId: id})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	j := resp.GetJob()
	if j == nil {
		return nil, ErrJobNotFound
	}
	return &Job{
		ID:              j.GetId(),
		JobKey:          j.GetJobKey(),
		ScheduleType:    j.GetScheduleType(),
		CronExpr:        j.GetCronExpr(),
		IntervalSeconds: j.GetIntervalSeconds(),
		Timezone:        j.GetTimezone(),
		PayloadJSON:     j.GetPayloadJson(),
	}, nil
}

func (d *defaultClient) FinishRun(ctx context.Context, runID, jobID int64, stStatus, summary, errCode, errMessage string, output any) error {
	outJSON := ""
	if output != nil {
		b, err := json.Marshal(output)
		if err != nil {
			return err
		}
		outJSON = string(b)
	}
	_, err := d.cli.FinishJobRun(ctx, &jobworkerpb.FinishJobRunReq{
		RunId:        runID,
		JobId:        jobID,
		Status:       stStatus,
		Summary:      summary,
		ErrorCode:    errCode,
		ErrorMessage: errMessage,
		OutputJson:   outJSON,
	})
	return err
}
