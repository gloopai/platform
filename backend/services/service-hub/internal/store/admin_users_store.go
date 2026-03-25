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
		Select("id, username, password_hash, status").
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
	ID       int64
	Username string
	Status   int64
}

func (s *AdminUsersStore) List(ctx context.Context) ([]AdminUserPublic, error) {
	var out []AdminUserPublic
	if err := s.db.WithContext(ctx).
		Table("admin_users").
		Select("id, username, status").
		Order("id ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

