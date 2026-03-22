export type MerchantCollectGrant = {
  pay_product_id: number
  merchant_rate_bps?: number | null
}

export type MerchantPayoutGrant = {
  payout_product_id: number
  merchant_rate_bps?: number | null
}

export type AdminMerchantInfo = {
  merchant_id: string
  api_secret: string
  status: number
  default_collect_rate_bps: number
  default_payout_rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
  balance: number
  pay_product_ids?: number[]
  payout_product_ids?: number[]
  collect_grants?: MerchantCollectGrant[]
  payout_grants?: MerchantPayoutGrant[]
}

export type PayProductRow = { id: number; code: string; name: string }

export type MerchantForm = {
  merchant_id: string
  api_secret: string
  status: number
  default_collect_rate_bps: number
  default_payout_rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
}

export function emptyMerchantForm(): MerchantForm {
  return {
    merchant_id: '',
    api_secret: '',
    status: 1,
    default_collect_rate_bps: 0,
    default_payout_rate_bps: 0,
    notify_url: '',
    return_url: '',
    ip_whitelist: '',
  }
}

export function merchantToForm(m: AdminMerchantInfo): MerchantForm {
  return {
    merchant_id: m.merchant_id,
    api_secret: m.api_secret,
    status: m.status,
    default_collect_rate_bps: m.default_collect_rate_bps,
    default_payout_rate_bps: m.default_payout_rate_bps,
    notify_url: m.notify_url,
    return_url: m.return_url,
    ip_whitelist: m.ip_whitelist,
  }
}
