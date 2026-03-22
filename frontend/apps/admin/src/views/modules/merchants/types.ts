export type AdminMerchantInfo = {
  merchant_id: string
  api_secret: string
  status: number
  rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
  balance: number
  pay_product_ids?: number[]
}

export type PayProductRow = { id: number; code: string; name: string }

export type MerchantForm = {
  merchant_id: string
  api_secret: string
  status: number
  rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
}

export function emptyMerchantForm(): MerchantForm {
  return {
    merchant_id: '',
    api_secret: '',
    status: 1,
    rate_bps: 0,
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
    rate_bps: m.rate_bps,
    notify_url: m.notify_url,
    return_url: m.return_url,
    ip_whitelist: m.ip_whitelist,
  }
}
