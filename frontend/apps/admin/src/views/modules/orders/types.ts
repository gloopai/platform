export type AdminOrderRow = {
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  pay_product_id: number
  pay_product_code: string
  paid_amount: number
  fee_mode: number
  fee_rate_bps: number
  fee_fixed_amount: number
  fee_amount: number
  net_amount: number
  upstream_trade_no: string
  created_at: number
}

export type AdminOrdersResp = {
  orders: AdminOrderRow[]
}
