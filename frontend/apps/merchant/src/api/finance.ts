import { merchantConsoleGet, merchantConsolePost } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantFundLogsResp, MerchantTransferPayinToPayoutResp } from '@/types/merchant.api'

export async function fetchMerchantFundLogs(limit = 50): Promise<MerchantFundLogsResp> {
  return merchantConsoleGet<MerchantFundLogsResp>(MERCHANT_API.fundLogs, { limit })
}

export async function transferPayinToPayout(amount: number, reason = 'MERCHANT_MANUAL_TRANSFER'): Promise<MerchantTransferPayinToPayoutResp> {
  return merchantConsolePost<MerchantTransferPayinToPayoutResp>(MERCHANT_API.transferPayinToPayout, { amount, reason })
}
