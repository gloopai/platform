export type MerchantPayinGrant = {
  payin_product_id: number
  merchant_rate_bps?: number | null
}

export type MerchantPayoutGrant = {
  payout_product_id: number
  merchant_rate_bps?: number | null
  fee_mode: number
  fee_fixed_amount: number
}

export type AdminMerchantInfo = {
  merchant_id: string
  app_id: string
  email: string
  app_secret: string
  status: number
  default_payin_rate_bps: number
  default_payout_rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
  payin_balance: number
  available_balance: number
  payin_product_ids?: number[]
  payout_product_ids?: number[]
  payin_grants?: MerchantPayinGrant[]
  payout_grants?: MerchantPayoutGrant[]
}

export type ProductRow = { id: number; code: string; name: string }

export type MerchantForm = {
  merchant_id: string
  email: string
  status: number
  notify_url: string
  return_url: string
  ip_whitelist: string
}

export function emptyMerchantForm(): MerchantForm {
  return {
    merchant_id: '',
    email: '',
    status: 1,
    notify_url: '',
    return_url: '',
    ip_whitelist: '',
  }
}

export function merchantToForm(m: AdminMerchantInfo): MerchantForm {
  return {
    merchant_id: m.merchant_id,
    email: m.email || '',
    status: m.status,
    notify_url: m.notify_url,
    return_url: m.return_url,
    ip_whitelist: m.ip_whitelist,
  }
}
