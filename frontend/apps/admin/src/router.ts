import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import AuditPage from './views/pages/AuditPage.vue'
import ChannelsPage from './views/pages/ChannelsPage.vue'
import MerchantsPage from './views/pages/MerchantsPage.vue'
import PayProductsPage from './views/pages/PayProductsPage.vue'
import ModulePlaceholderPage from './views/pages/ModulePlaceholderPage.vue'
import StatsPage from './views/pages/StatsPage.vue'

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
        { path: 'routing', component: ModulePlaceholderPage },
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
