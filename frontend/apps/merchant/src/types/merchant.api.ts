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
