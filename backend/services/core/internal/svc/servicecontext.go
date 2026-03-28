package svc

import (
	"context"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/setup"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/core/internal/config"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                 config.Config
	Gorm                   *gorm.DB
	Merchants              *store.MerchantsStore
	MerchantPayinProducts  *store.MerchantPayinProductsStore
	MerchantPayoutProducts *store.MerchantPayoutProductsStore
	Settle                 *store.SettleStore
	Channels               *store.ChannelsStore
	PayinProducts          *store.PayinProductsStore
	PayoutProducts         *store.PayoutProductsStore
	RoutingSummary         *store.RoutingSummaryStore
	RuntimeConfig          *consulx.ConfigStore
	MerchantSnapshot       *kvcache.MerchantSnapshot
	ChannelSnapshot        *kvcache.ChannelSnapshot
	PayinProductSnapshot   *kvcache.PayinProductSnapshot
	PayoutProductSnapshot  *kvcache.PayoutProductSnapshot
	ChannelDrivers         *channeldriver.Registry
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
	var runtimeCfg *consulx.ConfigStore
	var merchantSnap *kvcache.MerchantSnapshot
	var channelSnap *kvcache.ChannelSnapshot
	var payinProdSnap *kvcache.PayinProductSnapshot
	var payoutProdSnap *kvcache.PayoutProductSnapshot
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
		merchantSnap = kvcache.NewMerchantSnapshot(cfg)
		merchantSnap.Start(context.Background())
		channelSnap = kvcache.NewChannelSnapshot(cfg)
		channelSnap.Start(context.Background())
		payinProdSnap = kvcache.NewPayinProductSnapshot(cfg)
		payinProdSnap.Start(context.Background())
		payoutProdSnap = kvcache.NewPayoutProductSnapshot(cfg)
		payoutProdSnap.Start(context.Background())
	}
	reg := channeldriver.NewRegistry()
	_ = setup.RegisterDefaultMockPSPs(reg)
	return &ServiceContext{
		Config:                 c,
		Gorm:                   gdb,
		Merchants:              store.NewMerchantsStore(gdb),
		MerchantPayinProducts:  store.NewMerchantPayinProductsStore(gdb),
		MerchantPayoutProducts: store.NewMerchantPayoutProductsStore(gdb),
		Settle:                 store.NewSettleStore(gdb),
		Channels:               store.NewChannelsStore(gdb),
		PayinProducts:          store.NewPayinProductsStore(gdb),
		PayoutProducts:         store.NewPayoutProductsStore(gdb),
		RoutingSummary:         store.NewRoutingSummaryStore(gdb),
		RuntimeConfig:          runtimeCfg,
		MerchantSnapshot:       merchantSnap,
		ChannelSnapshot:        channelSnap,
		PayinProductSnapshot:   payinProdSnap,
		PayoutProductSnapshot:  payoutProdSnap,
		ChannelDrivers:         reg,
	}
}
