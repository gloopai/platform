package store

import (
	"context"
	"database/sql"
	"time"
)

type AdminSession struct {
	AdminId   int64
	TokenHash string
	ExpiresAt time.Time
}

type MerchantSession struct {
	MerchantId string
	TokenHash  string
	ExpiresAt  time.Time
}

type SessionsStore struct {
	db *sql.DB
}

func NewSessionsStore(db *sql.DB) *SessionsStore {
	return &SessionsStore{db: db}
}

func (s *SessionsStore) CreateAdminSession(ctx context.Context, adminId int64, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO admin_sessions (admin_id, token_hash, expires_at, created_at)
VALUES (?, ?, ?, NOW())
`, adminId, tokenHash, expiresAt)
	return err
}

func (s *SessionsStore) GetAdminSession(ctx context.Context, tokenHash string) (*AdminSession, error) {
	var r AdminSession
	if err := s.db.QueryRowContext(ctx, `
SELECT admin_id, token_hash, expires_at
FROM admin_sessions
WHERE token_hash = ? AND expires_at > NOW()
LIMIT 1
`, tokenHash).Scan(&r.AdminId, &r.TokenHash, &r.ExpiresAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SessionsStore) DeleteAdminSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

func (s *SessionsStore) CreateMerchantSession(ctx context.Context, merchantId string, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO merchant_sessions (merchant_id, token_hash, expires_at, created_at)
VALUES (?, ?, ?, NOW())
`, merchantId, tokenHash, expiresAt)
	return err
}

func (s *SessionsStore) GetMerchantSession(ctx context.Context, tokenHash string) (*MerchantSession, error) {
	var r MerchantSession
	if err := s.db.QueryRowContext(ctx, `
SELECT merchant_id, token_hash, expires_at
FROM merchant_sessions
WHERE token_hash = ? AND expires_at > NOW()
LIMIT 1
`, tokenHash).Scan(&r.MerchantId, &r.TokenHash, &r.ExpiresAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SessionsStore) DeleteMerchantSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM merchant_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

