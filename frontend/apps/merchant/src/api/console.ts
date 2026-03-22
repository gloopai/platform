import { merchantConsoleGet } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantSummary } from '@/types/merchant.api'

export async function fetchMerchantSummary(): Promise<MerchantSummary> {
  return merchantConsoleGet<MerchantSummary>(MERCHANT_API.summary)
}
