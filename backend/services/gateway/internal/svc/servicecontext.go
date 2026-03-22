// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"

	"github.com/gloopai/pay/channel/channelclient"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/store"
	"github.com/gloopai/pay/merchant/merchantclient"
	"github.com/gloopai/pay/order/orderclient"
	"github.com/gloopai/pay/settle/settleclient"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	MerchantSignMiddleware        rest.Middleware
	AdminAuthMiddleware           rest.Middleware
	MerchantConsoleAuthMiddleware rest.Middleware

	Channels   *store.ChannelsStore
	FundLogs   *store.FundLogsStore
	NotifyLogs *store.NotifyLogsStore
	AdminUsers *store.AdminUsersStore
	Sessions   *store.SessionsStore

	OrderRpc    orderclient.Order
	SettleRpc   settleclient.Settle
	ChannelRpc  channelclient.Channel
	MerchantRpc merchantclient.Merchant

	NsqProducer *nsq.Producer

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

	consulx.RegisterResolver()

	orderCli := zrpc.MustNewClient(c.OrderRpc)
	settleCli := zrpc.MustNewClient(c.SettleRpc)
	channelCli := zrpc.MustNewClient(c.ChannelRpc)
	merchantCli := zrpc.MustNewClient(c.MerchantRpc)

	producer, err := nsq.NewProducer(c.Nsq.NsqdTCPAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	if err := producer.Ping(); err != nil {
		panic(err)
	}

	channelsStore := store.NewChannelsStore(sqlDB)
	fundLogsStore := store.NewFundLogsStore(sqlDB)
	notifyLogsStore := store.NewNotifyLogsStore(sqlDB)
	adminUsersStore := store.NewAdminUsersStore(sqlDB)
	sessionsStore := store.NewSessionsStore(sqlDB)
	var runtimeCfg *consulx.ConfigStore
	if cfg, err := consulx.NewConfigStore("", consulx.GlobalConfigPrefix(), consulx.ServiceConfigPrefix(c.Name)); err == nil {
		cfg.Start()
		runtimeCfg = cfg
	}

	return &ServiceContext{
		Config: c,

		MerchantSignMiddleware:        middleware.NewMerchantSignMiddleware(merchantclient.NewMerchant(merchantCli)).Handle,
		AdminAuthMiddleware:           middleware.NewAdminAuthMiddleware(c.AdminToken, sessionsStore).Handle,
		MerchantConsoleAuthMiddleware: middleware.NewMerchantConsoleAuthMiddleware(sessionsStore, merchantclient.NewMerchant(merchantCli)).Handle,

		Channels:   channelsStore,
		FundLogs:   fundLogsStore,
		NotifyLogs: notifyLogsStore,
		AdminUsers: adminUsersStore,
		Sessions:   sessionsStore,

		OrderRpc:    orderclient.NewOrder(orderCli),
		SettleRpc:   settleclient.NewSettle(settleCli),
		ChannelRpc:  channelclient.NewChannel(channelCli),
		MerchantRpc: merchantclient.NewMerchant(merchantCli),

		NsqProducer:   producer,
		RuntimeConfig: runtimeCfg,
	}
}
