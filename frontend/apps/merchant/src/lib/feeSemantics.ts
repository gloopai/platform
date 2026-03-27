/**
 * 与管理台 `apps/admin/src/lib/feeSemantics.ts` 保持一致；商户端展示/订单列表复用。
 */

export const FeeMode = {
  RateOnly: 1,
  FixedOnly: 2,
  FixedPlusRate: 3,
} as const

export function feeModeOptionLabel(mode: number): string {
  if (mode === FeeMode.FixedOnly) return '仅固定'
  if (mode === FeeMode.FixedPlusRate) return '固定+比例'
  return '仅比例'
}

export function feeModeDescription(mode: number): string {
  if (mode === FeeMode.FixedOnly) return '仅固定（分）'
  if (mode === FeeMode.FixedPlusRate) return '固定（分）+ 比例（%）'
  return '仅比例（%）'
}

export const LABEL_RATE_BPS = '比例费率（%）'
export const LABEL_FIXED_FEN = '固定手续费（分）'
