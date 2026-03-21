package store

import (
	"context"
	"database/sql"
)

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Status       int64
}

type AdminUsersStore struct {
	db *sql.DB
}

func NewAdminUsersStore(db *sql.DB) *AdminUsersStore {
	return &AdminUsersStore{db: db}
}

func (s *AdminUsersStore) FindByUsername(ctx context.Context, username string) (*AdminUser, error) {
	var u AdminUser
	if err := s.db.QueryRowContext(ctx, `
SELECT id, username, password_hash, status
FROM admin_users
WHERE username = ?
LIMIT 1
`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Status); err != nil {
		return nil, err
	}
	return &u, nil
}
