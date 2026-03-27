/**
 * 与管理台 `apps/admin/src/lib/ratePercent.ts` 一致。
 * 比例费率：DB/API 为万分比整数，展示为百分数：`percent = bps / 100`。
 */

export const RATE_BPS_DIVISOR = 10_000

export function bpsToPercent(bps: number): number {
  return bps / 100
}

export function percentToBps(percent: number): number {
  if (!Number.isFinite(percent) || percent < 0) return 0
  return Math.round(percent * 100)
}

export function formatPercentFromBps(bps: number, fractionDigits = 2): string {
  if (!Number.isFinite(bps)) return '—'
  return `${(bps / 100).toFixed(fractionDigits)}%`
}

export function bpsToPercentInputValue(bps: number): number {
  return Math.round(bps) / 100
}
