/** 管理台订单行；手续费字段为下单时快照，与商户授权字段同语义（命名见 docs/payment-fee-naming.md）。 */
export type AdminOrderRow = {
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  payin_product_id: number
  payin_product_code: string
  paid_amount: number
  /** 计费模式 1/2/3 */
  fee_mode: number
  /** 比例部分：万分比整数（= round(百分数×100)），与 merchant_rate_bps 同语义 */
  fee_rate_bps: number
  /** 固定部分（分）；与商户代付授权 payout grant 的 fee_fixed_amount 同语义 */
  fee_fixed_amount: number
  fee_amount: number
  net_amount: number
  upstream_trade_no: string
  created_at: number
}

export type AdminOrdersResp = {
  orders: AdminOrderRow[]
  total?: number
}
