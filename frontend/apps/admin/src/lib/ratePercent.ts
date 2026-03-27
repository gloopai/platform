/**
 * 比例费率：API/DB 存万分比整数 `bps`，与百分数换算为 `bps = round(percent × 100)`，
 * 与后端 `fee = amount * bps / 10000`（金额单位：分）一致。
 */

export const RATE_BPS_DIVISOR = 10_000

export function bpsToPercent(bps: number): number {
  return bps / 100
}

/** 表单输入：百分数 → 万分比整数 */
export function percentToBps(percent: number): number {
  if (!Number.isFinite(percent) || percent < 0) return 0
  return Math.round(percent * 100)
}

/** 只读展示，如订单/产品列表 */
export function formatPercentFromBps(bps: number, fractionDigits = 2): string {
  if (!Number.isFinite(bps)) return '—'
  return `${(bps / 100).toFixed(fractionDigits)}%`
}

/** `<input type="number">` 的受控值，避免浮点噪声 */
export function bpsToPercentInputValue(bps: number): number {
  return Math.round(bps) / 100
}
