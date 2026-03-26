package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdminPermission struct {
	ID       int64
	PermKey  string
	Label    string
	Category string
	MenuKey  string
	Status   int64
}

type AdminApiRule struct {
	ID          int64
	Method      string
	PathPattern string
	PermKey     string
	Status      int64
	Remark      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AdminRbacConfigStore struct {
	db *gorm.DB
}

func NewAdminRbacConfigStore(db *gorm.DB) *AdminRbacConfigStore {
	return &AdminRbacConfigStore{db: db}
}

func (s *AdminRbacConfigStore) ListPermissions(ctx context.Context) ([]AdminPermission, error) {
	var out []AdminPermission
	if err := s.db.WithContext(ctx).
		Table("admin_permissions").
		Select("id, perm_key, label, category, menu_key, status").
		Order("menu_key ASC, category ASC, perm_key ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacConfigStore) CreatePermission(ctx context.Context, permKey, label, category, menuKey string, status int64) (*AdminPermission, error) {
	permKey = strings.TrimSpace(permKey)
	label = strings.TrimSpace(label)
	category = strings.TrimSpace(category)
	menuKey = strings.TrimSpace(menuKey)
	if permKey == "" || label == "" {
		return nil, errors.New("perm_key and label required")
	}
	if status == 0 {
		status = 1
	}
	p := &AdminPermission{PermKey: permKey, Label: label, Category: category, MenuKey: menuKey, Status: status}
	if err := s.db.WithContext(ctx).Table("admin_permissions").Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdminRbacConfigStore) UpdatePermission(ctx context.Context, id int64, label, category, menuKey string, status int64) (*AdminPermission, error) {
	label = strings.TrimSpace(label)
	category = strings.TrimSpace(category)
	menuKey = strings.TrimSpace(menuKey)
	if id <= 0 || label == "" {
		return nil, errors.New("id and label required")
	}
	if err := s.db.WithContext(ctx).
		Table("admin_permissions").
		Where("id = ?", id).
		Updates(map[string]any{"label": label, "category": category, "menu_key": menuKey, "status": status}).Error; err != nil {
		return nil, err
	}
	var out AdminPermission
	if err := s.db.WithContext(ctx).
		Table("admin_permissions").
		Select("id, perm_key, label, category, menu_key, status").
		Where("id = ?", id).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AdminRbacConfigStore) DeletePermission(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id required")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// remove role bindings first
		if err := tx.Table("admin_role_permissions").Where("perm_id = ?", id).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		return tx.Table("admin_permissions").Where("id = ?", id).Delete(&struct{}{}).Error
	})
}

func (s *AdminRbacConfigStore) GetRolePermKeys(ctx context.Context, roleID int64) ([]string, error) {
	if roleID <= 0 {
		return nil, errors.New("role_id required")
	}
	type row struct{ PermKey string }
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("admin_role_permissions rp").
		Joins("JOIN admin_permissions p ON p.id = rp.perm_id").
		Where("rp.role_id = ?", roleID).
		Select("p.perm_key as perm_key").
		Order("p.perm_key ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		k := strings.TrimSpace(r.PermKey)
		if k != "" {
			out = append(out, k)
		}
	}
	return out, nil
}

func (s *AdminRbacConfigStore) SetRolePermKeys(ctx context.Context, roleID int64, permKeys []string) error {
	if roleID <= 0 {
		return errors.New("role_id required")
	}
	keys := make([]string, 0, len(permKeys))
	seen := map[string]struct{}{}
	for _, k := range permKeys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("admin_role_permissions").Where("role_id = ?", roleID).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		// resolve perm ids
		type permRow struct{ ID int64 }
		var perms []permRow
		if err := tx.Table("admin_permissions").Select("id").Where("perm_key IN ?", keys).Find(&perms).Error; err != nil {
			return err
		}
		if len(perms) == 0 {
			return nil
		}
		now := time.Now()
		type bind struct {
			RoleID    int64     `gorm:"column:role_id"`
			PermID    int64     `gorm:"column:perm_id"`
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		ins := make([]bind, 0, len(perms))
		for _, p := range perms {
			if p.ID <= 0 {
				continue
			}
			ins = append(ins, bind{RoleID: roleID, PermID: p.ID, CreatedAt: now})
		}
		if len(ins) == 0 {
			return nil
		}
		return tx.Table("admin_role_permissions").Create(&ins).Error
	})
}

func (s *AdminRbacConfigStore) ListApiRules(ctx context.Context) ([]AdminApiRule, error) {
	var out []AdminApiRule
	if err := s.db.WithContext(ctx).
		Table("admin_api_rules").
		Select("id, method, path_pattern, perm_key, status, remark, created_at, updated_at").
		Order("method ASC, path_pattern ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacConfigStore) ListEnabledApiRules(ctx context.Context) ([]AdminApiRule, error) {
	var out []AdminApiRule
	if err := s.db.WithContext(ctx).
		Table("admin_api_rules").
		Select("id, method, path_pattern, perm_key, status, remark").
		Where("status = 1").
		Order("method ASC, path_pattern ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminRbacConfigStore) UpsertApiRule(ctx context.Context, method, pathPattern, permKey string, status int64, remark string) (*AdminApiRule, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	pathPattern = strings.TrimSpace(pathPattern)
	permKey = strings.TrimSpace(permKey)
	remark = strings.TrimSpace(remark)
	if method == "" || pathPattern == "" || permKey == "" {
		return nil, errors.New("method, path_pattern, perm_key required")
	}
	if status == 0 {
		status = 1
	}
	// raw upsert
	if err := s.db.WithContext(ctx).Exec(
		`INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE perm_key = VALUES(perm_key), status = VALUES(status), remark = VALUES(remark)`,
		method, pathPattern, permKey, status, remark,
	).Error; err != nil {
		return nil, err
	}
	var out AdminApiRule
	if err := s.db.WithContext(ctx).
		Table("admin_api_rules").
		Select("id, method, path_pattern, perm_key, status, remark").
		Where("method = ? AND path_pattern = ?", method, pathPattern).
		Limit(1).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AdminRbacConfigStore) DeleteApiRule(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id required")
	}
	return s.db.WithContext(ctx).Table("admin_api_rules").Where("id = ?", id).Delete(&struct{}{}).Error
}
