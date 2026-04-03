package jobworker

import (
	"context"
	"errors"
	"time"

	jobworkerpb "github.com/gloopai/platform/common/pb/jobworker"
	"github.com/gloopai/platform/service-hub/internal/store"
	"gorm.io/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements [jobworkerpb.JobWorkerRuntimeServer] for job-worker processes (execution plane).
type Server struct {
	jobworkerpb.UnimplementedJobWorkerRuntimeServer
	st *store.ScheduledJobsStore
}

func NewServer(st *store.ScheduledJobsStore) *Server {
	return &Server{st: st}
}

func (s *Server) Heartbeat(ctx context.Context, req *jobworkerpb.HeartbeatReq) (*jobworkerpb.HeartbeatResp, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if err := s.st.UpsertHeartbeat(ctx, req.GetWorkerId(), req.GetHostname()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &jobworkerpb.HeartbeatResp{}, nil
}

func (s *Server) ResetStaleRunningRuns(ctx context.Context, req *jobworkerpb.ResetStaleRunningRunsReq) (*jobworkerpb.ResetStaleRunningRunsResp, error) {
	if req == nil || req.GetMaxAgeSeconds() <= 0 {
		return &jobworkerpb.ResetStaleRunningRunsResp{}, nil
	}
	n, err := s.st.ResetStaleRunningRuns(ctx, time.Duration(req.GetMaxAgeSeconds())*time.Second)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &jobworkerpb.ResetStaleRunningRunsResp{RowsAffected: n}, nil
}

func (s *Server) EnqueueDueJobs(ctx context.Context, req *jobworkerpb.EnqueueDueJobsReq) (*jobworkerpb.EnqueueDueJobsResp, error) {
	limit := int64(100)
	if req != nil && req.GetLimit() > 0 {
		limit = req.GetLimit()
	}
	n, err := s.st.EnqueueDueJobs(ctx, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &jobworkerpb.EnqueueDueJobsResp{Enqueued: n}, nil
}

func (s *Server) ClaimQueuedRuns(ctx context.Context, req *jobworkerpb.ClaimQueuedRunsReq) (*jobworkerpb.ClaimQueuedRunsResp, error) {
	if req == nil || req.GetWorkerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "worker_id required")
	}
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	runs, err := s.st.ClaimQueuedRuns(ctx, req.GetWorkerId(), limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := &jobworkerpb.ClaimQueuedRunsResp{}
	for _, r := range runs {
		out.Runs = append(out.Runs, &jobworkerpb.ClaimedRun{
			RunId:   r.ID,
			JobId:   r.JobID,
			JobKey:  r.JobKey,
			Attempt: r.Attempt,
		})
	}
	return out, nil
}

func (s *Server) GetScheduledJob(ctx context.Context, req *jobworkerpb.GetScheduledJobReq) (*jobworkerpb.GetScheduledJobResp, error) {
	if req == nil || req.GetJobId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "job_id required")
	}
	j, err := s.st.GetJobByID(ctx, req.GetJobId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &jobworkerpb.GetScheduledJobResp{
		Job: &jobworkerpb.ScheduledJobRow{
			Id:              j.ID,
			JobKey:          j.JobKey,
			ScheduleType:    j.ScheduleType,
			CronExpr:        j.CronExpr,
			IntervalSeconds: j.IntervalSeconds,
			Timezone:        j.Timezone,
			PayloadJson:     j.PayloadJSON,
		},
	}, nil
}

func (s *Server) FinishJobRun(ctx context.Context, req *jobworkerpb.FinishJobRunReq) (*jobworkerpb.FinishJobRunResp, error) {
	if req == nil || req.GetRunId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "run_id required")
	}
	if err := s.st.FinishRunWithOutputJSON(ctx, req.GetRunId(), req.GetStatus(), req.GetSummary(), req.GetErrorCode(), req.GetErrorMessage(), req.GetOutputJson()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &jobworkerpb.FinishJobRunResp{}, nil
}
