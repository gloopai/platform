<template>
  <div
    class="pointer-events-none fixed inset-x-0 top-4 z-[2147483600] flex flex-col items-center gap-2 px-4 sm:top-5"
    aria-live="polite"
    aria-relevant="additions"
  >
    <div class="flex w-full max-w-md flex-col gap-2">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto flex items-start gap-3 rounded-xl border px-4 py-3 shadow-lg backdrop-blur-sm"
        :class="panelClass(t.variant)"
        role="status"
      >
        <span class="mt-0.5 shrink-0 text-sm" aria-hidden="true">{{ iconText(t.variant) }}</span>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUiToast, type UiToastVariant } from './useUiToast'

const { toasts, dismiss } = useUiToast()

function panelClass(v: UiToastVariant) {
  switch (v) {
    case 'error':
      return 'border-rose-200 bg-rose-50/95 text-rose-900 ring-1 ring-rose-100'
    case 'info':
      return 'border-slate-200 bg-white/95 text-slate-800 ring-1 ring-slate-100'
    default:
      return 'border-emerald-200 bg-emerald-50/95 text-emerald-900 ring-1 ring-emerald-100'
  }
}

function iconText(v: UiToastVariant) {
  if (v === 'error') return '!'
  if (v === 'info') return 'i'
  return 'ok'
}
</script>
