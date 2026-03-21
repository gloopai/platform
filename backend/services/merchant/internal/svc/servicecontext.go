package svc

import (
	"database/sql"

	consulconfig "github.com/gloopai/pay/common/consul/config"
	"github.com/gloopai/pay/merchant/internal/config"
	"github.com/gloopai/pay/merchant/internal/store"
	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config        config.Config
	Sql           *sql.DB
	Merchants     *store.MerchantsStore
	RuntimeConfig *consulconfig.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", c.Mysql.DataSource)
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	var runtimeCfg *consulconfig.Store
	if cfg, err := consulconfig.NewStore("", consulconfig.GlobalPrefix(), consulconfig.ServicePrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}
	return &ServiceContext{
		Config:        c,
		Sql:           sqlDB,
		Merchants:     store.NewMerchantsStore(sqlDB),
		RuntimeConfig: runtimeCfg,
	}
}
