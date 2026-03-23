import { merchantConsoleGet } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantProductStatsResp, MerchantSummary } from '@/types/merchant.api'

export async function fetchMerchantSummary(): Promise<MerchantSummary> {
  return merchantConsoleGet<MerchantSummary>(MERCHANT_API.summary)
}

export async function fetchMerchantStatsByProduct(): Promise<MerchantProductStatsResp> {
  return merchantConsoleGet<MerchantProductStatsResp>(MERCHANT_API.statsByProduct)
}
