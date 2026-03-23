import { ref } from 'vue'

export type AdminToastVariant = 'success' | 'error' | 'info'

export type AdminToastItem = {
  id: number
  message: string
  variant: AdminToastVariant
}

const toasts = ref<AdminToastItem[]>([])
let nextId = 1

/**
 * 全局 Toast（单例）。任意组件内调用 `useAdminToast()` 共享同一条提示队列。
 * 需在 `AdminLayout` 中挂载 `<AdminToastHost />` 才会显示。
 */
export function useAdminToast() {
  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  /**
   * @param message 提示文案
   * @param opts.variant 默认 success；error 用于失败提示
   * @param opts.duration 毫秒，默认 3200
   * @returns 本条 id，可配合 dismiss 提前关闭
   */
  function show(
    message: string,
    opts?: { variant?: AdminToastVariant; duration?: number },
  ): number {
    const id = nextId++
    const variant = opts?.variant ?? 'success'
    const duration = opts?.duration ?? 3200
    toasts.value = [...toasts.value, { id, message, variant }]
    window.setTimeout(() => dismiss(id), duration)
    return id
  }

  /** 与 show 等价，语义化成功保存 / 编辑完成 */
  function success(message = '保存成功') {
    return show(message, { variant: 'success' })
  }

  function error(message: string, duration = 4200) {
    return show(message, { variant: 'error', duration })
  }

  function info(message: string, duration = 3200) {
    return show(message, { variant: 'info', duration })
  }

  return { toasts, show, dismiss, success, error, info }
}
