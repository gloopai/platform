package svc

import (
	"strings"
	"time"

	"github.com/gloopai/platform/common/dbdsn"
	"github.com/gloopai/platform/common/gormx"
	ntf "github.com/gloopai/platform/common/notify"
	"github.com/gloopai/platform/service-hub/internal/config"
	hubnotify "github.com/gloopai/platform/service-hub/internal/notify"
	"github.com/gloopai/platform/service-hub/internal/store"
	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config              config.Config
	AdminUsers          *store.AdminUsersStore
	AdminRbac           *store.AdminRbacStore
	AdminRbacCfg        *store.AdminRbacConfigStore
	AdminOpLogs         *store.AdminOperationLogsStore
	ScheduledJobs       *store.ScheduledJobsStore
	GlobalSettings      *store.GlobalSettingsStore
	PortalNotifications *store.PortalNotificationsStore
	NotifyPublisher     *hubnotify.Publisher
}

func NewServiceContext(c config.Config) *ServiceContext {
	gdb := gormx.MustOpenMySQL(dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone))
	sqlDB, err := gdb.DB()
	if err != nil {
		panic(err)
	}
	if v := c.Mysql.MaxOpenConns; v > 0 {
		sqlDB.SetMaxOpenConns(v)
	}
	if v := c.Mysql.MaxIdleConns; v > 0 {
		sqlDB.SetMaxIdleConns(v)
	}
	if sec := c.Mysql.ConnMaxLifetimeSeconds; sec > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(sec) * time.Second)
	}
	return newServiceContextForDB(c, gdb)
}

// NewServiceContextWithRuntime wires stores and optional NSQ on an existing DB (e.g. embedded in pay core).
func NewServiceContextWithRuntime(gdb *gorm.DB, nsqdAddr, portalTopic string) *ServiceContext {
	var c config.Config
	c.Nsq.NsqdTCPAddr = strings.TrimSpace(nsqdAddr)
	c.Nsq.PortalNotifyTopic = strings.TrimSpace(portalTopic)
	return newServiceContextForDB(c, gdb)
}

func newServiceContextForDB(c config.Config, gdb *gorm.DB) *ServiceContext {
	portalNotifStore := store.NewPortalNotificationsStore(gdb)

	var notifyPublisher *hubnotify.Publisher
	if addr := strings.TrimSpace(c.Nsq.NsqdTCPAddr); addr != "" {
		producer, err := nsq.NewProducer(addr, nsq.NewConfig())
		if err != nil {
			panic(err)
		}
		if err := producer.Ping(); err != nil {
			panic(err)
		}
		topic := strings.TrimSpace(c.Nsq.PortalNotifyTopic)
		if topic == "" {
			topic = ntf.PortalNotifyTopic
		}
		notifyPublisher = hubnotify.NewPublisher(producer, topic, portalNotifStore)
	}

	return &ServiceContext{
		Config:              c,
		AdminUsers:          store.NewAdminUsersStore(gdb),
		AdminRbac:           store.NewAdminRbacStore(gdb),
		AdminRbacCfg:        store.NewAdminRbacConfigStore(gdb),
		AdminOpLogs:         store.NewAdminOperationLogsStore(gdb),
		ScheduledJobs:       store.NewScheduledJobsStore(gdb),
		GlobalSettings:      store.NewGlobalSettingsStore(gdb),
		PortalNotifications: portalNotifStore,
		NotifyPublisher:     notifyPublisher,
	}
}
