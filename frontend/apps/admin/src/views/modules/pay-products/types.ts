export type PayProduct = {
  id: number
  code: string
  name: string
  sort_order: number
  enabled: boolean
}

export type PayProductBinding = {
  id: number
  pay_product_id: number
  channel_id: number
  channel_name: string
  weight: number
  enabled: boolean
}

export type PayProductChannelOption = {
  id: number
  name: string
  pay_type: string
}
