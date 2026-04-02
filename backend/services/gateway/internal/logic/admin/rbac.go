package logic

import (
	"context"
	"sort"
	"strings"

	"github.com/gloopai/platform/gateway/internal/middleware"
	"github.com/gloopai/platform/gateway/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminRbac struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRbac(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRbac {
	return &AdminRbac{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type AdminMenuEntry struct {
	Kind     string               `json:"kind"`
	To       string               `json:"to,omitempty"`
	Label    string               `json:"label"`
	Icon     string               `json:"icon"`
	Key      string               `json:"key,omitempty"`
	Children []AdminMenuChildLink `json:"children,omitempty"`
}

type AdminMenuChildLink struct {
	To    string `json:"to"`
	Label string `json:"label"`
}

// AvatarMenuLink 头像下拉中的入口（仅一级页面）。
type AvatarMenuLink struct {
	To    string `json:"to"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

// MyMenuResp 当前登录管理员可见：侧栏结构 + 头像菜单链接。
type MyMenuResp struct {
	Sidebar     []AdminMenuEntry `json:"sidebar"`
	AvatarLinks []AvatarMenuLink `json:"avatar_links"`
}

func menuPlacementStr(placement string) string {
	p := strings.TrimSpace(strings.ToLower(placement))
	if p == "avatar" {
		return "avatar"
	}
	return "left"
}

// MyMenu 返回当前登录管理员可见菜单（侧栏 + 头像下拉）。
func (a *AdminRbac) MyMenu() (*MyMenuResp, error) {
	adminID := middleware.AdminIdFromContext(a.ctx)
	if adminID <= 0 {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	rows, err := a.svcCtx.ServiceHub.GetAdminRbacMyMenus(a.ctx, adminID)
	if err != nil {
		return nil, err
	}

	type menuRow struct {
		ID        int64
		ParentID  int64
		MenuKey   string
		Label     string
		Icon      string
		Kind      int64
		Path      string
		SortOrder int64
		Placement string
	}

	all := make([]menuRow, 0, len(rows))
	for _, m := range rows {
		if m == nil {
			continue
		}
		all = append(all, menuRow{
			ID:        m.GetId(),
			ParentID:  m.GetParentId(),
			MenuKey:   m.GetMenuKey(),
			Label:     m.GetLabel(),
			Icon:      m.GetIcon(),
			Kind:      m.GetKind(),
			Path:      strings.TrimSpace(m.GetPath()),
			SortOrder: m.GetSortOrder(),
			Placement: m.GetPlacement(),
		})
	}

	avatarRows := make([]menuRow, 0)
	leftRows := make([]menuRow, 0, len(all))
	for _, r := range all {
		if menuPlacementStr(r.Placement) == "avatar" {
			if r.Kind == 1 && r.Path != "" {
				avatarRows = append(avatarRows, r)
			}
			continue
		}
		leftRows = append(leftRows, r)
	}

	sort.Slice(avatarRows, func(i, j int) bool {
		if avatarRows[i].SortOrder != avatarRows[j].SortOrder {
			return avatarRows[i].SortOrder < avatarRows[j].SortOrder
		}
		return avatarRows[i].ID < avatarRows[j].ID
	})
	avatarLinks := make([]AvatarMenuLink, 0, len(avatarRows))
	for _, r := range avatarRows {
		avatarLinks = append(avatarLinks, AvatarMenuLink{To: r.Path, Label: r.Label, Icon: r.Icon})
	}

	children := make(map[int64][]menuRow)
	roots := make([]menuRow, 0)
	for _, r := range leftRows {
		if r.ParentID == 0 {
			roots = append(roots, r)
		} else {
			children[r.ParentID] = append(children[r.ParentID], r)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		if roots[i].SortOrder != roots[j].SortOrder {
			return roots[i].SortOrder < roots[j].SortOrder
		}
		return roots[i].ID < roots[j].ID
	})

	out := make([]AdminMenuEntry, 0, len(roots))
	for _, r := range roots {
		if r.Kind == 1 { // leaf
			if r.Path == "" {
				continue
			}
			out = append(out, AdminMenuEntry{
				Kind:  "leaf",
				To:    r.Path,
				Label: r.Label,
				Icon:  r.Icon,
			})
			continue
		}
		if r.Kind != 2 {
			continue
		}
		k := r.MenuKey
		if k == "" {
			k = "group_" + r.Label
		}
		cs := children[r.ID]
		sort.Slice(cs, func(i, j int) bool {
			if cs[i].SortOrder != cs[j].SortOrder {
				return cs[i].SortOrder < cs[j].SortOrder
			}
			return cs[i].ID < cs[j].ID
		})
		links := make([]AdminMenuChildLink, 0, len(cs))
		for _, c := range cs {
			if c.Kind != 1 || strings.TrimSpace(c.Path) == "" {
				continue
			}
			links = append(links, AdminMenuChildLink{To: c.Path, Label: c.Label})
		}
		if len(links) == 0 {
			continue
		}
		out = append(out, AdminMenuEntry{
			Kind:     "group",
			Key:      k,
			Label:    r.Label,
			Icon:     r.Icon,
			Children: links,
		})
	}
	return &MyMenuResp{Sidebar: out, AvatarLinks: avatarLinks}, nil
}
