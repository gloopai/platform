// 管理台：通知推送测试（调用 service-hub PublishPortalNotification）
package handler

import (
	"net/http"

	"github.com/gloopai/pay/common/pb/servicehub"
	"github.com/gloopai/pay/gateway/internal/apiresp"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
)

// AdminNotificationTestHandler POST /v1/admin/notifications/test — 向当前登录管理员发一条测试通知（master 令牌则全员广播）。
func AdminNotificationTestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			apiresp.Fail(w, apiresp.CodeInvalidParams, "method not allowed")
			return
		}
		uid := middleware.AdminIdFromContext(r.Context())
		req := &servicehub.PublishPortalNotificationReq{
			Portal:        servicehub.NotificationPortal_NOTIFICATION_PORTAL_ADMIN,
			Title:         "测试推送",
			Body:          "这是一条来自管理台的测试通知。",
			Severity:      "info",
			LinkPath:      "/ops",
			LinkQueryJson: "{}",
			MetaJson:      "{}",
		}
		if uid > 0 {
			req.Broadcast = false
			req.AdminUserIds = []int64{uid}
		} else {
			req.Broadcast = true
		}
		resp, err := svcCtx.ServiceHub.PublishPortalNotification(r.Context(), req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, map[string]string{"notification_id": resp.GetNotificationId()})
	}
}
