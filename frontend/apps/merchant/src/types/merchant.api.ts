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
  balance: number
  merchant_id: string
}

export type MerchantOrderItem = {
  order_no: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  /** 支付产品编码（微信/支付宝/mock 等），与开放平台 pay_type 一致 */
  pay_product_code: string
  /** 管理台配置的展示名；缺省时前端仅用 `pay_product_code` */
  pay_product_name?: string
  paid_amount: number
  upstream_trade_no: string
  created_at: number
}

export type MerchantOrdersListResp = {
  orders: MerchantOrderItem[]
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
  pay_product_id: number
  pay_product_code: string
  pay_product_name?: string
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
