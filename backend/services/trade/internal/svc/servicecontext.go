package svc

import (
	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/mockpsp"
	"github.com/gloopai/pay/channeldriver/mockpsp2"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/common/grpcclient/channelclient"
	"github.com/gloopai/pay/trade/internal/config"
	"github.com/gloopai/pay/trade/internal/store"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	Gorm           *gorm.DB
	Redis          *redis.Client
	PayOrders      *store.PayinOrdersStore
	PayoutOrders   *store.PayoutOrdersStore
	OrderStats     *store.OrderStatsStore
	NotifyLogs     *store.NotifyLogsStore
	ChannelDrivers *channeldriver.Registry
	ChannelRpc     channelclient.Channel
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
	coreCli := zrpc.MustNewClient(c.CoreRpc)
	reg := channeldriver.NewRegistry()
	_ = mockpsp.RegisterAll(reg, mockpsp.New(mockpsp.DefaultDriverKey))
	_ = mockpsp2.RegisterAll(reg, mockpsp2.New(mockpsp2.DefaultDriverKey))
	return &ServiceContext{
		Config:         c,
		Gorm:           gdb,
		Redis:          rdb,
		PayOrders:      store.NewPayinOrdersStore(gdb),
		PayoutOrders:   store.NewPayoutOrdersStore(gdb),
		OrderStats:     store.NewOrderStatsStore(gdb),
		NotifyLogs:     store.NewNotifyLogsStore(gdb),
		ChannelDrivers: reg,
		ChannelRpc:     channelclient.NewChannel(coreCli),
	}
}
