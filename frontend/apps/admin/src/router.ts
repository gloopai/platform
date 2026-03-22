import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import AuditPage from './views/pages/AuditPage.vue'
import ChannelsPage from './views/modules/channels/ChannelsPage.vue'
import MerchantsPage from './views/modules/merchants/MerchantsPage.vue'
import PayProductsPage from './views/modules/pay-products/PayProductsPage.vue'
import RouteStrategyPage from './views/modules/routing/RouteStrategyPage.vue'
import ModulePlaceholderPage from './views/pages/ModulePlaceholderPage.vue'
import StatsPage from './views/modules/stats/StatsPage.vue'

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
        { path: 'merchant-products', component: PayProductsPage },
        { path: 'routing', component: RouteStrategyPage },
        { path: 'channel-health', component: ModulePlaceholderPage },
        { path: 'orders', component: ModulePlaceholderPage },
        { path: 'refunds', component: ModulePlaceholderPage },
        { path: 'reconcile', component: ModulePlaceholderPage },
        { path: 'settlement', component: ModulePlaceholderPage },
        { path: 'risk', component: ModulePlaceholderPage },
        { path: 'audit', component: AuditPage },
        { path: 'notifications', component: ModulePlaceholderPage },
        { path: 'system', component: ModulePlaceholderPage },
        { path: 'ops', component: ModulePlaceholderPage },
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
