package svc

import (
	"database/sql"

	"github.com/gloopai/pay/order/internal/config"
	"github.com/gloopai/pay/order/internal/store"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config
	Sql    *sql.DB
	Redis  *redis.Client
	Orders *store.OrdersStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", c.Mysql.DataSource)
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
	return &ServiceContext{
		Config: c,
		Sql:    sqlDB,
		Redis:  rdb,
		Orders: store.NewOrdersStore(sqlDB),
	}
}
