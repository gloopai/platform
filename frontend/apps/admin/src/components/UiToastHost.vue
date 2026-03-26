<template>
  <div class="toast toast-top toast-center z-[2147483600]">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="pointer-events-auto flex min-w-[280px] max-w-[560px] items-center gap-2 rounded-full border px-3 py-1.5 shadow-lg"
      :class="chipClass(t.variant)"
      role="status"
    >
      <span aria-hidden="true" class="opacity-90" v-html="iconSvg(t.variant)" />
      <span class="truncate text-xs font-semibold tracking-wide">{{ t.message }}</span>
      <button
        type="button"
        class="ml-auto inline-flex h-5 w-5 items-center justify-center rounded-full text-[11px] opacity-70 transition hover:opacity-100"
        aria-label="关闭"
        @click="dismiss(t.id)"
      >
        ✕
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUiToast, type UiToastVariant } from '../composables/useUiToast'

const { toasts, dismiss } = useUiToast()

function chipClass(v: UiToastVariant) {
  switch (v) {
    case 'error':
      return 'border-red-500 bg-red-500 text-white'
    case 'warning':
      return 'border-orange-500 bg-orange-500 text-white'
    case 'info':
      return 'border-blue-500 bg-blue-500 text-white'
    default:
      return 'border-green-600 bg-green-600 text-white'
  }
}

function iconSvg(v: UiToastVariant): string {
  if (v === 'error') {
    return `<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  if (v === 'info') {
    return `<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
  }
  if (v === 'warning') {
    return `<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 4h.01M10.29 3.86l-7.5 13A1 1 0 003.66 18h16.68a1 1 0 00.87-1.5l-7.5-13a1 1 0 00-1.74 0z" /></svg>`
  }
  return `<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
}
</script>
