import { onScopeDispose, ref } from 'vue'

export type UiToastVariant = 'success' | 'error' | 'info' | 'warning'

export type UiToastItem = {
  id: number
  message: string
  variant: UiToastVariant
}

type UiToastStore = {
  toasts: UiToastItem[]
  nextId: number
  listeners: Set<(items: UiToastItem[]) => void>
}

const TOAST_STORE_KEY = '__admin_ui_toast_store__'
const root = globalThis as typeof globalThis & { [TOAST_STORE_KEY]?: UiToastStore }

if (!root[TOAST_STORE_KEY]) {
  root[TOAST_STORE_KEY] = {
    toasts: [],
    nextId: 1,
    listeners: new Set(),
  }
}

const store = root[TOAST_STORE_KEY]!

function publish() {
  const snapshot = store.toasts
  for (const notify of store.listeners) notify(snapshot)
}

export function useUiToast() {
  const toasts = ref<UiToastItem[]>(store.toasts)

  function subscribe(notify: (items: UiToastItem[]) => void) {
    store.listeners.add(notify)
    notify(store.toasts)
    return () => {
      store.listeners.delete(notify)
    }
  }

  const unsubscribe = subscribe((items) => {
    toasts.value = items
  })
  onScopeDispose(unsubscribe)

  function dismiss(id: number) {
    store.toasts = store.toasts.filter((t) => t.id !== id)
    publish()
  }

  function show(message: string, opts?: { variant?: UiToastVariant; duration?: number }): number {
    const id = store.nextId++
    const variant = opts?.variant ?? 'success'
    const duration = opts?.duration ?? 3200
    const item = { id, message, variant }
    store.toasts = [...store.toasts, item]
    publish()
    globalThis.setTimeout(() => dismiss(id), duration)
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

  function warning(message: string, duration = 3600) {
    return show(message, { variant: 'warning', duration })
  }

  return { toasts, show, dismiss, success, error, info, warning }
}
