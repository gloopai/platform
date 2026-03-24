/**
 * 商户控制台 JSON 模型（与 gateway /v1/merchant/* 响应对齐，便于联调与重构）。
 */

export type MerchantLoginResponse = {
  token: string
  expires_at: number
  merchant_id: string
}

export type MerchantSummary = {
  today_amount: number
  today_count: number
  success_rate: number
  payin_balance: number
  available_balance: number
  merchant_id: string
  api_secret: string
  notify_url: string
  ip_whitelist: string
}

export type MerchantUpdateConfigReq = {
  notify_url: string
  ip_whitelist: string
}

export type MerchantUpdateConfigResp = {
  merchant_id: string
  api_secret: string
  notify_url: string
  ip_whitelist: string
}

export type MerchantProductStatsItem = {
  payin_product_code: string
  payin_product_name: string
  order_count: number
  paid_amount: number
  paid_count: number
  failed_count: number
  success_rate_pct: number
}

export type MerchantProductStatsResp = {
  date: string
  merchant_id: string
  items: MerchantProductStatsItem[]
}

export type MerchantOrderItem = {
  order_no: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  /** 支付产品编码（微信/支付宝/mock 等），与开放平台 payin_type 一致 */
  payin_product_code: string
  /** 管理台配置的展示名；缺省时前端仅用 `payin_product_code` */
  payin_product_name?: string
  paid_amount: number
  fee_mode: number
  fee_rate_bps: number
  fee_fixed_amount: number
  fee_amount: number
  net_amount: number
  upstream_trade_no: string
  created_at: number
}

export type MerchantOrdersListResp = {
  orders: MerchantOrderItem[]
  /** 满足筛选条件的总条数；旧网关未返回时可缺省 */
  total?: number
}

export type MerchantNotifyLogItem = {
  id: number
  notify_url: string
  attempt: number
  http_status: number
  response_body: string
  error_msg: string
  created_at: number
}

export type MerchantOrderDetail = {
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  payin_product_id: number
  payin_product_code: string
  payin_product_name?: string
  fee_mode: number
  fee_rate_bps: number
  fee_fixed_amount: number
  fee_amount: number
  net_amount: number
  return_url: string
  notify_url: string
  upstream_trade_no: string
}

export type MerchantOrderDetailResp = {
  order: MerchantOrderDetail
  logs: MerchantNotifyLogItem[]
}

export type MerchantFundLogItem = {
  id: number
  order_no: string
  change_type: string
  amount: number
  balance_before: number
  balance_after: number
  reason: string
  created_at: number
}

export type MerchantFundLogsResp = {
  logs: MerchantFundLogItem[]
}

export type MerchantTransferPayinToPayoutResp = {
  ok: boolean
  payin_balance: number
  available_balance: number
}
