package svc

import (
	"context"
	"strings"

	"github.com/gloopai/platform/common/gatewaymw"
	"github.com/gloopai/platform/service-hub/hubclient"
)

// oplogServiceHub adapts service-hub RPC to [gatewaymw.OpLogHub] for admin operation logs.
type oplogServiceHub struct {
	sh hubclient.ServiceHub
}

func (h oplogServiceHub) ListAdminApiRules(ctx context.Context, page, pageSize int64, q, permKey string) ([]*hubclient.AdminApiRule, int64, error) {
	return h.sh.ListAdminApiRules(ctx, page, pageSize, q, permKey)
}

func (h oplogServiceHub) FetchAdminUsername(ctx context.Context, adminUserID int64) string {
	u, err := h.sh.GetAdminUserById(ctx, adminUserID)
	if err != nil || u == nil {
		return ""
	}
	return strings.TrimSpace(u.GetUsername())
}

func (h oplogServiceHub) RecordOpLog(ctx context.Context, rec gatewaymw.OpLogRecord) error {
	return h.sh.RecordAdminOperationLog(ctx, &hubclient.RecordAdminOperationLogReq{
		RequestId:     rec.RequestID,
		AdminUserId:   rec.AdminUserID,
		AdminUsername: rec.AdminUsername,
		OperatorIp:    rec.OperatorIP,
		UserAgent:     rec.UserAgent,
		Method:        rec.Method,
		Path:          rec.Path,
		QueryString:   rec.QueryString,
		PermKey:       rec.PermKey,
		HttpStatus:    rec.HTTPStatus,
		Success:       rec.Success,
		DurationMs:    rec.DurationMs,
		ErrorMessage:  rec.ErrorMessage,
	})
}
