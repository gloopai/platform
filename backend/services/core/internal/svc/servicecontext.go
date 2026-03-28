package svc

import (
	"context"
	"fmt"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/setup"
	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/core/internal/channelbind"
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
	MerchantSnapshot            *kvcache.MerchantSnapshot
	ChannelSnapshot             *kvcache.ChannelSnapshot
	PayinProductSnapshot        *kvcache.PayinProductSnapshot
	PayoutProductSnapshot       *kvcache.PayoutProductSnapshot
	MerchantPayinGrantsSnapshot *kvcache.MerchantPayinGrantsSnapshot
	MerchantPayoutGrantsSnapshot *kvcache.MerchantPayoutGrantsSnapshot
	PayinProductBindingsSnapshot *kvcache.PayinProductBindingsSnapshot
	PayoutProductBindingsSnapshot *kvcache.PayoutProductBindingsSnapshot
	// ChannelDrivers is owned only by core: RegisterChannelDriver + GetChannelDriver (channel bind resolver).
	// Gateway and trade should use channel gRPC to core instead of embedding a Registry long term.
	ChannelDrivers *channeldriver.Registry

	channelBindResolver channeldriver.ChannelResolver
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
	var merPayinGrants *kvcache.MerchantPayinGrantsSnapshot
	var merPayoutGrants *kvcache.MerchantPayoutGrantsSnapshot
	var payinBindSnap *kvcache.PayinProductBindingsSnapshot
	var payoutBindSnap *kvcache.PayoutProductBindingsSnapshot
	if cfg, err := consulx.NewConfigStore("", configkv.GlobalConfigPrefix(), configkv.ServiceConfigPrefix(c.Name)); err == nil {
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
		merPayinGrants = kvcache.NewMerchantPayinGrantsSnapshot(cfg)
		merPayinGrants.Start(context.Background())
		merPayoutGrants = kvcache.NewMerchantPayoutGrantsSnapshot(cfg)
		merPayoutGrants.Start(context.Background())
		payinBindSnap = kvcache.NewPayinProductBindingsSnapshot(cfg)
		payinBindSnap.Start(context.Background())
		payoutBindSnap = kvcache.NewPayoutProductBindingsSnapshot(cfg)
		payoutBindSnap.Start(context.Background())
	}
	reg := channeldriver.NewRegistry()
	_ = setup.RegisterDefaultMockPSPs(reg)
	chStore := store.NewChannelsStore(gdb)
	bindRes := channelbind.NewResolver(chStore, channelSnap)
	return &ServiceContext{
		Config:                 c,
		Gorm:                   gdb,
		Merchants:              store.NewMerchantsStore(gdb),
		MerchantPayinProducts:  store.NewMerchantPayinProductsStore(gdb),
		MerchantPayoutProducts: store.NewMerchantPayoutProductsStore(gdb),
		Settle:                 store.NewSettleStore(gdb),
		Channels:               chStore,
		PayinProducts:          store.NewPayinProductsStore(gdb),
		PayoutProducts:         store.NewPayoutProductsStore(gdb),
		RoutingSummary:         store.NewRoutingSummaryStore(gdb),
		RuntimeConfig:          runtimeCfg,
		MerchantSnapshot:               merchantSnap,
		ChannelSnapshot:                channelSnap,
		PayinProductSnapshot:           payinProdSnap,
		PayoutProductSnapshot:          payoutProdSnap,
		MerchantPayinGrantsSnapshot:    merPayinGrants,
		MerchantPayoutGrantsSnapshot:   merPayoutGrants,
		PayinProductBindingsSnapshot:   payinBindSnap,
		PayoutProductBindingsSnapshot:  payoutBindSnap,
		ChannelDrivers:                 reg,
		channelBindResolver:            bindRes,
	}
}

// GetChannelDriver returns a cached [channeldriver.ChannelDriver] for one channel row.
// This is the supported entrypoint for channel_id + merged config inside core.
func (s *ServiceContext) GetChannelDriver(ctx context.Context, channelID int64) (channeldriver.ChannelDriver, error) {
	if s == nil || s.ChannelDrivers == nil {
		return nil, fmt.Errorf("svc: ChannelDrivers not configured")
	}
	if s.channelBindResolver == nil {
		return nil, fmt.Errorf("svc: channel bind resolver not configured")
	}
	return s.ChannelDrivers.GetChannelDriver(ctx, channelID, s.channelBindResolver)
}

// InvalidateChannelDriverCache drops the in-process channel driver cache for a row after
// admin updates channel_config (or equivalent).
func (s *ServiceContext) InvalidateChannelDriverCache(channelID int64) {
	if s == nil || s.ChannelDrivers == nil {
		return
	}
	s.ChannelDrivers.InvalidateChannelDriver(channelID)
}

// OpenAPIMemoryReady is true when Consul-backed routing snapshots are wired (hot path can avoid DB).
func (s *ServiceContext) OpenAPIMemoryReady() bool {
	if s == nil || s.RuntimeConfig == nil {
		return false
	}
	return s.PayinProductSnapshot != nil &&
		s.PayinProductBindingsSnapshot != nil &&
		s.ChannelSnapshot != nil &&
		s.MerchantPayinGrantsSnapshot != nil
}
