import { ref } from 'vue'

export type UiToastVariant = 'success' | 'error' | 'info'

export type UiToastItem = {
  id: number
  message: string
  variant: UiToastVariant
}

const toasts = ref<UiToastItem[]>([])
let nextId = 1

export function useUiToast() {
  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function show(
    message: string,
    opts?: { variant?: UiToastVariant; duration?: number },
  ): number {
    const id = nextId++
    const variant = opts?.variant ?? 'success'
    const duration = opts?.duration ?? 3200
    toasts.value = [...toasts.value, { id, message, variant }]
    window.setTimeout(() => dismiss(id), duration)
    return id
  }

  function success(message = '操作成功') {
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
