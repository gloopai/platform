export type StatsTotals = {
  order_count: number
  paid_amount: number
  paid_count: number
  failed_count: number
  pending_count: number
  closed_count: number
  conversion_rate_pct: number
  terminal_success_rate_pct: number
}

export type StatsProductRow = {
  product_code: string
  product_name: string
  order_count: number
  paid_amount: number
  paid_count: number
  failed_count: number
  conversion_rate_pct: number
  terminal_success_rate_pct: number
}

export type StatsChannelRow = {
  channel_id: number
  channel_name: string
  order_count: number
  paid_amount: number
  paid_count: number
  failed_count: number
  conversion_rate_pct: number
  terminal_success_rate_pct: number
}

export type StatsOverview = {
  range: string
  totals: StatsTotals
  by_payin_product: StatsProductRow[]
  by_channel: StatsChannelRow[]
  enabled_channels: number
  fused_channels: number
}
