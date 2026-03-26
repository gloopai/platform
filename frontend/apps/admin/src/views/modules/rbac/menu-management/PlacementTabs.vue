<template>
  <div class="flex flex-wrap gap-2 border-b border-slate-200 pb-3">
    <button
      v-for="t in tabs"
      :key="t.key"
      type="button"
      class="rounded-lg px-3 py-2 text-sm font-medium transition"
      :class="
        modelValue === t.key
          ? 'bg-indigo-600 text-white shadow-sm'
          : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'
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
