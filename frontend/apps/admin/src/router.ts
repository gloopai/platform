import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import ChannelHealthPage from './views/modules/channel-health/ChannelHealthPage.vue'
import ChannelsPage from './views/modules/channels/ChannelsPage.vue'
import MerchantsPage from './views/modules/merchants/MerchantsPage.vue'
import PayinPayProductsPage from './views/modules/pay-products/PayinPayProductsPage.vue'
import PayoutPayProductsPage from './views/modules/pay-products/PayoutPayProductsPage.vue'
import RouteStrategyPage from './views/modules/routing/RouteStrategyPage.vue'
import PayOrdersPage from './views/modules/orders/PayOrdersPage.vue'
import PayoutOrdersPage from './views/modules/orders/PayoutOrdersPage.vue'
import OpsPage from './views/modules/ops/OpsPage.vue'
import RefundsPage from './views/modules/refunds/RefundsPage.vue'
import ReconcilePage from './views/modules/reconcile/ReconcilePage.vue'
import SettlementPage from './views/modules/settlement/SettlementPage.vue'
import StatsPage from './views/modules/stats/StatsPage.vue'
import SystemPage from './views/modules/system/SystemPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage },
    {
      path: '/',
      component: AdminLayout,
      children: [
        { path: '', redirect: '/stats' },
        { path: 'stats', component: StatsPage },
        { path: 'merchants', component: MerchantsPage },
        { path: 'channels', component: ChannelsPage },
        { path: 'merchant-payin-products', component: PayinPayProductsPage },
        { path: 'merchant-payout-products', component: PayoutPayProductsPage },
        { path: 'routing', component: RouteStrategyPage },
        { path: 'channel-health', component: ChannelHealthPage },
        { path: 'payin-orders', component: PayOrdersPage },
        { path: 'payout-orders', component: PayoutOrdersPage },
        { path: 'refunds', component: RefundsPage },
        { path: 'reconcile', component: ReconcilePage },
        { path: 'settlement', component: SettlementPage },
        { path: 'system', component: SystemPage },
        { path: 'ops', component: OpsPage },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.path === '/login') return true
  const tok = localStorage.getItem('admin_token')
  if (!tok) return '/login'
  return true
})
