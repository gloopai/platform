import { merchantConsoleGet, merchantConsolePost } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantOrderDetailResp, MerchantOrdersListResp } from '@/types/merchant.api'

export type MerchantOrderListQuery = {
  order_no?: string
  status?: string
  limit?: number
}

export async function fetchMerchantOrders(params: MerchantOrderListQuery): Promise<MerchantOrdersListResp> {
  return merchantConsoleGet<MerchantOrdersListResp>(MERCHANT_API.orders, params)
}

export async function fetchMerchantOrderDetail(orderNo: string): Promise<MerchantOrderDetailResp> {
  return merchantConsoleGet<MerchantOrderDetailResp>(MERCHANT_API.orderDetail, { order_no: orderNo })
}

export async function postRetryMerchantNotify(orderNo: string): Promise<{ ok: boolean }> {
  return merchantConsolePost<{ ok: boolean }>(MERCHANT_API.retryNotify, { order_no: orderNo })
}
