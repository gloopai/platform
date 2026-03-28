/**
 * 与网关 GET/POST /v1/admin/channels 一致。
 * 费率字段为平台相对通道(PSP)成本；命名前缀 channel_ 与商户侧对客字段区分。
 */
export type AdminChannel = {
  id: number
  name: string
  payin_type: string
  /** 通道对接自由 JSON 文本（整段存库） */
  channel_config: string
  weight: number
  min_amount: number
  max_amount: number
  supports_payin: boolean
  supports_payout: boolean
  /** 通道代收成本 — 比例费率存万分比整数（= round(百分数×100)） */
  channel_payin_rate_bps: number
  /** 通道代付成本 — 同上 */
  channel_payout_rate_bps: number
  /** 通道代付计费模式：1/2/3，与 fee_mode 枚举一致 */
  channel_payout_fee_mode: number
  /** 通道代付固定手续费（分） */
  channel_payout_fixed_fee: number
  enabled: boolean
  fuse_enabled: boolean
}

export function emptyChannelForm(): AdminChannel {
  return {
    id: 0,
    name: '',
    payin_type: '',
    channel_config: '',
    weight: 100,
    min_amount: 0,
    max_amount: 0,
    supports_payin: true,
    supports_payout: false,
    channel_payin_rate_bps: 0,
    channel_payout_rate_bps: 0,
    channel_payout_fee_mode: 1,
    channel_payout_fixed_fee: 0,
    enabled: true,
    fuse_enabled: false,
  }
}
