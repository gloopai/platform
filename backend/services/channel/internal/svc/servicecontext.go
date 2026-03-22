package svc

import (
	"database/sql"

	"github.com/gloopai/pay/channel/internal/config"
	"github.com/gloopai/pay/channel/internal/store"
	"github.com/gloopai/pay/common/consulx"
	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config        config.Config
	Sql           *sql.DB
	Store         *store.ChannelsStore
	RuntimeConfig *consulx.ConfigStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", c.Mysql.DataSource)
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
		Config:        c,
		Sql:           sqlDB,
		Store:         store.NewChannelsStore(sqlDB),
		RuntimeConfig: runtimeCfg,
	}
}
