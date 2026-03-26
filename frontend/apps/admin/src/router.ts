import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import ChannelHealthPage from './views/modules/channel-health/ChannelHealthPage.vue'
import ChannelsPage from './views/modules/channels/ChannelsPage.vue'
import MerchantsPage from './views/modules/merchants/MerchantsPage.vue'
import PayinProductsPage from './views/modules/pay-products/PayinProductsPage.vue'
import PayoutProductsPage from './views/modules/pay-products/PayoutProductsPage.vue'
import RouteStrategyPage from './views/modules/routing/RouteStrategyPage.vue'
import PayinOrdersPage from './views/modules/orders/PayinOrdersPage.vue'
import PayoutOrdersPage from './views/modules/orders/PayoutOrdersPage.vue'
import OpsPage from './views/modules/ops/OpsPage.vue'
import RefundsPage from './views/modules/refunds/RefundsPage.vue'
import ReconcilePage from './views/modules/reconcile/ReconcilePage.vue'
import SettlementPage from './views/modules/settlement/SettlementPage.vue'
import StatsPage from './views/modules/stats/StatsPage.vue'
import SystemPage from './views/modules/system/SystemPage.vue'
import RbacLayout from './views/modules/rbac/RbacLayout.vue'
import RbacOverviewPage from './views/modules/rbac/RbacOverviewPage.vue'
import MenuManagementPage from './views/modules/rbac/menu-management/MenuManagementPage.vue'
import RbacRolesPage from './views/modules/rbac/RbacRolesPage.vue'
import RbacAdminUsersPage from './views/modules/rbac/RbacAdminUsersPage.vue'

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
        { path: 'merchant-payin-products', component: PayinProductsPage },
        { path: 'merchant-payout-products', component: PayoutProductsPage },
        { path: 'routing', component: RouteStrategyPage },
        { path: 'channel-health', component: ChannelHealthPage },
        { path: 'payin-orders', component: PayinOrdersPage },
        { path: 'payout-orders', component: PayoutOrdersPage },
        { path: 'refunds', component: RefundsPage },
        { path: 'reconcile', component: ReconcilePage },
        { path: 'settlement', component: SettlementPage },
        { path: 'system', component: SystemPage },
        { path: 'ops', component: OpsPage },
        {
          path: 'rbac',
          component: RbacLayout,
          redirect: '/rbac/overview',
          children: [
            { path: 'overview', component: RbacOverviewPage },
            { path: 'menus', component: MenuManagementPage },
            { path: 'roles', component: RbacRolesPage },
            { path: 'features', redirect: to => ({ path: '/rbac/menus', query: { ...to.query, tab: 'other' } }) },
            { path: 'admin-users', component: RbacAdminUsersPage },
            { path: 'permissions', redirect: '/rbac/menus?tab=other' },
            { path: 'api-rules', redirect: '/rbac/menus?tab=other' },
          ],
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.path === '/login') return true
  const tok = localStorage.getItem('admin_token')
  if (!tok) return '/login'

  try {
    const raw = localStorage.getItem('admin_allowed_paths')
    if (raw) {
      const allowed = JSON.parse(raw) as string[]
      if (Array.isArray(allowed) && allowed.length) {
        if (!allowed.includes(to.path)) return allowed[0]
      }
    }
  } catch {
  }
  return true
})
