/** 商户订单状态（与平台 order.status 数值约定一致） */

export function orderStatusLabel(status: number): string {
  if (status === 0) return '待支付'
  if (status === 1) return '成功'
  if (status === 2) return '失败'
  if (status === 3) return '已关闭'
  return `未知(${status})`
}

export function orderStatusBadgeClass(status: number): string {
  if (status === 1) return 'bg-slate-200/90 text-slate-800'
  if (status === 2) return 'bg-rose-100 text-rose-800'
  if (status === 3) return 'bg-slate-100 text-slate-600'
  return 'bg-amber-100 text-amber-900'
}
