import { ref } from 'vue'

export type UiToastVariant = 'success' | 'error' | 'info'

export type UiToastItem = {
  id: number
  message: string
  variant: UiToastVariant
}

type UiToastStore = {
  toasts: ReturnType<typeof ref<UiToastItem[]>>
  nextId: number
}

const GLOBAL_TOAST_STORE_KEY = '__gloop_shared_ui_toast_store__'
const globalScope = globalThis as typeof globalThis & {
  [GLOBAL_TOAST_STORE_KEY]?: UiToastStore
}

if (!globalScope[GLOBAL_TOAST_STORE_KEY]) {
  globalScope[GLOBAL_TOAST_STORE_KEY] = {
    toasts: ref<UiToastItem[]>([]),
    nextId: 1,
  }
}

const store = globalScope[GLOBAL_TOAST_STORE_KEY]!

export function useUiToast() {
  function dismiss(id: number) {
    store.toasts.value = store.toasts.value.filter((t) => t.id !== id)
  }

  function show(
    message: string,
    opts?: { variant?: UiToastVariant; duration?: number },
  ): number {
    const id = store.nextId++
    const variant = opts?.variant ?? 'success'
    const duration = opts?.duration ?? 3200
    store.toasts.value = [...store.toasts.value, { id, message, variant }]
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

  return { toasts: store.toasts, show, dismiss, success, error, info }
}
