package svc

import (
	"database/sql"

	"github.com/gloopai/pay/settle/internal/config"
	"github.com/gloopai/pay/settle/internal/store"
	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config config.Config
	Sql    *sql.DB
	Settle *store.SettleStore
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
		Settle: store.NewSettleStore(sqlDB),
	}
}
