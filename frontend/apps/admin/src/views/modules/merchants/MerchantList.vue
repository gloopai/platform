<template>
  <div
    class="col-span-12 flex max-h-[min(70vh,560px)] flex-col rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-4"
  >
    <div class="flex shrink-0 items-center justify-between gap-2">
      <div class="text-xs font-semibold text-slate-500">商户列表</div>
      <span v-if="!loading && merchants.length" class="font-mono text-[11px] text-slate-400">
        {{ visibleCount }}/{{ merchants.length }}
      </span>
    </div>

    <div class="mt-2 shrink-0">
      <input
        v-model.trim="query"
        type="search"
        autocomplete="off"
        placeholder="搜索商户 ID…"
        class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
      />
    </div>

    <div v-if="loading" class="mt-3 shrink-0 text-sm text-slate-500">加载中...</div>
    <div
      v-else
      class="mt-1 min-h-0 flex-1 overflow-y-auto overscroll-contain pt-2 [scrollbar-gutter:stable]"
    >
      <div class="space-y-2 pr-0.5">
        <button
          v-for="m in filteredMerchants"
          :key="m.merchant_id"
          type="button"
          class="w-full rounded-xl border px-3 py-3 text-left hover:bg-slate-50"
          :class="selectedId === m.merchant_id ? 'border-slate-900' : 'border-slate-200'"
          @click="$emit('select', m.merchant_id)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <div class="truncate font-mono text-sm font-semibold text-slate-900">{{ m.merchant_id }}</div>
              <div class="mt-1 text-xs text-slate-500">余额 {{ formatMoney(m.balance) }}</div>
            </div>
            <div class="flex shrink-0 flex-col items-end gap-1">
              <span
                v-if="m.status === 1"
                class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
              >
                启用
              </span>
              <span v-else class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700">锁定</span>
            </div>
          </div>
        </button>
        <div
          v-if="!filteredMerchants.length"
          class="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-sm text-slate-500"
        >
          无匹配商户
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import type { AdminMerchantInfo } from './types'

const props = defineProps<{
  merchants: AdminMerchantInfo[]
  loading: boolean
  selectedId: string | null
}>()

defineEmits<{
  select: [merchantId: string]
}>()

const query = ref('')

function formatMoney(v: number) {
  return `¥ ${(v / 100).toFixed(2)}`
}

const filteredMerchants = computed(() => {
  const list = props.merchants
  const s = query.value.trim().toLowerCase()
  if (!s) return list
  return list.filter((m) => (m.merchant_id || '').toLowerCase().includes(s))
})

const visibleCount = computed(() => filteredMerchants.value.length)
</script>
