import { merchantConsoleGet, merchantConsolePost } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantFundLogsResp, MerchantTransferCollectToPayoutResp } from '@/types/merchant.api'

export async function fetchMerchantFundLogs(limit = 50): Promise<MerchantFundLogsResp> {
  return merchantConsoleGet<MerchantFundLogsResp>(MERCHANT_API.fundLogs, { limit })
}

export async function transferCollectToPayout(amount: number, reason = 'MERCHANT_MANUAL_TRANSFER'): Promise<MerchantTransferCollectToPayoutResp> {
  return merchantConsolePost<MerchantTransferCollectToPayoutResp>(MERCHANT_API.transferCollectToPayout, { amount, reason })
}
