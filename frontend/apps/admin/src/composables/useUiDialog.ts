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

type UiDialogStore = {
  queue: DialogTask[]
  nextId: number
  active: boolean
}

const DIALOG_STORE_KEY = '__admin_ui_dialog_store__'

function getStore(): UiDialogStore {
  const root = globalThis as Record<string, unknown>
  if (!root[DIALOG_STORE_KEY]) {
    root[DIALOG_STORE_KEY] = {
      queue: [],
      nextId: 1,
      active: false,
    } satisfies UiDialogStore
  }
  return root[DIALOG_STORE_KEY] as UiDialogStore
}

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

function pumpQueue(store: UiDialogStore) {
  if (store.active) return
  const task = store.queue.shift()
  if (!task) return
  store.active = true
  presentTask(task)
    .then((confirmed) => task.resolve(confirmed))
    .finally(() => {
      store.active = false
      pumpQueue(store)
    })
}

function presentTask(task: DialogTask): Promise<boolean> {
  if (typeof document === 'undefined') return Promise.resolve(true)
  return new Promise((resolve) => {
    try {
      const modal = document.createElement('dialog')
      modal.className = 'modal modal-open'

      const panel = document.createElement('div')
      panel.className = 'modal-box w-full max-w-md'

      const title = document.createElement('h3')
      title.className = 'text-lg font-semibold'
      title.textContent = task.options.title

      const message = document.createElement('p')
      message.className = 'py-3 whitespace-pre-wrap text-sm leading-relaxed opacity-80'
      message.textContent = task.options.message

      const footer = document.createElement('div')
      footer.className = 'modal-action'

      let finished = false
      const done = (ok: boolean) => {
        if (finished) return
        finished = true
        modal.removeEventListener('close', onClose)
        modal.remove()
        resolve(ok)
      }
      const onClose = () => done(false)
      modal.addEventListener('close', onClose)

      if (!task.options.hideCancel) {
        const cancelBtn = document.createElement('button')
        cancelBtn.type = 'button'
        cancelBtn.className = 'btn btn-sm btn-outline border-base-content text-base-content hover:bg-base-200'
        cancelBtn.textContent = task.options.cancelText
        cancelBtn.onclick = () => done(false)
        footer.appendChild(cancelBtn)
      }

      const confirmBtn = document.createElement('button')
      confirmBtn.type = 'button'
      confirmBtn.className =
        task.options.variant === 'danger'
          ? 'btn btn-sm border-0 bg-error text-error-content hover:bg-error/90'
          : 'btn btn-sm border-0 bg-black text-white hover:bg-zinc-800'
      confirmBtn.textContent = task.options.confirmText
      confirmBtn.onclick = () => done(true)
      footer.appendChild(confirmBtn)

      panel.appendChild(title)
      panel.appendChild(message)
      panel.appendChild(footer)
      modal.appendChild(panel)

      const backdropForm = document.createElement('form')
      backdropForm.method = 'dialog'
      backdropForm.className = 'modal-backdrop'
      const backdropBtn = document.createElement('button')
      backdropBtn.type = 'submit'
      backdropBtn.textContent = 'close'
      backdropForm.appendChild(backdropBtn)
      modal.appendChild(backdropForm)

      document.body.appendChild(modal)
      if (typeof modal.showModal === 'function') {
        modal.showModal()
      }
    } catch {
      if (typeof window !== 'undefined' && typeof window.confirm === 'function') {
        resolve(window.confirm(`${task.options.title}\n\n${task.options.message}`))
      } else {
        resolve(false)
      }
    }
  })
}

export function useUiDialog() {
  const store = getStore()

  function open(options: UiDialogOptions): Promise<boolean> {
    return new Promise((resolve) => {
      store.queue.push({
        id: store.nextId++,
        options: normalizeOptions(options),
        resolve,
      })
      pumpQueue(store)
    })
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

  return { open, confirm, danger, alert }
}
