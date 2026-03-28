package svc

import (
	"context"

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
	RuntimeConfig          *consulx.ConfigStore
	MerchantSnapshot       *kvcache.MerchantSnapshot
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
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
		merchantSnap = kvcache.NewMerchantSnapshot(cfg)
		merchantSnap.Start(context.Background())
	}
	return &ServiceContext{
		Config:                 c,
		Gorm:                   gdb,
		Merchants:              store.NewMerchantsStore(gdb),
		MerchantPayinProducts:  store.NewMerchantPayinProductsStore(gdb),
		MerchantPayoutProducts: store.NewMerchantPayoutProductsStore(gdb),
		Settle:                 store.NewSettleStore(gdb),
		RuntimeConfig:          runtimeCfg,
		MerchantSnapshot:       merchantSnap,
	}
}
