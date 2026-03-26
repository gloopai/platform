/**
 * 商户端相关 URL 单点维护，改网关路径时只改此处。
 */
export const MERCHANT_API = {
  login: '/v1/merchant/login',
  logout: '/v1/merchant/logout',
  summary: '/v1/merchant/summary',
  config: '/v1/merchant/config',
  statsByProduct: '/v1/merchant/stats/by_product',
  products: '/v1/merchant/products',
  payinOrders: '/v1/merchant/payin_orders',
  payoutOrders: '/v1/merchant/payout_orders',
  orderDetail: '/v1/merchant/order/detail',
  retryNotify: '/v1/merchant/order/retry_notify',
  changePassword: '/v1/merchant/password/change',
  fundLogs: '/v1/merchant/fund_logs',
  transferPayinToPayout: '/v1/merchant/balance/transfer_payin_to_payout',
} as const

/** 开放网关（非 X-Merchant-Token） */
export const OPEN_API = {
  payinOrder: '/v1/payin/order',
  payoutOrder: '/v1/payout/order',
  queryPayinOrder: '/v1/payin/query',
  queryPayoutOrder: '/v1/payout/query',
  callbackNotify: '/v1/callback/notify',
} as const
