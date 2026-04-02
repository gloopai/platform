package handler

import (
	"net/http"

	"github.com/gloopai/platform/common/jobkeys"
	"github.com/gloopai/platform/gateway/internal/apiresp"
	"github.com/gloopai/platform/gateway/internal/logic"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/gloopai/platform/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminListJobWorkerNodesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.ListJobWorkerNodes()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminListScheduledJobKeysHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		apiresp.OK(w, types.AdminScheduledJobKeysResp{
			Keys:    jobkeys.RegisteredKeys(),
			Pattern: `^[a-z][a-z0-9_]{0,62}$`,
		})
	}
}

func AdminListScheduledJobsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminScheduledJobsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.ListScheduledJobs(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminCreateScheduledJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCreateScheduledJobReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.CreateScheduledJob(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminUpdateScheduledJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminUpdateScheduledJobReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.UpdateScheduledJob(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminToggleScheduledJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminToggleScheduledJobReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.ToggleScheduledJob(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminRunScheduledJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminRunScheduledJobReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.RunScheduledJobNow(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminListScheduledJobRunsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminScheduledJobRunsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.ListScheduledJobRuns(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminGetScheduledJobRunHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminScheduledJobRunIdReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.GetScheduledJobRun(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func AdminRetryScheduledJobRunHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminScheduledJobRunIdReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminJobs(r.Context(), svcCtx)
		resp, err := l.RetryScheduledJobRun(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}
