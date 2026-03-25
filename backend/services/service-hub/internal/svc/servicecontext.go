package svc

import (
	"time"

	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/service-hub/internal/config"
	"github.com/gloopai/pay/service-hub/internal/store"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	AdminUsers     *store.AdminUsersStore
	GlobalSettings *store.GlobalSettingsStore
	PayoutOrders   *store.PayoutOrdersStore
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
	return &ServiceContext{
		Config:         c,
		AdminUsers:     store.NewAdminUsersStore(gdb),
		GlobalSettings: store.NewGlobalSettingsStore(gdb),
		PayoutOrders:   store.NewPayoutOrdersStore(gdb),
	}
}

