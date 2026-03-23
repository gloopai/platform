export type PayinProduct = {
  id: number
  code: string
  name: string
  sort_order: number
  enabled: boolean
}

export type PayinProductBinding = {
  id: number
  payin_product_id?: number
  payout_product_id?: number
  channel_id: number
  channel_name: string
  weight: number
  enabled: boolean
}

export type PayinProductChannelOption = {
  id: number
  name: string
  pay_type: string
  supports_payin?: boolean
  supports_payout?: boolean
}
