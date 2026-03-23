/**
 * 商户端相关 URL 单点维护，改网关路径时只改此处。
 */
export const MERCHANT_API = {
  login: '/v1/merchant/login',
  logout: '/v1/merchant/logout',
  summary: '/v1/merchant/summary',
  statsByProduct: '/v1/merchant/stats/by_product',
  payOrders: '/v1/merchant/pay_orders',
  payoutOrders: '/v1/merchant/payout_orders',
  orderDetail: '/v1/merchant/order/detail',
  retryNotify: '/v1/merchant/order/retry_notify',
  fundLogs: '/v1/merchant/fund_logs',
} as const

/** 开放网关（非 X-Merchant-Token） */
export const OPEN_API = {
  payOrder: '/v1/pay/order',
  queryOrder: '/v1/pay/query',
  callbackNotify: '/v1/callback/notify',
} as const
