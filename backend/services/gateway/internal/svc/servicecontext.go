// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"

	"github.com/gloopai/pay/channel/channelclient"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/consulresolver"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/order/orderclient"
	"github.com/gloopai/pay/settle/settleclient"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Sql    *sql.DB

	MerchantSignMiddleware rest.Middleware
	AdminAuthMiddleware    rest.Middleware
	MerchantConsoleAuthMiddleware rest.Middleware

	Merchants  *store.MerchantsStore
	Channels   *store.ChannelsStore
	Orders     *store.OrdersStore
	FundLogs   *store.FundLogsStore
	NotifyLogs *store.NotifyLogsStore
	AdminUsers *store.AdminUsersStore
	Sessions   *store.SessionsStore

	OrderRpc   orderclient.Order
	SettleRpc  settleclient.Settle
	ChannelRpc channelclient.Channel

	NsqProducer *nsq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlDB, err := sql.Open("mysql", c.Mysql.DataSource)
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	consulresolver.Register()

	orderCli := zrpc.MustNewClient(c.OrderRpc)
	settleCli := zrpc.MustNewClient(c.SettleRpc)
	channelCli := zrpc.MustNewClient(c.ChannelRpc)

	producer, err := nsq.NewProducer(c.Nsq.NsqdTCPAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	if err := producer.Ping(); err != nil {
		panic(err)
	}

	merchantsStore := store.NewMerchantsStore(sqlDB)
	channelsStore := store.NewChannelsStore(sqlDB)
	ordersStore := store.NewOrdersStore(sqlDB)
	fundLogsStore := store.NewFundLogsStore(sqlDB)
	notifyLogsStore := store.NewNotifyLogsStore(sqlDB)
	adminUsersStore := store.NewAdminUsersStore(sqlDB)
	sessionsStore := store.NewSessionsStore(sqlDB)

	return &ServiceContext{
		Config: c,
		Sql:    sqlDB,

		MerchantSignMiddleware: middleware.NewMerchantSignMiddleware(merchantsStore).Handle,
		AdminAuthMiddleware:    middleware.NewAdminAuthMiddleware(c.AdminToken, sessionsStore).Handle,
		MerchantConsoleAuthMiddleware: middleware.NewMerchantConsoleAuthMiddleware(sessionsStore, merchantsStore).Handle,

		Merchants:  merchantsStore,
		Channels:   channelsStore,
		Orders:     ordersStore,
		FundLogs:   fundLogsStore,
		NotifyLogs: notifyLogsStore,
		AdminUsers: adminUsersStore,
		Sessions:   sessionsStore,

		OrderRpc:   orderclient.NewOrder(orderCli),
		SettleRpc:  settleclient.NewSettle(settleCli),
		ChannelRpc: channelclient.NewChannel(channelCli),

		NsqProducer: producer,
	}
}
