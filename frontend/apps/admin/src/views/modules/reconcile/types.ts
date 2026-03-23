import type { StatsChannelRow, StatsProductRow, StatsTotals } from '../stats/types'

/** 与 GET /v1/admin/reconcile/day 一致 */
export type ReconcileDayOverview = {
  date: string
  totals: StatsTotals
  by_pay_product: StatsProductRow[]
  by_channel: StatsChannelRow[]
}
