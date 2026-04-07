import { createRouter, createWebHistory } from 'vue-router'

import {
  adminGet,
  clearAdminSession,
  readAdminMfaGate,
  setAdminMfaGate,
} from './lib/adminApi'
import LoginPage from './views/LoginPage.vue'
import MfaSetupPage from './views/MfaSetupPage.vue'
import AdminLayout from './views/AdminLayout.vue'
import HomePage from './views/HomePage.vue'
import OpsPage from './views/modules/ops/OpsPage.vue'
import SystemPage from './views/modules/system/SystemPage.vue'
import RbacLayout from './views/modules/rbac/RbacLayout.vue'
import MenuManagementPage from './views/modules/rbac/menu-management/MenuManagementPage.vue'
import RbacRolesPage from './views/modules/rbac/RbacRolesPage.vue'
import RbacAdminUsersPage from './views/modules/rbac/RbacAdminUsersPage.vue'
import RbacFeaturePointsPage from './views/modules/rbac/RbacFeaturePointsPage.vue'
import RbacApiRulesPage from './views/modules/rbac/RbacApiRulesPage.vue'
import RbacOverviewPage from './views/modules/rbac/RbacOverviewPage.vue'
import OperationLogsPage from './views/modules/system/OperationLogsPage.vue'
import ScheduledJobsPage from './views/modules/jobs/ScheduledJobsPage.vue'
import JobWorkerNodesPage from './views/modules/jobs/JobWorkerNodesPage.vue'
import ScheduledJobRunsPage from './views/modules/jobs/ScheduledJobRunsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage },
    { path: '/mfa-setup', component: MfaSetupPage },
    {
      path: '/',
      component: AdminLayout,
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', component: HomePage },
        { path: 'system', component: SystemPage },
        { path: 'system/op-logs', component: OperationLogsPage },
        { path: 'scheduled-jobs', component: ScheduledJobsPage },
        { path: 'job-worker-nodes', component: JobWorkerNodesPage },
        { path: 'scheduled-job-runs', component: ScheduledJobRunsPage },
        { path: 'ops', component: OpsPage },
        {
          path: 'rbac',
          component: RbacLayout,
          redirect: '/rbac/overview',
          children: [
            { path: 'overview', component: RbacOverviewPage },
            { path: 'menus', component: MenuManagementPage },
            { path: 'roles', component: RbacRolesPage },
            { path: 'features', component: RbacFeaturePointsPage },
            { path: 'api-rules', component: RbacApiRulesPage },
            { path: 'admin-users', component: RbacAdminUsersPage },
            { path: 'permissions', redirect: '/rbac/features' },
          ],
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.path === '/login') return true

  const tok = localStorage.getItem('admin_token')
  if (!tok) {
    return '/login'
  }

  const gate = readAdminMfaGate()
  if (gate) {
    if (!gate.ok) {
      if (to.path !== '/mfa-setup') return '/mfa-setup'
      return true
    }
    if (to.path === '/mfa-setup') return '/home'
  } else {
    try {
      const me = await adminGet<{ mfa_enabled: number }>('/v1/admin/me')
      const complete = me.mfa_enabled === 1
      setAdminMfaGate(complete)
      if (!complete) {
        if (to.path !== '/mfa-setup') return '/mfa-setup'
        return true
      }
      if (to.path === '/mfa-setup') return '/home'
    } catch {
      clearAdminSession()
      localStorage.removeItem('admin_allowed_paths')
      return '/login'
    }
  }

  try {
    const raw = localStorage.getItem('admin_allowed_paths')
    if (raw) {
      const allowed = JSON.parse(raw) as string[]
      if (Array.isArray(allowed) && allowed.length) {
        // 工作台为默认落地页，避免因缓存里尚未写入 /home 而被误拦
        if (to.path === '/home') return true
        if (
          (to.path === '/rbac/features' || to.path === '/rbac/api-rules') &&
          allowed.includes('/rbac/menus')
        ) {
          return true
        }
        if (!allowed.includes(to.path)) return allowed[0]
      }
    }
  } catch {
  }
  return true
})
