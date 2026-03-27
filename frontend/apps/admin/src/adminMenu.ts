/**
 * 聚合支付 · 总管理台导航与页面说明
 * 与 docs/admin-platform.md 对应；未实现接口在占位页标注「待接入」。
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

/** 侧栏顺序：单页 + 多级分组 */
export const defaultAdminMenu: AdminMenuEntry[] = [
  {
    kind: 'leaf',
    to: '/stats',
    label: '系统概览',
    icon: 'chart',
  },
  {
    kind: 'group',
    key: 'merchant',
    label: '商户管理',
    icon: 'briefcase',
    children: [
      { to: '/merchants', label: '商户列表' },
      { to: '/merchants/deposit', label: '资金存入' },
      { to: '/settlement/logs', label: '资金流水' },
      { to: '/settlement/withdrawals', label: '提现申请' },
      { to: '/settlement/withdrawals/list', label: '提现申请列表' },
    ],
  },
  {
    kind: 'group',
    key: 'trade',
    label: '交易与资金',
    icon: 'credit',
    children: [
      { to: '/payin-orders', label: '代收订单' },
      { to: '/payout-orders', label: '代付订单' },
      { to: '/refunds', label: '退款与差错' },
      { to: '/reconcile', label: '对账中心' },
    ],
  },
  {
    kind: 'group',
    key: 'channel',
    label: '产品与通道',
    icon: 'layers',
    children: [
      { to: '/merchant-payin-products', label: '代收产品' },
      { to: '/merchant-payout-products', label: '代付产品' },
      { to: '/channels', label: '通道管理' },
      { to: '/routing', label: '路由策略' },
      { to: '/channel-health', label: '通道监控' },
    ],
  },
  {
    kind: 'group',
    key: 'rbac',
    label: '权限与安全',
    icon: 'shield',
    children: [
      { to: '/rbac/overview', label: '配置总览' },
      { to: '/rbac/menus', label: '菜单管理' },
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
      { to: '/ops', label: '运维监控' },
    ],
  },
]

/** 顶栏面包屑标题 */
export const adminPathTitle: Record<string, string> = {
  '/stats': '系统概览',
  '/merchants': '商户列表',
  '/merchants/deposit': '资金存入',
  '/merchant-payin-products': '代收产品',
  '/merchant-payout-products': '代付产品与通道',
  '/channels': '通道管理',
  '/routing': '路由策略',
  '/channel-health': '通道监控',
  '/payin-orders': '代收订单',
  '/payout-orders': '代付订单',
  '/refunds': '退款与差错',
  '/reconcile': '对账中心',
  '/settlement': '商户提现',
  '/settlement/logs': '资金流水',
  '/settlement/withdrawals': '提现申请',
  '/settlement/withdrawals/list': '提现申请列表',
  '/system': '系统管理',
  '/ops': '运维监控',
  '/rbac': '权限与安全',
  '/rbac/overview': '配置总览',
  '/rbac/menus': '菜单管理',
  '/rbac/features': '功能点',
  '/rbac/api-rules': '接口规则',
  '/rbac/roles': '角色与授权',
  '/rbac/admin-users': '后台用户',
}

/** 占位页文案（路径 -> 说明）；当前无路由使用通用占位页，保留结构供后续模块。 */
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
