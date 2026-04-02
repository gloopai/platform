package notify

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gloopai/platform/common/pb/servicehub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PersistRow 为 portal_notifications 表字段（与具体 ORM 无关，由服务映射）。
type PersistRow struct {
	ID                    string
	Portal                string
	Broadcast             int // 0/1
	Title                 string
	Body                  string
	Severity              string
	LinkPath              string
	LinkQueryJSON         string
	MetaJSON              string
	TargetAdminIDsJSON    string
	TargetMerchantIDsJSON string
	CreatedAt             time.Time
}

// Result 为 Build 输出：NSQ 字节 + 可选入库行。
type Result struct {
	ID      string
	NSQBody []byte
	Row     *PersistRow
}

// Build 校验 proto，生成 Envelope、NSQ 线包与 PersistRow；无 I/O。
func Build(req *servicehub.PublishPortalNotificationReq) (*Result, error) {
	if req == nil {
		return nil, errStatus(codes.InvalidArgument, "request required")
	}
	portal := req.GetPortal()
	title := strings.TrimSpace(req.GetTitle())
	if portal == servicehub.NotificationPortal_NOTIFICATION_PORTAL_UNSPECIFIED {
		return nil, errStatus(codes.InvalidArgument, "portal required")
	}
	if title == "" {
		return nil, errStatus(codes.InvalidArgument, "title required")
	}
	broadcast := req.GetBroadcast()
	if !broadcast {
		switch portal {
		case servicehub.NotificationPortal_NOTIFICATION_PORTAL_ADMIN:
			if len(req.GetAdminUserIds()) == 0 {
				return nil, errStatus(codes.InvalidArgument, "admin_user_ids required when not broadcast")
			}
		case servicehub.NotificationPortal_NOTIFICATION_PORTAL_MERCHANT:
			if len(req.GetMerchantIds()) == 0 {
				return nil, errStatus(codes.InvalidArgument, "merchant_ids required when not broadcast")
			}
		default:
			return nil, errStatus(codes.InvalidArgument, "unknown portal")
		}
	}

	id := uuid.NewString()
	sev := strings.TrimSpace(req.GetSeverity())
	if sev == "" {
		sev = "info"
	}
	linkQuery := strings.TrimSpace(req.GetLinkQueryJson())
	if linkQuery == "" {
		linkQuery = "{}"
	}
	meta := strings.TrimSpace(req.GetMetaJson())
	if meta == "" {
		meta = "{}"
	}

	portalStr := portalString(portal)
	env := Envelope{
		Event:         "notification",
		Portal:        portalStr,
		ID:            id,
		Title:         title,
		Body:          strings.TrimSpace(req.GetBody()),
		Severity:      sev,
		LinkPath:      strings.TrimSpace(req.GetLinkPath()),
		LinkQueryJSON: linkQuery,
		MetaJSON:      meta,
	}
	envBytes, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal envelope: %w", err)
	}

	wire := PortalNSQMessage{
		Portal:       portalStr,
		Broadcast:    broadcast,
		AdminUserIDs: req.GetAdminUserIds(),
		MerchantIDs:  req.GetMerchantIds(),
		Envelope:     envBytes,
	}
	body, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("marshal nsq wire: %w", err)
	}

	bcast := 0
	if broadcast {
		bcast = 1
	}
	adminTargets, err := json.Marshal(req.GetAdminUserIds())
	if err != nil {
		return nil, fmt.Errorf("marshal admin targets: %w", err)
	}
	merTargets, err := json.Marshal(req.GetMerchantIds())
	if err != nil {
		return nil, fmt.Errorf("marshal merchant targets: %w", err)
	}

	row := &PersistRow{
		ID:                    id,
		Portal:                portalStr,
		Broadcast:             bcast,
		Title:                 title,
		Body:                  strings.TrimSpace(req.GetBody()),
		Severity:              sev,
		LinkPath:              strings.TrimSpace(req.GetLinkPath()),
		LinkQueryJSON:         linkQuery,
		MetaJSON:              meta,
		TargetAdminIDsJSON:    string(adminTargets),
		TargetMerchantIDsJSON: string(merTargets),
		CreatedAt:             time.Now(),
	}

	return &Result{ID: id, NSQBody: body, Row: row}, nil
}

func portalString(p servicehub.NotificationPortal) string {
	switch p {
	case servicehub.NotificationPortal_NOTIFICATION_PORTAL_ADMIN:
		return "admin"
	case servicehub.NotificationPortal_NOTIFICATION_PORTAL_MERCHANT:
		return "merchant"
	default:
		return "unspecified"
	}
}

func errStatus(c codes.Code, msg string) error {
	return status.Error(c, msg)
}

// GRPCStatusOrInternal 保留 gRPC status；否则转为 Internal。
func GRPCStatusOrInternal(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Error(codes.Internal, err.Error())
}
