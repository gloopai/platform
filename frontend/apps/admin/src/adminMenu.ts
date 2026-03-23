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
export const adminMenu: AdminMenuEntry[] = [
  {
    kind: 'leaf',
    to: '/stats',
    label: '系统概览',
    icon: 'chart',
  },
  {
    kind: 'group',
    key: 'merchant',
    label: '商户与接入',
    icon: 'briefcase',
    children: [
      { to: '/merchants', label: '商户管理' },
      { to: '/merchant-products', label: '代收产品与通道' },
      { to: '/merchant-payout-products', label: '代付产品与通道' },
      { to: '/developer-docs', label: '开发文档' },
    ],
  },
  {
    kind: 'group',
    key: 'channel',
    label: '通道与路由',
    icon: 'layers',
    children: [
      { to: '/channels', label: '通道管理' },
      { to: '/routing', label: '路由策略' },
      { to: '/channel-health', label: '通道监控' },
    ],
  },
  {
    kind: 'group',
    key: 'trade',
    label: '交易与资金',
    icon: 'credit',
    children: [
      { to: '/pay-orders', label: '代收订单' },
      { to: '/payout-orders', label: '代付订单' },
      { to: '/refunds', label: '退款与差错' },
      { to: '/reconcile', label: '对账中心' },
      { to: '/settlement', label: '结算与提现' },
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
  '/merchants': '商户管理',
  '/merchant-products': '代收产品与通道',
  '/merchant-payout-products': '代付产品与通道',
  '/developer-docs': '开发文档',
  '/channels': '通道管理',
  '/routing': '路由策略',
  '/channel-health': '通道监控',
  '/pay-orders': '代收订单',
  '/payout-orders': '代付订单',
  '/refunds': '退款与差错',
  '/reconcile': '对账中心',
  '/settlement': '结算与提现',
  '/system': '系统管理',
  '/ops': '运维监控',
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

export function findGroupKeyForPath(path: string): string | null {
  for (const e of adminMenu) {
    if (e.kind === 'group' && pathBelongsToGroup(path, e)) return e.key
  }
  return null
}
