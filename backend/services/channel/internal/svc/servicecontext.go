package svc

import (
	"database/sql"

	"github.com/gloopai/pay/channel/internal/config"
	"github.com/gloopai/pay/channel/internal/store"
	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config config.Config
	Sql    *sql.DB
	Store  *store.ChannelsStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", c.Mysql.DataSource)
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		Sql:    sqlDB,
		Store:  store.NewChannelsStore(sqlDB),
	}
}
