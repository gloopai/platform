package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdminOperationLogRow struct {
	ID            int64
	CreatedAt     time.Time
	RequestID     string
	AdminUserID   int64
	AdminUsername string
	OperatorIP    string
	UserAgent     string
	Method        string
	Path          string
	QueryString   string
	PermKey       string
	HTTPStatus    int32
	Success       bool
	DurationMs    int64
	ErrorMessage  string
}

type AdminOperationLogsStore struct {
	db *gorm.DB
}

func NewAdminOperationLogsStore(db *gorm.DB) *AdminOperationLogsStore {
	return &AdminOperationLogsStore{db: db}
}

func (s *AdminOperationLogsStore) Insert(ctx context.Context, row *AdminOperationLogRow) error {
	if row == nil {
		return nil
	}
	return s.db.WithContext(ctx).Exec(`
INSERT INTO admin_operation_logs (
  request_id, admin_user_id, admin_username, operator_ip, user_agent, method, path, query_string,
  perm_key, http_status, success, duration_ms, error_message
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(row.RequestID),
		row.AdminUserID,
		strings.TrimSpace(row.AdminUsername),
		strings.TrimSpace(row.OperatorIP),
		strings.TrimSpace(row.UserAgent),
		strings.ToUpper(strings.TrimSpace(row.Method)),
		strings.TrimSpace(row.Path),
		strings.TrimSpace(row.QueryString),
		strings.TrimSpace(row.PermKey),
		row.HTTPStatus,
		boolToTiny(row.Success),
		row.DurationMs,
		strings.TrimSpace(row.ErrorMessage),
	).Error
}

func (s *AdminOperationLogsStore) List(
	ctx context.Context,
	startSec, endSec, adminUserID int64,
	method, pathKeyword, permKey string,
	success *bool,
	limit, offset int64,
) ([]AdminOperationLogRow, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var where []string
	var args []any
	if startSec > 0 {
		where = append(where, "created_at >= ?")
		args = append(args, time.Unix(startSec, 0))
	}
	if endSec > 0 {
		where = append(where, "created_at < ?")
		args = append(args, time.Unix(endSec, 0))
	}
	if adminUserID > 0 {
		where = append(where, "admin_user_id = ?")
		args = append(args, adminUserID)
	}
	if v := strings.ToUpper(strings.TrimSpace(method)); v != "" {
		where = append(where, "method = ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(pathKeyword); v != "" {
		where = append(where, "path LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if v := strings.TrimSpace(permKey); v != "" {
		where = append(where, "perm_key = ?")
		args = append(args, v)
	}
	if success != nil {
		if *success {
			where = append(where, "success = 1")
		} else {
			where = append(where, "success = 0")
		}
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM admin_operation_logs %s", whereClause)
	var total int64
	if err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	selArgs := append(append([]any{}, args...), limit, offset)
	selSQL := fmt.Sprintf(`
SELECT id, created_at, request_id, admin_user_id, admin_username, operator_ip, user_agent, method, path,
       query_string, perm_key, http_status, success, duration_ms, error_message
  FROM admin_operation_logs
  %s
 ORDER BY id DESC
 LIMIT ? OFFSET ?`, whereClause)
	rows, err := s.db.WithContext(ctx).Raw(selSQL, selArgs...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]AdminOperationLogRow, 0, limit)
	for rows.Next() {
		var r AdminOperationLogRow
		var ok int8
		if err := rows.Scan(
			&r.ID, &r.CreatedAt, &r.RequestID, &r.AdminUserID, &r.AdminUsername, &r.OperatorIP, &r.UserAgent, &r.Method, &r.Path,
			&r.QueryString, &r.PermKey, &r.HTTPStatus, &ok, &r.DurationMs, &r.ErrorMessage,
		); err != nil {
			return nil, 0, err
		}
		r.Success = ok != 0
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func boolToTiny(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
