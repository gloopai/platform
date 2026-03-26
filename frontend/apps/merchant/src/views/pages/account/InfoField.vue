<template>
  <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
    <div class="text-xs text-slate-500">{{ label }}</div>
    <div class="mt-1 flex items-center gap-2">
      <div class="min-w-0 flex-1 truncate text-sm font-medium text-slate-900" :title="value">{{ value }}</div>
      <slot name="extra" />
      <button
        v-if="copyable"
        type="button"
        class="shrink-0 rounded-lg border border-slate-200 bg-white px-2 py-1 text-[11px] font-semibold text-slate-700 hover:border-slate-300"
        @click="copyValue"
      >
        复制
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    label: string
    value: string
    copyable?: boolean
  }>(),
  {
    copyable: false,
  },
)

async function copyValue() {
  if (!props.value || props.value === '-') return
  try {
    await navigator.clipboard.writeText(props.value)
  } catch {
    // noop
  }
}
</script>
