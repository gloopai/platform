// 管理台：系统管理
package handler

import (
	"net/http"

	"github.com/gloopai/pay/gateway/internal/apiresp"
	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminListUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.ListAdminUsers()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminDisplaySettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminDisplaySettingsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.GetDisplaySettings(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminCreateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCreateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.CreateAdminUser(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminUpdateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminUpdateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.UpdateAdminUser(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminDeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminDeleteUserReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.DeleteAdminUser(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminResetUserPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminResetUserPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.ResetAdminUserPassword(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminMfaSetupHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminMfaSetupReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.SetupAdminUserMfa(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminMfaConfirmHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminMfaConfirmReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.ConfirmAdminUserMfa(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminMfaDisableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminMfaDisableReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.DisableAdminUserMfa(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminUpdateDisplaySettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminDisplaySettingsUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSystem(r.Context(), svcCtx)
		resp, err := l.UpdateDisplaySettings(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
