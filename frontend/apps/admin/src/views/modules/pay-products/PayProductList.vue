<template>
  <div class="col-span-12 flex max-h-[min(70vh,560px)] flex-col rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-4">
    <div class="flex shrink-0 items-center justify-between gap-2">
      <div class="text-xs font-semibold text-slate-500">支付产品</div>
      <span v-if="!loading && products.length" class="font-mono text-[11px] text-slate-400">
        {{ visibleCount }}/{{ products.length }}
      </span>
    </div>

    <div class="mt-2 shrink-0">
      <input
        v-model.trim="query"
        type="search"
        autocomplete="off"
        placeholder="搜索编码、名称、ID…"
        class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
      />
    </div>

    <div v-if="loading" class="mt-3 shrink-0 text-sm text-slate-500">加载中...</div>
    <div
      v-else
      class="mt-1 min-h-0 flex-1 overflow-y-auto overscroll-contain pt-2 [-ms-overflow-style:none] [scrollbar-gutter:stable]"
    >
      <div class="space-y-2 pr-0.5">
        <button
          v-for="p in filteredProducts"
          :key="p.id"
          type="button"
          class="w-full rounded-xl border px-3 py-3 text-left hover:bg-slate-50"
          :class="selectedId === p.id ? 'border-slate-900' : 'border-slate-200'"
          @click="$emit('select', p.id)"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-semibold text-slate-900">{{ p.name }}</div>
              <div class="mt-1 truncate font-mono text-xs text-slate-500">{{ p.code }}</div>
            </div>
            <span
              v-if="p.enabled"
              class="shrink-0 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
            >
              启用
            </span>
            <span v-else class="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">
              停用
            </span>
          </div>
        </button>
        <div v-if="!filteredProducts.length" class="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-sm text-slate-500">
          无匹配产品
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import type { PayProduct } from './types'

const props = defineProps<{
  products: PayProduct[]
  loading: boolean
  selectedId: number | null
}>()

defineEmits<{
  select: [id: number]
}>()

const query = ref('')

const filteredProducts = computed(() => {
  const list = props.products
  const s = query.value.trim().toLowerCase()
  if (!s) return list
  return list.filter((p) => {
    const idStr = String(p.id)
    return (
      idStr.includes(s) ||
      (p.code || '').toLowerCase().includes(s) ||
      (p.name || '').toLowerCase().includes(s)
    )
  })
})

const visibleCount = computed(() => filteredProducts.value.length)
</script>
