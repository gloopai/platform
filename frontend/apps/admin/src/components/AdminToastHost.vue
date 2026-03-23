<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed inset-x-0 top-4 z-[500] flex flex-col items-center gap-2 px-4 sm:top-5"
      aria-live="polite"
      aria-relevant="additions"
    >
      <TransitionGroup name="admin-toast" tag="div" class="flex w-full max-w-md flex-col gap-2">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="pointer-events-auto flex items-start gap-3 rounded-xl border px-4 py-3 shadow-lg backdrop-blur-sm"
          :class="panelClass(t.variant)"
          role="status"
        >
          <span class="mt-0.5 shrink-0" aria-hidden="true" v-html="iconSvg(t.variant)" />
          <span class="min-w-0 flex-1 text-sm font-medium leading-snug">{{ t.message }}</span>
          <button
            type="button"
            class="shrink-0 rounded-md p-1 text-slate-400 transition hover:bg-black/5 hover:text-slate-700"
            aria-label="关闭"
            @click="dismiss(t.id)"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useAdminToast, type AdminToastVariant } from '../composables/useAdminToast'

const { toasts, dismiss } = useAdminToast()

function panelClass(v: AdminToastVariant) {
  switch (v) {
    case 'error':
      return 'border-rose-200 bg-rose-50/95 text-rose-900 ring-1 ring-rose-100'
    case 'info':
      return 'border-slate-200 bg-white/95 text-slate-800 ring-1 ring-slate-100'
    default:
      return 'border-emerald-200 bg-emerald-50/95 text-emerald-900 ring-1 ring-emerald-100'
  }
}

function iconSvg(v: AdminToastVariant): string {
  if (v === 'error') {
    return `<svg class="h-5 w-5 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  if (v === 'info') {
    return `<svg class="h-5 w-5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  return `<svg class="h-5 w-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
}
</script>

<style scoped>
.admin-toast-enter-active,
.admin-toast-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}
.admin-toast-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}
.admin-toast-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
.admin-toast-move {
  transition: transform 0.2s ease;
}
</style>
