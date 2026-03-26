package store

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Status       int64
	MfaSecret    string
	MfaEnabled   int64
}

type AdminUsersStore struct {
	db *gorm.DB
}

func NewAdminUsersStore(db *gorm.DB) *AdminUsersStore {
	return &AdminUsersStore{db: db}
}

func (s *AdminUsersStore) FindByUsername(ctx context.Context, username string) (*AdminUser, error) {
	var u AdminUser
	tx := s.db.WithContext(ctx).
		Table("admin_users").
		Select("id, username, password_hash, status, mfa_secret, mfa_enabled").
		Where("username = ?", username).
		Limit(1).
		Take(&u)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &u, nil
}

type AdminUserPublic struct {
	ID         int64
	Username   string
	Status     int64
	MfaEnabled int64
}

func (s *AdminUsersStore) List(ctx context.Context) ([]AdminUserPublic, error) {
	var out []AdminUserPublic
	if err := s.db.WithContext(ctx).
		Table("admin_users").
		Select("id, username, status, mfa_enabled").
		Order("id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminUsersStore) GetByID(ctx context.Context, id int64) (*AdminUser, error) {
	var u AdminUser
	tx := s.db.WithContext(ctx).
		Table("admin_users").
		Select("id, username, password_hash, status, mfa_secret, mfa_enabled").
		Where("id = ?", id).
		Limit(1).
		Take(&u)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &u, nil
}

func (s *AdminUsersStore) Create(ctx context.Context, username, passwordHash string, status int64) (*AdminUserPublic, error) {
	if status == 0 {
		status = 1
	}
	if err := s.db.WithContext(ctx).
		Table("admin_users").
		Create(map[string]any{
			"username":      username,
			"password_hash": passwordHash,
			"status":        status,
			"mfa_secret":    "",
			"mfa_enabled":   0,
		}).Error; err != nil {
		return nil, err
	}
	row, err := s.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &AdminUserPublic{
		ID: row.ID, Username: row.Username, Status: row.Status, MfaEnabled: row.MfaEnabled,
	}, nil
}

func (s *AdminUsersStore) Update(ctx context.Context, id int64, status int64, passwordHash, mfaSecret *string, mfaEnabled *int64) (*AdminUserPublic, error) {
	updates := map[string]any{
		"status": status,
	}
	if passwordHash != nil {
		updates["password_hash"] = *passwordHash
	}
	if mfaSecret != nil {
		updates["mfa_secret"] = *mfaSecret
	}
	if mfaEnabled != nil {
		updates["mfa_enabled"] = *mfaEnabled
	}
	tx := s.db.WithContext(ctx).Table("admin_users").Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	row, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &AdminUserPublic{
		ID: row.ID, Username: row.Username, Status: row.Status, MfaEnabled: row.MfaEnabled,
	}, nil
}

func (s *AdminUsersStore) Delete(ctx context.Context, id int64) error {
	tx := s.db.WithContext(ctx).Table("admin_users").Where("id = ?", id).Delete(nil)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
