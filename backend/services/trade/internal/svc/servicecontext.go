package svc

import (
	"database/sql"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/trade/internal/config"
	"github.com/gloopai/pay/trade/internal/store"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config              config.Config
	Sql                 *sql.DB
	Redis               *redis.Client
	PayOrders           *store.PayinOrdersStore
	PayoutOrders        *store.PayoutOrdersStore
	Channels            *store.ChannelsStore
	MerchantPayProducts *store.MerchantPayinProductsStore
	PayProducts         *store.PayinProductsStore
	PayoutProducts      *store.PayoutProductsStore
	OrderStats          *store.OrderStatsStore
	RoutingSummary      *store.RoutingSummaryStore
	NotifyLogs          *store.NotifyLogsStore
	RuntimeConfig       *consulx.ConfigStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone))
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BizRedis.Addr,
		Password: c.BizRedis.Password,
		DB:       c.BizRedis.DB,
	})
	var runtimeCfg *consulx.ConfigStore
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}
	return &ServiceContext{
		Config:              c,
		Sql:                 sqlDB,
		Redis:               rdb,
		PayOrders:           store.NewPayinOrdersStore(sqlDB),
		PayoutOrders:        store.NewPayoutOrdersStore(sqlDB),
		Channels:            store.NewChannelsStore(sqlDB),
		MerchantPayProducts: store.NewMerchantPayinProductsStore(sqlDB),
		PayProducts:         store.NewPayinProductsStore(sqlDB),
		PayoutProducts:      store.NewPayoutProductsStore(sqlDB),
		OrderStats:          store.NewOrderStatsStore(sqlDB),
		RoutingSummary:      store.NewRoutingSummaryStore(sqlDB),
		NotifyLogs:          store.NewNotifyLogsStore(sqlDB),
		RuntimeConfig:       runtimeCfg,
	}
}
