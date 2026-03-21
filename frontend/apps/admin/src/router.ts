import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import AuditPage from './views/pages/AuditPage.vue'
import ChannelsPage from './views/pages/ChannelsPage.vue'
import MerchantsPage from './views/pages/MerchantsPage.vue'
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
        { path: 'audit', component: AuditPage },
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
