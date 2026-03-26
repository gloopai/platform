<template>
  <Teleport to="body">
    <Transition name="ui-dialog-fade">
      <div
        v-if="current"
        class="fixed inset-0 z-[600] flex items-center justify-center bg-slate-950/45 px-4"
        @click.self="closeCurrent(false)"
      >
        <Transition name="ui-dialog-pop">
          <section
            v-if="current"
            class="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-5 shadow-2xl"
            role="dialog"
            aria-modal="true"
            :aria-label="current.options.title"
          >
            <header class="flex items-start gap-3">
              <span class="mt-0.5 shrink-0" aria-hidden="true" v-html="iconSvg(current.options.variant)" />
              <div class="min-w-0">
                <h3 class="text-base font-semibold text-slate-900">{{ current.options.title }}</h3>
                <p class="mt-2 whitespace-pre-wrap text-sm leading-relaxed text-slate-600">
                  {{ current.options.message }}
                </p>
              </div>
            </header>
            <footer class="mt-5 flex justify-end gap-2">
              <button
                v-if="!current.options.hideCancel"
                type="button"
                class="min-w-[72px] rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm transition duration-150 hover:border-slate-300 hover:bg-slate-50 active:scale-[0.99] focus:outline-none focus:ring-2 focus:ring-slate-300/60"
                @click="closeCurrent(false)"
              >
                {{ current.options.cancelText }}
              </button>
              <button
                type="button"
                class="min-w-[78px] rounded-lg px-3 py-1.5 text-xs font-semibold text-white shadow-sm transition duration-150 active:scale-[0.99]"
                :class="confirmButtonClass(current.options.variant)"
                @click="closeCurrent(true)"
              >
                {{ current.options.confirmText }}
              </button>
            </footer>
          </section>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useUiDialog, type UiDialogVariant } from './useUiDialog'

const { current, closeCurrent } = useUiDialog()

function confirmButtonClass(variant: UiDialogVariant) {
  if (variant === 'danger') {
    return 'bg-rose-600 hover:bg-rose-700 focus:outline-none focus:ring-2 focus:ring-rose-500/40'
  }
  return 'bg-slate-900 hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-400/40'
}

function iconSvg(variant: UiDialogVariant): string {
  if (variant === 'danger') {
    return `<svg class="h-5 w-5 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 4h.01M5.07 19h13.86c1.54 0 2.5-1.67 1.73-3L13.73 4c-.77-1.33-2.69-1.33-3.46 0L3.34 16c-.77 1.33.19 3 1.73 3z" /></svg>`
  }
  return `<svg class="h-5 w-5 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
}
</script>

<style scoped>
.ui-dialog-fade-enter-active,
.ui-dialog-fade-leave-active {
  transition: opacity 0.18s ease;
}
.ui-dialog-fade-enter-from,
.ui-dialog-fade-leave-to {
  opacity: 0;
}
.ui-dialog-pop-enter-active,
.ui-dialog-pop-leave-active {
  transition:
    transform 0.18s ease,
    opacity 0.18s ease;
}
.ui-dialog-pop-enter-from,
.ui-dialog-pop-leave-to {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
