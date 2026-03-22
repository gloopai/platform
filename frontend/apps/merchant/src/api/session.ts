import { merchantConsolePost } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'

export async function postMerchantLogout(): Promise<{ ok: boolean }> {
  return merchantConsolePost<{ ok: boolean }>(MERCHANT_API.logout)
}
