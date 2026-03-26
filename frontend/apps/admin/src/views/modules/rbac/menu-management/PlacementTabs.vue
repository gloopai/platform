<template>
  <div class="inline-flex rounded-xl border border-slate-200 bg-slate-100/80 p-1 shadow-sm">
    <button
      v-for="t in tabs"
      :key="t.key"
      type="button"
      class="rounded-lg px-4 py-2 text-sm font-medium transition"
      :class="
        modelValue === t.key
          ? 'bg-white text-slate-900 shadow-sm ring-1 ring-slate-200'
          : 'text-slate-600 hover:text-slate-800'
      "
      @click="$emit('update:modelValue', t.key)"
    >
      {{ t.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import type { MenuMgmtTab } from './types'

defineProps<{
  modelValue: MenuMgmtTab
}>()

defineEmits<{
  (e: 'update:modelValue', v: MenuMgmtTab): void
}>()

const tabs: { key: MenuMgmtTab; label: string }[] = [
  { key: 'left', label: '左侧菜单' },
  { key: 'avatar', label: '头像下方' },
  { key: 'other', label: '其它功能' },
]
</script>
