import { onScopeDispose, ref } from 'vue'

export type UiToastVariant = 'success' | 'error' | 'info'

export type UiToastItem = {
  id: number
  message: string
  variant: UiToastVariant
}

type UiToastStore = {
  toasts: UiToastItem[]
  nextId: number
  listeners: Set<(items: UiToastItem[]) => void>
  eventsBound: boolean
  domContainer: HTMLElement | null
  domItems: Map<number, HTMLElement>
}

type ToastShowEventDetail = UiToastItem
type ToastDismissEventDetail = { id: number }

const TOAST_SHOW_EVENT = '__gloop_shared_ui_toast_show__'
const TOAST_DISMISS_EVENT = '__gloop_shared_ui_toast_dismiss__'

const GLOBAL_TOAST_STORE_KEY = '__gloop_shared_ui_toast_store__'
const globalScope = globalThis as typeof globalThis & {
  [GLOBAL_TOAST_STORE_KEY]?: UiToastStore
}

if (!globalScope[GLOBAL_TOAST_STORE_KEY]) {
  globalScope[GLOBAL_TOAST_STORE_KEY] = {
    toasts: [],
    nextId: 1,
    listeners: new Set(),
    eventsBound: false,
    domContainer: null,
    domItems: new Map(),
  }
}

const store = globalScope[GLOBAL_TOAST_STORE_KEY]!

function bindWindowEvents() {
  if (store.eventsBound || typeof window === 'undefined') return

  window.addEventListener(TOAST_SHOW_EVENT, (ev: Event) => {
    const ce = ev as CustomEvent<ToastShowEventDetail>
    const item = ce.detail
    if (!item || typeof item.id !== 'number') return
    if (!store.toasts.some((t) => t.id === item.id)) {
      store.toasts = [...store.toasts, item]
      publish()
    }
  })

  window.addEventListener(TOAST_DISMISS_EVENT, (ev: Event) => {
    const ce = ev as CustomEvent<ToastDismissEventDetail>
    const id = ce.detail?.id
    if (typeof id !== 'number') return
    const next = store.toasts.filter((t) => t.id !== id)
    if (next.length !== store.toasts.length) {
      store.toasts = next
      publish()
    }
  })

  store.eventsBound = true
}

function publish() {
  const snapshot = store.toasts
  for (const notify of store.listeners) notify(snapshot)
}

function ensureDomContainer(): HTMLElement | null {
  if (typeof document === 'undefined') return null
  if (store.domContainer && document.body.contains(store.domContainer)) return store.domContainer
  const root = document.createElement('div')
  root.setAttribute('data-gloop-toast-root', '1')
  Object.assign(root.style, {
    position: 'fixed',
    top: '16px',
    left: '50%',
    transform: 'translateX(-50%)',
    width: 'min(92vw, 420px)',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    zIndex: '2147483600',
    pointerEvents: 'none',
  } satisfies Partial<CSSStyleDeclaration>)
  document.body.appendChild(root)
  store.domContainer = root
  return root
}

function domTheme(variant: UiToastVariant): { bg: string; border: string; color: string } {
  if (variant === 'error') return { bg: '#fff1f2', border: '#fecdd3', color: '#9f1239' }
  if (variant === 'info') return { bg: '#ffffff', border: '#e2e8f0', color: '#1f2937' }
  return { bg: '#ecfdf5', border: '#a7f3d0', color: '#065f46' }
}

function domIcon(variant: UiToastVariant): string {
  if (variant === 'error') return '!'
  if (variant === 'info') return 'i'
  return '✓'
}

function renderDomToast(item: UiToastItem, duration: number, onDismiss: (id: number) => void) {
  const root = ensureDomContainer()
  if (!root) return

  const el = document.createElement('div')
  const theme = domTheme(item.variant)
  el.setAttribute('data-toast-id', String(item.id))
  Object.assign(el.style, {
    pointerEvents: 'auto',
    borderRadius: '12px',
    border: `1px solid ${theme.border}`,
    background: theme.bg,
    color: theme.color,
    boxShadow: '0 10px 24px -12px rgba(15,23,42,0.35)',
    padding: '10px 12px',
    fontSize: '14px',
    fontWeight: '600',
    lineHeight: '1.4',
    display: 'flex',
    alignItems: 'flex-start',
    gap: '8px',
  } satisfies Partial<CSSStyleDeclaration>)

  const icon = document.createElement('div')
  icon.textContent = domIcon(item.variant)
  Object.assign(icon.style, {
    width: '18px',
    height: '18px',
    borderRadius: '9999px',
    border: `1px solid ${theme.border}`,
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '12px',
    fontWeight: '700',
    flexShrink: '0',
    marginTop: '1px',
  } satisfies Partial<CSSStyleDeclaration>)
  el.appendChild(icon)

  const msg = document.createElement('div')
  msg.textContent = item.message
  msg.style.flex = '1'
  el.appendChild(msg)

  const closeBtn = document.createElement('button')
  closeBtn.type = 'button'
  closeBtn.textContent = 'x'
  Object.assign(closeBtn.style, {
    border: 'none',
    background: 'transparent',
    color: '#64748b',
    cursor: 'pointer',
    fontSize: '13px',
    lineHeight: '1',
    marginTop: '1px',
  } satisfies Partial<CSSStyleDeclaration>)
  closeBtn.onclick = () => onDismiss(item.id)
  el.appendChild(closeBtn)

  root.appendChild(el)
  store.domItems.set(item.id, el)

  globalThis.setTimeout(() => {
    const current = store.domItems.get(item.id)
    if (current) {
      current.remove()
      store.domItems.delete(item.id)
    }
  }, duration + 80)
}

function removeDomToast(id: number) {
  const current = store.domItems.get(id)
  if (!current) return
  current.remove()
  store.domItems.delete(id)
}

function broadcastShow(item: UiToastItem) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent<ToastShowEventDetail>(TOAST_SHOW_EVENT, { detail: item }))
}

function broadcastDismiss(id: number) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent<ToastDismissEventDetail>(TOAST_DISMISS_EVENT, { detail: { id } }))
}

export function useUiToast() {
  bindWindowEvents()
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

  function dismiss(id: number, opts?: { silent?: boolean }) {
    store.toasts = store.toasts.filter((t) => t.id !== id)
    publish()
    removeDomToast(id)
    if (!opts?.silent) broadcastDismiss(id)
  }

  function show(
    message: string,
    opts?: { variant?: UiToastVariant; duration?: number },
  ): number {
    const id = store.nextId++
    const variant = opts?.variant ?? 'success'
    const duration = opts?.duration ?? 3200
    const item = { id, message, variant }
    store.toasts = [...store.toasts, item]
    publish()
    broadcastShow(item)
    renderDomToast(item, duration, dismiss)
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

  return { toasts, show, dismiss, success, error, info }
}
