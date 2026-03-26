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
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AdminRbacStore struct {
	db *gorm.DB
}

func NewAdminRbacStore(db *gorm.DB) *AdminRbacStore {
	return &AdminRbacStore{db: db}
}

func (s *AdminRbacStore) ListMenus(ctx context.Context) ([]AdminMenu, error) {
	var out []AdminMenu
	if err := s.db.WithContext(ctx).
		Table("admin_menus").
		Select("id, parent_id, menu_key, label, icon, kind, path, sort_order").
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
	isSuper, err := s.IsUserSuperAdmin(ctx, adminUserID)
	if err != nil {
		return nil, err
	}
	if isSuper {
		return s.ListMenus(ctx)
	}

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
		Select("DISTINCT m.id, m.parent_id, m.menu_key, m.label, m.icon, m.kind, m.path, m.sort_order").
		Order("m.parent_id ASC, m.sort_order ASC, m.id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacStore) ListPermKeysByUser(ctx context.Context, adminUserID int64) (isSuper bool, permKeys []string, err error) {
	isSuper, err = s.IsUserSuperAdmin(ctx, adminUserID)
	if err != nil {
		return false, nil, err
	}
	if isSuper {
		return true, []string{"*"}, nil
	}

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
