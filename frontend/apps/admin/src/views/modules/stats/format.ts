/** 订单金额：与后端一致，分为单位 */
export function formatYuan(cents: number): string {
  if (!Number.isFinite(cents)) return '—'
  return (cents / 100).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function formatPct(n: number): string {
  if (!Number.isFinite(n)) return '—'
  return `${n.toFixed(2)}%`
}
