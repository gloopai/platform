package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdminRole struct {
	ID        int64
	Code      string
	Name      string
	Status    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AdminMenu struct {
	ID        int64
	ParentID  int64
	MenuKey   string
	Label     string
	Icon      string
	Kind      int64
	Path      string
	SortOrder int64
	Placement string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AdminRbacStore struct {
	db *gorm.DB
}

func NewAdminRbacStore(db *gorm.DB) *AdminRbacStore {
	return &AdminRbacStore{db: db}
}

func (s *AdminRbacStore) grantMenuToSuperAdmin(ctx context.Context, menuID int64) error {
	if menuID <= 0 {
		return nil
	}
	return s.db.WithContext(ctx).Exec(
		`INSERT INTO admin_role_menus (role_id, menu_id)
		 SELECT id, ? FROM admin_roles WHERE code = 'super_admin' AND status = 1
		 ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id)`,
		menuID,
	).Error
}

func (s *AdminRbacStore) ListMenus(ctx context.Context) ([]AdminMenu, error) {
	var out []AdminMenu
	if err := s.db.WithContext(ctx).
		Table("admin_menus").
		Select("id, parent_id, menu_key, label, icon, kind, path, sort_order, placement").
		Order("parent_id ASC, sort_order ASC, id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacStore) ListRoles(ctx context.Context) ([]AdminRole, error) {
	var out []AdminRole
	if err := s.db.WithContext(ctx).
		Table("admin_roles").
		Select("id, code, name, status").
		Order("id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacStore) CreateRole(ctx context.Context, code, name string, status int64) (*AdminRole, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, errors.New("code and name required")
	}
	r := &AdminRole{Code: code, Name: name, Status: status}
	if err := s.db.WithContext(ctx).Table("admin_roles").Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (s *AdminRbacStore) UpdateRole(ctx context.Context, id int64, name string, status int64) (*AdminRole, error) {
	name = strings.TrimSpace(name)
	if id <= 0 || name == "" {
		return nil, errors.New("id and name required")
	}
	if err := s.db.WithContext(ctx).
		Table("admin_roles").
		Where("id = ?", id).
		Updates(map[string]any{"name": name, "status": status}).Error; err != nil {
		return nil, err
	}
	var r AdminRole
	if err := s.db.WithContext(ctx).
		Table("admin_roles").
		Select("id, code, name, status").
		Where("id = ?", id).
		Limit(1).
		Take(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *AdminRbacStore) DeleteRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id required")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("admin_user_roles").Where("role_id = ?", id).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if err := tx.Table("admin_role_menus").Where("role_id = ?", id).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		return tx.Table("admin_roles").Where("id = ?", id).Delete(&struct{}{}).Error
	})
}

func (s *AdminRbacStore) GetRoleMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	if roleID <= 0 {
		return nil, errors.New("role_id required")
	}
	type row struct{ MenuID int64 }
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("admin_role_menus").
		Select("menu_id").
		Where("role_id = ?", roleID).
		Order("menu_id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.MenuID)
	}
	return out, nil
}

func (s *AdminRbacStore) SetRoleMenuIDs(ctx context.Context, roleID int64, menuIDs []int64) error {
	if roleID <= 0 {
		return errors.New("role_id required")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("admin_role_menus").Where("role_id = ?", roleID).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		now := time.Now()
		type bind struct {
			RoleID    int64     `gorm:"column:role_id"`
			MenuID    int64     `gorm:"column:menu_id"`
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		ins := make([]bind, 0, len(menuIDs))
		for _, mid := range menuIDs {
			if mid <= 0 {
				continue
			}
			ins = append(ins, bind{RoleID: roleID, MenuID: mid, CreatedAt: now})
		}
		if len(ins) == 0 {
			return nil
		}
		return tx.Table("admin_role_menus").Create(&ins).Error
	})
}

func (s *AdminRbacStore) GetUserRoleIDs(ctx context.Context, adminUserID int64) ([]int64, error) {
	if adminUserID <= 0 {
		return nil, errors.New("admin_user_id required")
	}
	type row struct{ RoleID int64 }
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("admin_user_roles").
		Select("role_id").
		Where("admin_user_id = ?", adminUserID).
		Order("role_id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.RoleID)
	}
	return out, nil
}

func (s *AdminRbacStore) SetUserRoleIDs(ctx context.Context, adminUserID int64, roleIDs []int64) error {
	if adminUserID <= 0 {
		return errors.New("admin_user_id required")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("admin_user_roles").Where("admin_user_id = ?", adminUserID).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		now := time.Now()
		type bind struct {
			AdminUserID int64     `gorm:"column:admin_user_id"`
			RoleID      int64     `gorm:"column:role_id"`
			CreatedAt   time.Time `gorm:"column:created_at"`
		}
		ins := make([]bind, 0, len(roleIDs))
		for _, rid := range roleIDs {
			if rid <= 0 {
				continue
			}
			ins = append(ins, bind{AdminUserID: adminUserID, RoleID: rid, CreatedAt: now})
		}
		if len(ins) == 0 {
			return nil
		}
		return tx.Table("admin_user_roles").Create(&ins).Error
	})
}

func (s *AdminRbacStore) IsUserSuperAdmin(ctx context.Context, adminUserID int64) (bool, error) {
	if adminUserID <= 0 {
		return false, errors.New("admin_user_id required")
	}
	var cnt int64
	err := s.db.WithContext(ctx).
		Table("admin_user_roles ur").
		Joins("JOIN admin_roles ar ON ar.id = ur.role_id").
		Where("ur.admin_user_id = ? AND ar.code = ? AND ar.status = 1", adminUserID, "super_admin").
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (s *AdminRbacStore) ListMenusByUser(ctx context.Context, adminUserID int64) ([]AdminMenu, error) {
	roleIDs, err := s.GetUserRoleIDs(ctx, adminUserID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []AdminMenu{}, nil
	}

	var out []AdminMenu
	if err := s.db.WithContext(ctx).
		Table("admin_role_menus rm").
		Joins("JOIN admin_menus m ON m.id = rm.menu_id").
		Where("rm.role_id IN ?", roleIDs).
		Select("DISTINCT m.id, m.parent_id, m.menu_key, m.label, m.icon, m.kind, m.path, m.sort_order, m.placement").
		Order("m.parent_id ASC, m.sort_order ASC, m.id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacStore) ListPermKeysByUser(ctx context.Context, adminUserID int64) (isSuper bool, permKeys []string, err error) {
	roleIDs, err := s.GetUserRoleIDs(ctx, adminUserID)
	if err != nil {
		return false, nil, err
	}
	if len(roleIDs) == 0 {
		return false, []string{}, nil
	}

	type row struct{ PermKey string }
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("admin_role_permissions rp").
		Joins("JOIN admin_permissions p ON p.id = rp.perm_id").
		Where("rp.role_id IN ?", roleIDs).
		Where("p.status = 1").
		Select("DISTINCT p.perm_key as perm_key").
		Order("p.perm_key ASC").
		Find(&rows).Error; err != nil {
		return false, nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if strings.TrimSpace(r.PermKey) == "" {
			continue
		}
		out = append(out, r.PermKey)
	}
	return false, out, nil
}

func normalizeMenuPlacement(kind int64, placement string) (string, error) {
	if kind == 2 {
		return "left", nil
	}
	p := strings.TrimSpace(strings.ToLower(placement))
	if p == "" {
		p = "left"
	}
	if p != "left" && p != "avatar" {
		return "", errors.New("placement must be left or avatar")
	}
	return p, nil
}

func (s *AdminRbacStore) menuChildCount(ctx context.Context, id int64) (int64, error) {
	var cnt int64
	if err := s.db.WithContext(ctx).Table("admin_menus").Where("parent_id = ?", id).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (s *AdminRbacStore) CreateMenu(ctx context.Context, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error) {
	menuKey = strings.TrimSpace(menuKey)
	label = strings.TrimSpace(label)
	icon = strings.TrimSpace(icon)
	path = strings.TrimSpace(path)
	if menuKey == "" || label == "" {
		return nil, errors.New("menu_key and label required")
	}
	if kind != 1 && kind != 2 {
		return nil, errors.New("kind must be 1 (page) or 2 (group)")
	}
	if parentID > 0 {
		var cnt int64
		if err := s.db.WithContext(ctx).Table("admin_menus").Where("id = ?", parentID).Count(&cnt).Error; err != nil {
			return nil, err
		}
		if cnt == 0 {
			return nil, errors.New("parent menu not found")
		}
	}
	var exists int64
	if err := s.db.WithContext(ctx).Table("admin_menus").Where("menu_key = ?", menuKey).Count(&exists).Error; err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, errors.New("menu_key already exists")
	}
	pl, err := normalizeMenuPlacement(kind, placement)
	if err != nil {
		return nil, err
	}
	if pl == "avatar" {
		if parentID != 0 {
			return nil, errors.New("avatar menu items must be top-level (parent_id=0)")
		}
		if kind != 1 {
			return nil, errors.New("avatar placement only supports page items (kind=1)")
		}
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("avatar menu requires path")
		}
	}
	m := AdminMenu{
		ParentID:  parentID,
		MenuKey:   menuKey,
		Label:     label,
		Icon:      icon,
		Kind:      kind,
		Path:      path,
		SortOrder: sortOrder,
		Placement: pl,
	}
	if err := s.db.WithContext(ctx).Table("admin_menus").Create(&m).Error; err != nil {
		return nil, err
	}
	// Keep menu creation intuitive for super_admin operators.
	if err := s.grantMenuToSuperAdmin(ctx, m.ID); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *AdminRbacStore) UpdateMenu(ctx context.Context, id int64, parentID int64, menuKey, label, icon string, kind int64, path string, sortOrder int64, placement string) (*AdminMenu, error) {
	if id <= 0 {
		return nil, errors.New("id required")
	}
	menuKey = strings.TrimSpace(menuKey)
	label = strings.TrimSpace(label)
	icon = strings.TrimSpace(icon)
	path = strings.TrimSpace(path)
	if menuKey == "" || label == "" {
		return nil, errors.New("menu_key and label required")
	}
	if kind != 1 && kind != 2 {
		return nil, errors.New("kind must be 1 (page) or 2 (group)")
	}
	if parentID == id {
		return nil, errors.New("invalid parent")
	}
	if parentID > 0 {
		var cnt int64
		if err := s.db.WithContext(ctx).Table("admin_menus").Where("id = ?", parentID).Count(&cnt).Error; err != nil {
			return nil, err
		}
		if cnt == 0 {
			return nil, errors.New("parent menu not found")
		}
	}
	var dup int64
	if err := s.db.WithContext(ctx).Table("admin_menus").Where("menu_key = ? AND id <> ?", menuKey, id).Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, errors.New("menu_key already exists")
	}
	pl, err := normalizeMenuPlacement(kind, placement)
	if err != nil {
		return nil, err
	}
	if pl == "avatar" {
		if parentID != 0 {
			return nil, errors.New("avatar menu items must be top-level (parent_id=0)")
		}
		if kind != 1 {
			return nil, errors.New("avatar placement only supports page items (kind=1)")
		}
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("avatar menu requires path")
		}
		n, cerr := s.menuChildCount(ctx, id)
		if cerr != nil {
			return nil, cerr
		}
		if n > 0 {
			return nil, errors.New("cannot set avatar placement while menu has children")
		}
	}
	if err := s.db.WithContext(ctx).Table("admin_menus").Where("id = ?", id).Updates(map[string]any{
		"parent_id":  parentID,
		"menu_key":   menuKey,
		"label":      label,
		"icon":       icon,
		"kind":       kind,
		"path":       path,
		"sort_order": sortOrder,
		"placement":  pl,
	}).Error; err != nil {
		return nil, err
	}
	var out AdminMenu
	if err := s.db.WithContext(ctx).
		Table("admin_menus").
		Select("id, parent_id, menu_key, label, icon, kind, path, sort_order, placement").
		Where("id = ?", id).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AdminRbacStore) DeleteMenu(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id required")
	}
	var children int64
	if err := s.db.WithContext(ctx).Table("admin_menus").Where("parent_id = ?", id).Count(&children).Error; err != nil {
		return err
	}
	if children > 0 {
		return errors.New("请先删除或移动子菜单")
	}
	return s.db.WithContext(ctx).Table("admin_menus").Where("id = ?", id).Delete(&struct{}{}).Error
}
