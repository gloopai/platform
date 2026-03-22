import { merchantConsoleGet } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantFundLogsResp } from '@/types/merchant.api'

export async function fetchMerchantFundLogs(limit = 50): Promise<MerchantFundLogsResp> {
  return merchantConsoleGet<MerchantFundLogsResp>(MERCHANT_API.fundLogs, { limit })
}
