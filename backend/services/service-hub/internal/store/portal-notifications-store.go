package store

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// PortalNotificationRow is a persisted portal notification (real-time still goes via NSQ; offline users miss push but row remains for history/audit).
type PortalNotificationRow struct {
	ID                    string    `gorm:"column:id;primaryKey"`
	Portal                string    `gorm:"column:portal"`
	Broadcast             int       `gorm:"column:broadcast"` // 0/1
	Title                 string    `gorm:"column:title"`
	Body                  string    `gorm:"column:body"`
	Severity              string    `gorm:"column:severity"`
	LinkPath              string    `gorm:"column:link_path"`
	LinkQueryJSON         string    `gorm:"column:link_query_json"`
	MetaJSON              string    `gorm:"column:meta_json"`
	TargetAdminIDsJSON    string    `gorm:"column:target_admin_ids"`    // JSON array or "[]"
	TargetMerchantIDsJSON string    `gorm:"column:target_merchant_ids"` // JSON array or "[]"
	CreatedAt             time.Time `gorm:"column:created_at"`
}

func (PortalNotificationRow) TableName() string { return "portal_notifications" }

type PortalNotificationsStore struct {
	db *gorm.DB
}

func NewPortalNotificationsStore(db *gorm.DB) *PortalNotificationsStore {
	return &PortalNotificationsStore{db: db}
}

func (s *PortalNotificationsStore) Insert(ctx context.Context, row *PortalNotificationRow) error {
	return s.db.WithContext(ctx).Table("portal_notifications").Create(row).Error
}

func (s *PortalNotificationsStore) DeleteByID(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Table("portal_notifications").Where("id = ?", id).Delete(&PortalNotificationRow{}).Error
}
