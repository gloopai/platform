// 管理台：RBAC（菜单级）
package handler

import (
	"net/http"
	"strings"

	"github.com/gloopai/pay/gateway/internal/apiresp"
	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminMyMenuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminRbac(r.Context(), svcCtx)
		resp, err := l.MyMenu()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

// requireRbacManage 以“是否拥有 RBAC 配置相关菜单”为管理权限准入（含历史单页 menu.rbac 与拆分后的 menu.rbac_*）。
func requireRbacManage(svcCtx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) bool {
	adminID := middleware.AdminIdFromContext(r.Context())
	if adminID <= 0 {
		apiresp.Fail(w, apiresp.CodeUnauthorized, "unauthorized")
		return false
	}
	menus, err := svcCtx.ServiceHub.GetAdminRbacMyMenus(r.Context(), adminID)
	if err != nil {
		apiresp.Fail(w, apiresp.CodeInternal, err.Error())
		return false
	}
	for _, m := range menus {
		if m == nil {
			continue
		}
		k := strings.TrimSpace(m.GetMenuKey())
		if k == "menu.rbac" || strings.HasPrefix(k, "menu.rbac_") {
			return true
		}
	}
	apiresp.Fail(w, apiresp.CodeForbidden, "forbidden")
	return false
}

func AdminListRbacRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		roles, err := svcCtx.ServiceHub.ListAdminRoles(r.Context())
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"roles": roles})
	}
}

type createRoleReq struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status int64  `json:"status"`
}

func AdminCreateRbacRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req createRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		role, err := svcCtx.ServiceHub.CreateAdminRole(r.Context(), req.Code, req.Name, req.Status)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"role": role})
	}
}

type updateRoleReq struct {
	ID     int64  `path:"id"`
	Name   string `json:"name"`
	Status int64  `json:"status"`
}

type idPath struct {
	ID int64 `path:"id"`
}

func AdminUpdateRbacRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req updateRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		role, err := svcCtx.ServiceHub.UpdateAdminRole(r.Context(), req.ID, req.Name, req.Status)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"role": role})
	}
}

func AdminDeleteRbacRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var p idPath
		if err := httpx.Parse(r, &p); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.DeleteAdminRole(r.Context(), p.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

func AdminListRbacMenusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		menus, err := svcCtx.ServiceHub.ListAdminMenus(r.Context())
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"menus": menus})
	}
}

type createMenuReq struct {
	ParentID  int64  `json:"parent_id"`
	MenuKey   string `json:"menu_key"`
	Label     string `json:"label"`
	Icon      string `json:"icon"`
	Kind      int64  `json:"kind"`
	Path      string `json:"path"`
	SortOrder int64  `json:"sort_order"`
	Placement string `json:"placement"`
}

func AdminCreateRbacMenuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req createMenuReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		m, err := svcCtx.ServiceHub.CreateAdminMenu(r.Context(), req.ParentID, req.MenuKey, req.Label, req.Icon, req.Kind, req.Path, req.SortOrder, req.Placement)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"menu": m})
	}
}

type updateMenuReq struct {
	ID        int64  `path:"id"`
	ParentID  int64  `json:"parent_id"`
	MenuKey   string `json:"menu_key"`
	Label     string `json:"label"`
	Icon      string `json:"icon"`
	Kind      int64  `json:"kind"`
	Path      string `json:"path"`
	SortOrder int64  `json:"sort_order"`
	Placement string `json:"placement"`
}

func AdminUpdateRbacMenuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req updateMenuReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		m, err := svcCtx.ServiceHub.UpdateAdminMenu(r.Context(), req.ID, req.ParentID, req.MenuKey, req.Label, req.Icon, req.Kind, req.Path, req.SortOrder, req.Placement)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"menu": m})
	}
}

func AdminDeleteRbacMenuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var p idPath
		if err := httpx.Parse(r, &p); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.DeleteAdminMenu(r.Context(), p.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

func AdminGetRbacRoleMenusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var p idPath
		if err := httpx.Parse(r, &p); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		menuIDs, err := svcCtx.ServiceHub.GetAdminRoleMenus(r.Context(), p.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"menu_ids": menuIDs})
	}
}

type setRoleMenusReq struct {
	ID      int64   `path:"id"`
	MenuIDs []int64 `json:"menu_ids"`
}

func AdminSetRbacRoleMenusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req setRoleMenusReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.SetAdminRoleMenus(r.Context(), req.ID, req.MenuIDs)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

func AdminGetRbacUserRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p idPath
		if err := httpx.Parse(r, &p); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		roleIDs, err := svcCtx.ServiceHub.GetAdminUserRoles(r.Context(), p.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"role_ids": roleIDs})
	}
}

type setUserRolesReq struct {
	ID      int64   `path:"id"`
	RoleIDs []int64 `json:"role_ids"`
}

func AdminSetRbacUserRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setUserRolesReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.SetAdminUserRoles(r.Context(), req.ID, req.RoleIDs)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

// ---- permissions config ----

func AdminListRbacPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		rows, total, err := svcCtx.ServiceHub.ListAdminPermissions(r.Context(), 0, 0, "", "")
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"permissions": rows, "total": total})
	}
}

type createPermReq struct {
	PermKey  string `json:"perm_key"`
	Label    string `json:"label"`
	Category string `json:"category"`
	MenuKey  string `json:"menu_key"`
	Status   int64  `json:"status"`
}

func AdminCreateRbacPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req createPermReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		p, err := svcCtx.ServiceHub.CreateAdminPermission(r.Context(), req.PermKey, req.Label, req.Category, req.MenuKey, req.Status)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"permission": p})
	}
}

type updatePermReq struct {
	ID       int64  `path:"id"`
	Label    string `json:"label"`
	Category string `json:"category"`
	MenuKey  string `json:"menu_key"`
	Status   int64  `json:"status"`
}

func AdminUpdateRbacPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req updatePermReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		p, err := svcCtx.ServiceHub.UpdateAdminPermission(r.Context(), req.ID, req.Label, req.Category, req.MenuKey, req.Status)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"permission": p})
	}
}

func AdminDeleteRbacPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var pth idPath
		if err := httpx.Parse(r, &pth); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.DeleteAdminPermission(r.Context(), pth.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

// ---- role permissions ----

func AdminGetRbacRolePermKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var pth idPath
		if err := httpx.Parse(r, &pth); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		keys, err := svcCtx.ServiceHub.GetAdminRolePermKeys(r.Context(), pth.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"perm_keys": keys})
	}
}

type setRolePermKeysReq struct {
	ID       int64    `path:"id"`
	PermKeys []string `json:"perm_keys"`
}

func AdminSetRbacRolePermKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req setRolePermKeysReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.SetAdminRolePermKeys(r.Context(), req.ID, req.PermKeys)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}

// ---- api rules ----

func AdminListRbacApiRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		rows, total, err := svcCtx.ServiceHub.ListAdminApiRules(r.Context(), 0, 0, "", "")
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"rules": rows, "total": total})
	}
}

type upsertApiRuleReq struct {
	Method      string `json:"method"`
	PathPattern string `json:"path_pattern"`
	PermKey     string `json:"perm_key"`
	Status      int64  `json:"status"`
	Remark      string `json:"remark"`
}

func AdminUpsertRbacApiRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var req upsertApiRuleReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		row, err := svcCtx.ServiceHub.UpsertAdminApiRule(r.Context(), req.Method, req.PathPattern, req.PermKey, req.Remark, req.Status)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"rule": row})
	}
}

func AdminDeleteRbacApiRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireRbacManage(svcCtx, w, r) {
			return
		}
		var pth idPath
		if err := httpx.Parse(r, &pth); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		ok, err := svcCtx.ServiceHub.DeleteAdminApiRule(r.Context(), pth.ID)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]any{"ok": ok})
	}
}
