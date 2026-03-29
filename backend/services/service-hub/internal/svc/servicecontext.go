package svc

import (
	"strings"
	"time"

	"github.com/gloopai/pay/common/dbdsn"
	ntf "github.com/gloopai/pay/common/notify"
	"github.com/gloopai/pay/service-hub/internal/config"
	hubnotify "github.com/gloopai/pay/service-hub/internal/notify"
	"github.com/gloopai/pay/service-hub/internal/store"
	"github.com/nsqio/go-nsq"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config               config.Config
	AdminUsers           *store.AdminUsersStore
	AdminRbac            *store.AdminRbacStore
	AdminRbacCfg         *store.AdminRbacConfigStore
	GlobalSettings      *store.GlobalSettingsStore
	PortalNotifications *store.PortalNotificationsStore
	NotifyPublisher      *hubnotify.Publisher
}

func NewServiceContext(c config.Config) *ServiceContext {
	gdb, err := gorm.Open(mysql.Open(dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone)), &gorm.Config{})
	if err != nil {
		panic(err)
	}
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
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

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

	svcCtx := &ServiceContext{
		Config:              c,
		AdminUsers:          store.NewAdminUsersStore(gdb),
		AdminRbac:           store.NewAdminRbacStore(gdb),
		AdminRbacCfg:        store.NewAdminRbacConfigStore(gdb),
		GlobalSettings:      store.NewGlobalSettingsStore(gdb),
		PortalNotifications: portalNotifStore,
		NotifyPublisher:     notifyPublisher,
	}
	return svcCtx
}
