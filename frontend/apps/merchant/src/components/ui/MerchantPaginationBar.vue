<template>
  <div
    class="flex flex-col gap-3 border-t border-slate-200 px-4 py-3 text-sm text-slate-600 sm:flex-row sm:items-center sm:justify-between"
  >
    <div class="flex flex-wrap items-center gap-2">
      <span>共 <span class="font-semibold text-slate-900">{{ total }}</span> 条</span>
      <label class="flex items-center gap-2">
        <span class="text-slate-500">每页</span>
        <select
          :value="pageSize"
          class="rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-sm font-medium text-slate-800"
          @change="$emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="n in pageSizeOptions" :key="n" :value="n">{{ n }}</option>
        </select>
      </label>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 font-medium text-slate-800 disabled:opacity-40"
        :disabled="page <= 1"
        @click="$emit('update:page', page - 1)"
      >
        上一页
      </button>
      <span class="tabular-nums text-slate-500">
        第 <span class="font-medium text-slate-900">{{ page }}</span> / {{ safePageCount }} 页
      </span>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 font-medium text-slate-800 disabled:opacity-40"
        :disabled="page >= safePageCount"
        @click="$emit('update:page', page + 1)"
      >
        下一页
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    total: number
    page: number
    pageSize: number
    pageCount: number
    pageSizeOptions?: number[]
  }>(),
  { pageSizeOptions: () => [10, 20, 50, 100] },
)

defineEmits<{
  'update:page': [v: number]
  'update:pageSize': [v: number]
}>()

const safePageCount = computed(() => Math.max(1, props.pageCount))
</script>
