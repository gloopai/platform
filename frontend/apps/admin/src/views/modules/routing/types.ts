export type RoutingSummary = {
  algorithm_key: string
  algorithm_label: string
  enabled_pay_products: number
  enabled_payout_products: number
  enabled_channels: number
  active_bindings: number
  active_payout_bindings: number
  merchants_with_collect_whitelist: number
  merchants_with_payout_whitelist: number
  fused_channels: number
}
