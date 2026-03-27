/** 商户代收授权；merchant_rate_bps 与订单 fee_rate_bps 同语义（万分比整数=round(百分数×100)）。 */
export type MerchantPayinGrant = {
  payin_product_id: number
  /** 对客代收比例费率（万分比整数） */
  merchant_rate_bps?: number | null
}

/** 商户代付授权；fee_mode 与订单 fee_mode 同枚举（1 仅比例 2 仅固定 3 固定+比例）。 */
export type MerchantPayoutGrant = {
  payout_product_id: number
  /** 对客代付比例部分（万分比整数），与订单 fee_rate_bps 同语义 */
  merchant_rate_bps?: number | null
  /** 计费模式：1/2/3，见 docs/payment-fee-naming.md */
  fee_mode: number
  /** 对客代付固定部分（分），与订单 fee_fixed_amount 同语义 */
  fee_fixed_amount: number
}

export type AdminMerchantInfo = {
  merchant_id: string
  app_id: string
  email: string
  app_secret: string
  status: number
  notify_url: string
  return_url: string
  ip_whitelist: string
  withdraw_usdt_address?: string
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
  withdraw_usdt_address: string
}

export function emptyMerchantForm(): MerchantForm {
  return {
    merchant_id: '',
    email: '',
    status: 1,
    notify_url: '',
    return_url: '',
    ip_whitelist: '',
    withdraw_usdt_address: '',
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
    withdraw_usdt_address: m.withdraw_usdt_address || '',
  }
}
