import { useUiToast, type UiToastItem, type UiToastVariant } from './useUiToast'

export type AdminToastVariant = UiToastVariant
export type AdminToastItem = UiToastItem

/**
 * 兼容旧调用：内部已切到公共 useUiToast。
 */
export function useAdminToast() {
  return useUiToast()
}
