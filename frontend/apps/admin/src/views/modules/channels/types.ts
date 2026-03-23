/** 与网关 GET/POST /v1/admin/channels 字段一致 */
export type AdminChannel = {
  id: number
  name: string
  payin_type: string
  gateway_url: string
  upstream_merchant_no: string
  rsa_private_key: string
  sign_secret: string
  weight: number
  min_amount: number
  max_amount: number
  supports_payin: boolean
  supports_payout: boolean
  upstream_payin_rate_bps: number
  upstream_payout_rate_bps: number
  upstream_payout_fee_mode: number
  upstream_payout_fixed_fee: number
  enabled: boolean
  fuse_enabled: boolean
}

export function emptyChannelForm(): AdminChannel {
  return {
    id: 0,
    name: '',
    payin_type: '',
    gateway_url: '',
    upstream_merchant_no: '',
    rsa_private_key: '',
    sign_secret: '',
    weight: 100,
    min_amount: 0,
    max_amount: 0,
    supports_payin: true,
    supports_payout: false,
    upstream_payin_rate_bps: 0,
    upstream_payout_rate_bps: 0,
    upstream_payout_fee_mode: 1,
    upstream_payout_fixed_fee: 0,
    enabled: true,
    fuse_enabled: false,
  }
}
