package logic

import (
	"context"
	"strings"

	servicehubpb "github.com/gloopai/pay/common/pb/servicehub"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminJobs struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminJobs(ctx context.Context, svcCtx *svc.ServiceContext) *AdminJobs {
	return &AdminJobs{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func mapJob(x *servicehubpb.ScheduledJob) types.AdminScheduledJob {
	if x == nil {
		return types.AdminScheduledJob{}
	}
	return types.AdminScheduledJob{
		Id:                  x.GetId(),
		JobKey:              x.GetJobKey(),
		Name:                x.GetName(),
		Category:            x.GetCategory(),
		Enabled:             x.GetEnabled(),
		Builtin:             x.GetBuiltin(),
		ScheduleType:        x.GetScheduleType(),
		CronExpr:            x.GetCronExpr(),
		IntervalSeconds:     x.GetIntervalSeconds(),
		Timezone:            x.GetTimezone(),
		PayloadJson:         x.GetPayloadJson(),
		ConcurrencyPolicy:   x.GetConcurrencyPolicy(),
		MisfirePolicy:       x.GetMisfirePolicy(),
		MaxRetry:            x.GetMaxRetry(),
		RetryBackoffSeconds: x.GetRetryBackoffSeconds(),
		NextRunAt:           x.GetNextRunAt(),
		LastRunAt:           x.GetLastRunAt(),
		LastStatus:          x.GetLastStatus(),
		LastError:           x.GetLastError(),
		UpdatedBy:           x.GetUpdatedBy(),
	}
}

func mapRun(x *servicehubpb.ScheduledJobRun) types.AdminScheduledJobRun {
	if x == nil {
		return types.AdminScheduledJobRun{}
	}
	return types.AdminScheduledJobRun{
		Id:            x.GetId(),
		JobId:         x.GetJobId(),
		JobKey:        x.GetJobKey(),
		JobName:       x.GetJobName(),
		TriggerType:   x.GetTriggerType(),
		ScheduledAt:   x.GetScheduledAt(),
		StartedAt:     x.GetStartedAt(),
		FinishedAt:    x.GetFinishedAt(),
		DurationMs:    x.GetDurationMs(),
		Status:        x.GetStatus(),
		Attempt:       x.GetAttempt(),
		WorkerId:      x.GetWorkerId(),
		Summary:       x.GetSummary(),
		ErrorCode:     x.GetErrorCode(),
		ErrorMessage:  x.GetErrorMessage(),
		OutputJson:    x.GetOutputJson(),
		CorrelationId: x.GetCorrelationId(),
	}
}

func (a *AdminJobs) ListScheduledJobs(req *types.AdminScheduledJobsReq) (*types.AdminScheduledJobsResp, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	rows, total, err := a.svcCtx.ServiceHub.ListScheduledJobs(a.ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminScheduledJob, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapJob(r))
	}
	return &types.AdminScheduledJobsResp{Jobs: out, Total: total}, nil
}

func (a *AdminJobs) CreateScheduledJob(req *types.AdminCreateScheduledJobReq) (*types.AdminScheduledJob, error) {
	r, err := a.svcCtx.ServiceHub.CreateScheduledJob(a.ctx, &servicehubpb.CreateScheduledJobReq{
		JobKey:              strings.TrimSpace(req.JobKey),
		Name:                strings.TrimSpace(req.Name),
		Category:            strings.TrimSpace(req.Category),
		Enabled:             req.Enabled,
		ScheduleType:        strings.TrimSpace(req.ScheduleType),
		CronExpr:            strings.TrimSpace(req.CronExpr),
		IntervalSeconds:     req.IntervalSeconds,
		Timezone:            strings.TrimSpace(req.Timezone),
		PayloadJson:         strings.TrimSpace(req.PayloadJson),
		ConcurrencyPolicy:   strings.TrimSpace(req.ConcurrencyPolicy),
		MisfirePolicy:       strings.TrimSpace(req.MisfirePolicy),
		MaxRetry:            req.MaxRetry,
		RetryBackoffSeconds: req.RetryBackoffSeconds,
		UpdatedBy:           "admin",
	})
	if err != nil {
		return nil, err
	}
	out := mapJob(r)
	return &out, nil
}

func (a *AdminJobs) UpdateScheduledJob(req *types.AdminUpdateScheduledJobReq) (*types.AdminScheduledJob, error) {
	r, err := a.svcCtx.ServiceHub.UpdateScheduledJob(a.ctx, &servicehubpb.UpdateScheduledJobReq{
		Id:                  req.Id,
		Name:                strings.TrimSpace(req.Name),
		Category:            strings.TrimSpace(req.Category),
		ScheduleType:        strings.TrimSpace(req.ScheduleType),
		CronExpr:            strings.TrimSpace(req.CronExpr),
		IntervalSeconds:     req.IntervalSeconds,
		Timezone:            strings.TrimSpace(req.Timezone),
		PayloadJson:         strings.TrimSpace(req.PayloadJson),
		ConcurrencyPolicy:   strings.TrimSpace(req.ConcurrencyPolicy),
		MisfirePolicy:       strings.TrimSpace(req.MisfirePolicy),
		MaxRetry:            req.MaxRetry,
		RetryBackoffSeconds: req.RetryBackoffSeconds,
		NextRunAt:           req.NextRunAt,
		UpdatedBy:           "admin",
	})
	if err != nil {
		return nil, err
	}
	out := mapJob(r)
	return &out, nil
}

func (a *AdminJobs) ToggleScheduledJob(req *types.AdminToggleScheduledJobReq) (*types.AdminScheduledJob, error) {
	r, err := a.svcCtx.ServiceHub.ToggleScheduledJob(a.ctx, req.Id, req.Enabled, "admin")
	if err != nil {
		return nil, err
	}
	out := mapJob(r)
	return &out, nil
}

func (a *AdminJobs) RunScheduledJobNow(req *types.AdminRunScheduledJobReq) (*types.AdminSimpleOkResp, error) {
	ok, err := a.svcCtx.ServiceHub.RunScheduledJobNow(a.ctx, req.Id, strings.TrimSpace(req.CorrelationId))
	if err != nil {
		return nil, err
	}
	return &types.AdminSimpleOkResp{Ok: ok}, nil
}

func (a *AdminJobs) ListScheduledJobRuns(req *types.AdminScheduledJobRunsReq) (*types.AdminScheduledJobRunsResp, error) {
	rows, total, err := a.svcCtx.ServiceHub.ListScheduledJobRuns(a.ctx, &servicehubpb.ListScheduledJobRunsReq{
		JobId:       req.JobId,
		Status:      strings.TrimSpace(req.Status),
		TriggerType: strings.TrimSpace(req.TriggerType),
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminScheduledJobRun, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapRun(r))
	}
	return &types.AdminScheduledJobRunsResp{Runs: out, Total: total}, nil
}

func (a *AdminJobs) GetScheduledJobRun(req *types.AdminScheduledJobRunIdReq) (*types.AdminScheduledJobRunResp, error) {
	r, err := a.svcCtx.ServiceHub.GetScheduledJobRun(a.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.AdminScheduledJobRunResp{Run: mapRun(r)}, nil
}

func (a *AdminJobs) RetryScheduledJobRun(req *types.AdminScheduledJobRunIdReq) (*types.AdminSimpleOkResp, error) {
	ok, err := a.svcCtx.ServiceHub.RetryScheduledJobRun(a.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.AdminSimpleOkResp{Ok: ok}, nil
}

func mapJobWorkerNode(x *servicehubpb.JobWorkerNode) types.AdminJobWorkerNode {
	if x == nil {
		return types.AdminJobWorkerNode{}
	}
	return types.AdminJobWorkerNode{
		WorkerId:        x.GetWorkerId(),
		Hostname:        x.GetHostname(),
		LastHeartbeatAt: x.GetLastHeartbeatAt(),
		RunningTasks:    x.GetRunningTasks(),
		SuccessLastHour: x.GetSuccessLastHour(),
	}
}

func (a *AdminJobs) ListJobWorkerNodes() (*types.AdminJobWorkerNodesResp, error) {
	nodes, queued, err := a.svcCtx.ServiceHub.ListJobWorkerNodes(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.AdminJobWorkerNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, mapJobWorkerNode(n))
	}
	return &types.AdminJobWorkerNodesResp{Nodes: out, QueuedTotal: queued}, nil
}
