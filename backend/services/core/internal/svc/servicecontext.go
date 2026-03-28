package svc

import (
	"context"
	"fmt"

	"github.com/gloopai/pay/common/configkv"
	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/dbdsn"
	"github.com/gloopai/pay/core/internal/channelbridge"
	"github.com/gloopai/pay/core/internal/channelbridge/psp"
	ct "github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/config"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                        config.Config
	Gorm                          *gorm.DB
	Merchants                     *store.MerchantsStore
	MerchantPayinProducts         *store.MerchantPayinProductsStore
	MerchantPayoutProducts        *store.MerchantPayoutProductsStore
	Settle                        *store.SettleStore
	Channels                      *store.ChannelsStore
	PayinProducts                 *store.PayinProductsStore
	PayoutProducts                *store.PayoutProductsStore
	RoutingSummary                *store.RoutingSummaryStore
	RuntimeConfig                 *consulx.ConfigStore
	MerchantSnapshot              *kvcache.MerchantSnapshot
	ChannelSnapshot               *kvcache.ChannelSnapshot
	PayinProductSnapshot          *kvcache.PayinProductSnapshot
	PayoutProductSnapshot         *kvcache.PayoutProductSnapshot
	MerchantPayinGrantsSnapshot   *kvcache.MerchantPayinGrantsSnapshot
	MerchantPayoutGrantsSnapshot  *kvcache.MerchantPayoutGrantsSnapshot
	PayinProductBindingsSnapshot  *kvcache.PayinProductBindingsSnapshot
	PayoutProductBindingsSnapshot *kvcache.PayoutProductBindingsSnapshot
	// ChannelBridge owns routing (KV/DB) + PSP registry + channel resolver for this process.
	ChannelBridge *channelbridge.Bridge
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
	chStore := store.NewChannelsStore(gdb)
	reg := psp.NewRegistry()
	_ = psp.RegisterBuiltInDrivers(reg, chStore, channelSnap)
	bindRes := channelbridge.NewResolver(chStore, channelSnap)
	payinProdStore := store.NewPayinProductsStore(gdb)
	bridge := channelbridge.NewBridge(channelbridge.BridgeConfig{
		Channels:                     chStore,
		PayinProducts:                payinProdStore,
		Registry:                     reg,
		Resolver:                     bindRes,
		RuntimeConfig:                runtimeCfg,
		PayinProductSnapshot:         payinProdSnap,
		PayinProductBindingsSnapshot: payinBindSnap,
		ChannelSnapshot:              channelSnap,
		MerchantPayinGrantsSnapshot:  merPayinGrants,
	})
	return &ServiceContext{
		Config:                        c,
		Gorm:                          gdb,
		Merchants:                     store.NewMerchantsStore(gdb),
		MerchantPayinProducts:         store.NewMerchantPayinProductsStore(gdb),
		MerchantPayoutProducts:        store.NewMerchantPayoutProductsStore(gdb),
		Settle:                        store.NewSettleStore(gdb),
		Channels:                      chStore,
		PayinProducts:                 payinProdStore,
		PayoutProducts:                store.NewPayoutProductsStore(gdb),
		RoutingSummary:                store.NewRoutingSummaryStore(gdb),
		RuntimeConfig:                 runtimeCfg,
		MerchantSnapshot:              merchantSnap,
		ChannelSnapshot:               channelSnap,
		PayinProductSnapshot:          payinProdSnap,
		PayoutProductSnapshot:         payoutProdSnap,
		MerchantPayinGrantsSnapshot:   merPayinGrants,
		MerchantPayoutGrantsSnapshot:  merPayoutGrants,
		PayinProductBindingsSnapshot:  payinBindSnap,
		PayoutProductBindingsSnapshot: payoutBindSnap,
		ChannelBridge:                 bridge,
	}
}

// GetChannelDriver returns a cached ChannelDriver (psp/contracts) for one channel row.
// This is the supported entrypoint for channel_id + merged config inside core.
func (s *ServiceContext) GetChannelDriver(ctx context.Context, channelID int64) (ct.ChannelDriver, error) {
	if s == nil || s.ChannelBridge == nil {
		return nil, fmt.Errorf("svc: ChannelBridge not configured")
	}
	return s.ChannelBridge.GetDriver(ctx, channelID)
}

// InvalidateChannelDriverCache drops the in-process channel driver cache for a row after
// admin updates channel_config (or equivalent).
func (s *ServiceContext) InvalidateChannelDriverCache(channelID int64) {
	if s == nil || s.ChannelBridge == nil {
		return
	}
	s.ChannelBridge.InvalidateDriverCache(channelID)
}

// OpenAPIMemoryReady is true when Consul-backed routing snapshots are wired (hot path can avoid DB).
func (s *ServiceContext) OpenAPIMemoryReady() bool {
	return s != nil && s.ChannelBridge != nil && s.ChannelBridge.MemoryReady()
}
