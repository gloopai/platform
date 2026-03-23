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
      { to: '/orders', label: '全站订单' },
      { to: '/refunds', label: '退款与差错' },
      { to: '/reconcile', label: '对账中心' },
      { to: '/settlement', label: '结算与提现' },
    ],
  },
  {
    kind: 'group',
    key: 'risk',
    label: '风控与合规',
    icon: 'shield',
    children: [
      { to: '/risk', label: '风控规则' },
      { to: '/audit', label: '运营与审计' },
      { to: '/notifications', label: '公告与通知' },
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
  '/channels': '通道管理',
  '/routing': '路由策略',
  '/channel-health': '通道监控',
  '/orders': '全站订单',
  '/refunds': '退款与差错',
  '/reconcile': '对账中心',
  '/settlement': '结算与提现',
  '/risk': '风控规则',
  '/audit': '运营与审计',
  '/notifications': '公告与通知',
  '/system': '系统管理',
  '/ops': '运维监控',
}

/** 占位页文案（路径 -> 说明） */
export const adminPlaceholderMeta: Record<
  string,
  {
    summary: string
    bullets: string[]
    apiNote: string
  }
> = {
  '/refunds': {
    summary: '处理退款单、差错争议与长款挂账。',
    bullets: ['退款审核与批量', '差错登记与调账', '与通道对账差异联动'],
    apiNote: '待接入：/v1/admin/refunds',
  },
  '/reconcile': {
    summary: '按通道/商户/日期生成对账文件，标记差异并驱动结算。',
    bullets: ['对账批次与状态机', '差异类型：金额/笔数/状态', '自动调账或人工复核'],
    apiNote: '待接入：对账任务与文件存储',
  },
  '/settlement': {
    summary: '商户结算周期、提现审核、打款与手续费清算。',
    bullets: ['结算单生成与确认', '提现风控与打款通道', 'T+N、D+1 等策略'],
    apiNote: '待接入：/v1/admin/settlement',
  },
  '/risk': {
    summary: '交易前中后风控：限额、黑名单、设备指纹、反洗钱报送等。',
    bullets: ['规则引擎与名单库', '实时评分与阻断', '监管报表导出'],
    apiNote: '待接入：风控引擎与名单服务',
  },
  '/notifications': {
    summary: '向商户推送系统公告、维护窗口与政策变更。',
    bullets: ['公告编辑与定向推送', '站内信与 Webhook 摘要', '已读状态'],
    apiNote: '待接入：通知中心',
  },
  '/system': {
    summary: '管理员账号、角色权限、菜单与系统参数。',
    bullets: ['RBAC、数据权限（商户维度）', '操作审计留痕', '参数：全局开关、密钥轮换策略'],
    apiNote: '待接入：/v1/admin/users、roles、settings',
  },
  '/ops': {
    summary: '网关、RPC、队列、数据库与依赖服务的健康与容量。',
    bullets: ['服务拓扑与 QPS', '错误率与慢查询', '日志检索与 Trace ID'],
    apiNote: '待接入：可观测性平台对接',
  },
}

export function pathBelongsToGroup(path: string, group: AdminMenuGroup): boolean {
  return group.children.some((c) => c.to === path)
}

export function findGroupKeyForPath(path: string): string | null {
  for (const e of adminMenu) {
    if (e.kind === 'group' && pathBelongsToGroup(path, e)) return e.key
  }
  return null
}
