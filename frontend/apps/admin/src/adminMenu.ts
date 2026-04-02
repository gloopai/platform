/**
 * 管理端导航（脚手架）：权限与安全、系统与运维。
 */

export type AdminMenuLeaf = {
  kind: 'leaf'
  to: string
  label: string
  icon: string
}

export type AdminMenuGroup = {
  kind: 'group'
  key: string
  label: string
  icon: string
  children: { to: string; label: string }[]
}

export type AdminMenuEntry = AdminMenuLeaf | AdminMenuGroup

export const defaultAdminMenu: AdminMenuEntry[] = [
  {
    kind: 'leaf',
    to: '/home',
    label: '工作台',
    icon: 'chart',
  },
  {
    kind: 'group',
    key: 'rbac',
    label: '权限与安全',
    icon: 'shield',
    children: [
      { to: '/rbac/overview', label: '配置总览' },
      { to: '/rbac/menus', label: '菜单管理' },
      { to: '/rbac/features', label: '功能点' },
      { to: '/rbac/api-rules', label: '接口规则' },
      { to: '/rbac/roles', label: '角色与授权' },
      { to: '/rbac/admin-users', label: '后台用户' },
    ],
  },
  {
    kind: 'group',
    key: 'system',
    label: '系统与运维',
    icon: 'cog',
    children: [
      { to: '/system', label: '系统管理' },
      { to: '/system/op-logs', label: '操作日志' },
      { to: '/ops', label: '运维监控' },
      { to: '/scheduled-jobs', label: '定时任务' },
      { to: '/scheduled-job-runs', label: '任务日志' },
    ],
  },
]

export const adminPathTitle: Record<string, string> = {
  '/home': '工作台',
  '/system': '系统管理',
  '/system/op-logs': '操作日志',
  '/ops': '运维监控',
  '/scheduled-jobs': '定时任务',
  '/scheduled-job-runs': '任务日志',
  '/rbac': '权限与安全',
  '/rbac/overview': '配置总览',
  '/rbac/menus': '菜单管理',
  '/rbac/features': '功能点',
  '/rbac/api-rules': '接口规则',
  '/rbac/roles': '角色与授权',
  '/rbac/admin-users': '后台用户',
}

export const adminPlaceholderMeta: Record<
  string,
  {
    summary: string
    bullets: string[]
    apiNote: string
  }
> = {}

export function pathBelongsToGroup(path: string, group: AdminMenuGroup): boolean {
  return group.children.some((c) => c.to === path)
}

export function findGroupKeyForPath(path: string, menu: AdminMenuEntry[] = defaultAdminMenu): string | null {
  for (const e of menu) {
    if (e.kind === 'group' && pathBelongsToGroup(path, e)) return e.key
  }
  return null
}
