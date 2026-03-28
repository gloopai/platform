package svc

import (
	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/mockpsp"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/trade/internal/config"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                config.Config
	Gorm                  *gorm.DB
	Redis                 *redis.Client
	PayOrders             *store.PayinOrdersStore
	PayoutOrders          *store.PayoutOrdersStore
	Channels              *store.ChannelsStore
	MerchantPayinProducts *store.MerchantPayinProductsStore
	PayinProducts         *store.PayinProductsStore
	PayoutProducts        *store.PayoutProductsStore
	OrderStats            *store.OrderStatsStore
	RoutingSummary        *store.RoutingSummaryStore
	NotifyLogs            *store.NotifyLogsStore
	RuntimeConfig         *consulx.ConfigStore
	ChannelDrivers        *channeldriver.Registry
}

func NewServiceContext(c config.Config) *ServiceContext {
	gdb, err := gorm.Open(mysql.Open(dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone)), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if sqlDB, err := gdb.DB(); err != nil {
		panic(err)
	} else if err := sqlDB.Ping(); err != nil {
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
	reg := channeldriver.NewRegistry()
	_ = mockpsp.RegisterAll(reg, mockpsp.New(mockpsp.DefaultDriverKey))
	return &ServiceContext{
		Config:                c,
		Gorm:                  gdb,
		Redis:                 rdb,
		PayOrders:             store.NewPayinOrdersStore(gdb),
		PayoutOrders:          store.NewPayoutOrdersStore(gdb),
		Channels:              store.NewChannelsStore(gdb),
		MerchantPayinProducts: store.NewMerchantPayinProductsStore(gdb),
		PayinProducts:         store.NewPayinProductsStore(gdb),
		PayoutProducts:        store.NewPayoutProductsStore(gdb),
		OrderStats:            store.NewOrderStatsStore(gdb),
		RoutingSummary:        store.NewRoutingSummaryStore(gdb),
		NotifyLogs:            store.NewNotifyLogsStore(gdb),
		RuntimeConfig:         runtimeCfg,
		ChannelDrivers:        reg,
	}
}
