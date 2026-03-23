/** 金额：后端分为单位 */

export function formatCentsWithCurrency(amount: number, currency = 'CNY'): string {
  return `${(amount / 100).toFixed(2)} ${currency}`
}

export function formatYuanLabel(cents: number): string {
  return `¥ ${(cents / 100).toFixed(2)}`
}

export function formatUnixSeconds(ts: number): string {
  const d = new Date(ts * 1000)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString('zh-CN', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
