/**
 * 全站费率相关：单位、计费模式文案、表单标签。
 * 与 DB / API 字段对应关系见仓库 docs/payment-fee-naming.md。
 */

/** 计费模式，与 fee_mode、channel_payout_fee_mode 取值一致。 */
export const FeeMode = {
  /** 仅按比例（展示为百分数，存万分比整数） */
  RateOnly: 1,
  /** 仅按「分」收固定手续费 */
  FixedOnly: 2,
  /** 固定（分）+ 比例（百分数 / 存万分比） */
  FixedPlusRate: 3,
} as const

export type FeeModeNumber = (typeof FeeMode)[keyof typeof FeeMode]

/** 表格、下拉框等简短标签（统一用语，无歧义） */
export function feeModeOptionLabel(mode: number): string {
  if (mode === FeeMode.FixedOnly) return '仅固定'
  if (mode === FeeMode.FixedPlusRate) return '固定+比例'
  return '仅比例'
}

/** 带单位的完整说明（帮助文案、tooltip） */
export function feeModeDescription(mode: number): string {
  if (mode === FeeMode.FixedOnly) return '仅固定（分）'
  if (mode === FeeMode.FixedPlusRate) return '固定（分）+ 比例（%）'
  return '仅比例（%）'
}

/** 管理台下拉选项（value 与后端一致） */
export const FEE_MODE_SELECT_OPTIONS: { value: number; label: string }[] = [
  { value: FeeMode.RateOnly, label: '仅比例（%）' },
  { value: FeeMode.FixedOnly, label: '仅固定（分）' },
  { value: FeeMode.FixedPlusRate, label: '固定（分）+ 比例（%）' },
]

// —— 比例费率：界面用百分数；API/DB 仍为 *_rate_bps（= round(百分数×100)）——
/** merchant_rate_bps、fee_rate_bps、channel_*_rate_bps */
export const LABEL_RATE_BPS = '比例费率（%）'

// —— 固定金额（单位：分）——
/** fee_fixed_amount、channel_payout_fixed_fee */
export const LABEL_FIXED_FEN = '固定手续费（分）'

export const LABEL_FEE_MODE = '计费模式'

// —— 通道成本 ——
export const LABEL_CHANNEL_PAYIN_RATE = '通道代收 — 比例费率（%）'
export const LABEL_CHANNEL_PAYOUT_RATE = '通道代付 — 比例费率（%）'
export const LABEL_CHANNEL_PAYOUT_FEE_MODE = '通道代付 — 计费模式'
export const LABEL_CHANNEL_PAYOUT_FIXED = '通道代付 — 固定手续费（分）'

// —— 商户授权：对客规则 ——
export const LABEL_MERCHANT_PAYIN_RATE = '对客代收 — 比例费率（%）'
export const LABEL_MERCHANT_PAYOUT_FEE_MODE = '对客代付 — 计费模式'
export const LABEL_MERCHANT_PAYOUT_RATE = '对客代付 — 比例费率（%）'
export const LABEL_MERCHANT_PAYOUT_FIXED = '对客代付 — 固定手续费（分）'

/** 产品绑定等场景一句话（不含「对客/上游」前缀时） */
export const BLURB_RATE_BPS_UNIT =
  '比例费率在界面按百分数（%）填写与展示；接口与库中仍为整数万分比（= round(百分数×100)）。固定金额单位均为分。手续费(分)=金额(分)×bps/10000。'
