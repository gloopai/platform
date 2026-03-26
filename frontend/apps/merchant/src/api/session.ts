import { merchantConsolePost } from '@/lib/http'
import { MERCHANT_API } from '@/api/endpoints'
import type { MerchantChangePasswordReq, MerchantChangePasswordResp } from '@/types/merchant.api'

export async function postMerchantLogout(): Promise<{ ok: boolean }> {
  return merchantConsolePost<{ ok: boolean }>(MERCHANT_API.logout)
}

export async function postMerchantChangePassword(payload: MerchantChangePasswordReq): Promise<MerchantChangePasswordResp> {
  return merchantConsolePost<MerchantChangePasswordResp>(MERCHANT_API.changePassword, payload)
}
