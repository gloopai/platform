/** 金额：后端分为单位 */

export function formatCentsWithCurrency(amount: number, currency = 'CNY'): string {
  return `${(amount / 100).toFixed(2)} ${currency}`
}

export function formatYuanLabel(cents: number): string {
  return `¥ ${(cents / 100).toFixed(2)}`
}

export function formatUnixSeconds(ts: number): string {
  return new Date(ts * 1000).toLocaleString()
}
