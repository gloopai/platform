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

// AdminUserPublic 列表展示用，不含密码。
type AdminUserPublic struct {
	ID       int64
	Username string
	Status   int64
}

// List 管理台账号列表（只读）。
func (s *AdminUsersStore) List(ctx context.Context) ([]AdminUserPublic, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, username, status
FROM admin_users
ORDER BY id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AdminUserPublic
	for rows.Next() {
		var r AdminUserPublic
		if err := rows.Scan(&r.ID, &r.Username, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
