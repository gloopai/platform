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
        class="pointer-events-auto relative flex items-start gap-3 rounded-xl border bg-white px-4 py-3 shadow-[0_14px_34px_-14px_rgba(15,23,42,0.42)]"
        :class="panelClass(t.variant)"
        role="status"
      >
        <span
          class="absolute inset-y-2 left-1.5 w-1 rounded-full"
          :class="accentClass(t.variant)"
          aria-hidden="true"
        />
        <span
          class="ml-2 mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border"
          :class="iconWrapClass(t.variant)"
          aria-hidden="true"
          v-html="iconSvg(t.variant)"
        />
        <span class="min-w-0 flex-1 text-sm font-medium leading-snug text-slate-700">{{ t.message }}</span>
        <button
          type="button"
          class="shrink-0 rounded-md p-1 text-slate-300 transition hover:bg-slate-100 hover:text-slate-600"
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
      return 'border-rose-200'
    case 'info':
      return 'border-slate-300'
    default:
      return 'border-emerald-200'
  }
}

function accentClass(v: UiToastVariant) {
  switch (v) {
    case 'error':
      return 'bg-rose-400'
    case 'info':
      return 'bg-slate-400'
    default:
      return 'bg-emerald-400'
  }
}

function iconWrapClass(v: UiToastVariant) {
  switch (v) {
    case 'error':
      return 'border-rose-100 bg-rose-50/80'
    case 'info':
      return 'border-slate-200 bg-slate-50'
    default:
      return 'border-emerald-100 bg-emerald-50/80'
  }
}

function iconSvg(v: UiToastVariant): string {
  if (v === 'error') {
    return `<svg class="h-5 w-5 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  if (v === 'info') {
    return `<svg class="h-5 w-5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  return `<svg class="h-5 w-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
}
</script>
