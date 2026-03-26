import { computed, ref } from 'vue'

export type UiDialogVariant = 'default' | 'danger'

export type UiDialogOptions = {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: UiDialogVariant
  hideCancel?: boolean
}

type DialogTask = {
  id: number
  options: Required<UiDialogOptions>
  resolve: (confirmed: boolean) => void
}

const queue = ref<DialogTask[]>([])
let nextId = 1

function normalizeOptions(opts: UiDialogOptions): Required<UiDialogOptions> {
  return {
    title: opts.title ?? '提示',
    message: opts.message,
    confirmText: opts.confirmText ?? '确定',
    cancelText: opts.cancelText ?? '取消',
    variant: opts.variant ?? 'default',
    hideCancel: opts.hideCancel ?? false,
  }
}

export function useUiDialog() {
  const current = computed(() => queue.value[0] ?? null)

  function open(options: UiDialogOptions): Promise<boolean> {
    return new Promise((resolve) => {
      queue.value.push({
        id: nextId++,
        options: normalizeOptions(options),
        resolve,
      })
    })
  }

  function closeCurrent(confirmed: boolean) {
    const task = queue.value[0]
    if (!task) return
    task.resolve(confirmed)
    queue.value = queue.value.slice(1)
  }

  function confirm(message: string, title = '请确认') {
    return open({ title, message, variant: 'default', hideCancel: false })
  }

  function danger(message: string, title = '危险操作') {
    return open({
      title,
      message,
      variant: 'danger',
      confirmText: '确认',
      cancelText: '取消',
      hideCancel: false,
    })
  }

  function alert(message: string, title = '提示') {
    return open({
      title,
      message,
      hideCancel: true,
      confirmText: '我知道了',
    })
  }

  return { current, open, closeCurrent, confirm, danger, alert }
}
