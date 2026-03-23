/** 金额：后端分为单位 */
import { merchantDisplaySettings } from '@/lib/displaySettings'

export function formatCentsWithCurrency(amount: number, currency = 'CNY'): string {
  const symbol = merchantDisplaySettings.value.currency_symbol || '¥'
  const code = merchantDisplaySettings.value.currency_code || currency
  return `${symbol} ${(amount / 100).toFixed(2)} ${code}`
}

export function formatYuanLabel(cents: number): string {
  const symbol = merchantDisplaySettings.value.currency_symbol || '¥'
  return `${symbol} ${(cents / 100).toFixed(2)}`
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
