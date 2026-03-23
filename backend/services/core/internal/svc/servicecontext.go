package svc

import (
	"database/sql"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/core/internal/config"
	"github.com/gloopai/pay/core/internal/store"
	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config                 config.Config
	Sql                    *sql.DB
	Merchants              *store.MerchantsStore
	MerchantPayProducts    *store.MerchantPayinProductsStore
	MerchantPayoutProducts *store.MerchantPayoutProductsStore
	Settle                 *store.SettleStore
	RuntimeConfig          *consulx.ConfigStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", dbdsn.WithTimezone(c.Mysql.DataSource, c.Timezone))
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	var runtimeCfg *consulx.ConfigStore
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}
	return &ServiceContext{
		Config:                 c,
		Sql:                    sqlDB,
		Merchants:              store.NewMerchantsStore(sqlDB),
		MerchantPayProducts:    store.NewMerchantPayinProductsStore(sqlDB),
		MerchantPayoutProducts: store.NewMerchantPayoutProductsStore(sqlDB),
		Settle:                 store.NewSettleStore(sqlDB),
		RuntimeConfig:          runtimeCfg,
	}
}
