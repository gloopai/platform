/** 与网关 GET/POST /v1/admin/channels 字段一致 */
export type AdminChannel = {
  id: number
  name: string
  pay_type: string
  gateway_url: string
  upstream_merchant_no: string
  rsa_private_key: string
  sign_secret: string
  weight: number
  min_amount: number
  max_amount: number
  supports_collect: boolean
  supports_payout: boolean
  upstream_collect_cost_bps: number
  upstream_payout_cost_bps: number
  enabled: boolean
  fuse_enabled: boolean
}

export function emptyChannelForm(): AdminChannel {
  return {
    id: 0,
    name: '',
    pay_type: '',
    gateway_url: '',
    upstream_merchant_no: '',
    rsa_private_key: '',
    sign_secret: '',
    weight: 100,
    min_amount: 0,
    max_amount: 0,
    supports_collect: true,
    supports_payout: false,
    upstream_collect_cost_bps: 0,
    upstream_payout_cost_bps: 0,
    enabled: true,
    fuse_enabled: false,
  }
}
