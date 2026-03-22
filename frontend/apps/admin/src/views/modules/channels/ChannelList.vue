<template>
  <div
    class="col-span-12 flex max-h-[min(70vh,560px)] flex-col rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-4"
  >
    <div class="flex shrink-0 items-center justify-between gap-2">
      <div class="text-xs font-semibold text-slate-500">通道列表</div>
      <span v-if="!loading && channels.length" class="font-mono text-[11px] text-slate-400">
        {{ visibleCount }}/{{ channels.length }}
      </span>
    </div>

    <div class="mt-2 shrink-0">
      <input
        v-model.trim="query"
        type="search"
        autocomplete="off"
        placeholder="搜索名称、ID、支付类型…"
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
          v-for="c in filteredChannels"
          :key="c.id"
          type="button"
          class="w-full rounded-xl border px-3 py-3 text-left hover:bg-slate-50"
          :class="selectedId === c.id ? 'border-slate-900' : 'border-slate-200'"
          @click="$emit('select', c.id)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-semibold text-slate-900">{{ c.name }}</div>
              <div class="mt-1 truncate font-mono text-xs text-slate-500">
                #{{ c.id }} · {{ c.pay_type || '-' }}
              </div>
            </div>
            <div class="flex shrink-0 flex-col items-end gap-1">
              <span
                v-if="c.fuse_enabled"
                class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700"
              >
                熔断中
              </span>
              <span
                v-else-if="c.enabled"
                class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
              >
                运行中
              </span>
              <span v-else class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">
                已停用
              </span>
            </div>
          </div>
        </button>
        <div
          v-if="!filteredChannels.length"
          class="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-sm text-slate-500"
        >
          无匹配通道
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import type { AdminChannel } from './types'

const props = defineProps<{
  channels: AdminChannel[]
  loading: boolean
  selectedId: number | null
}>()

defineEmits<{
  select: [id: number]
}>()

const query = ref('')

const filteredChannels = computed(() => {
  const list = props.channels
  const s = query.value.trim().toLowerCase()
  if (!s) return list
  return list.filter((c) => {
    const idStr = String(c.id)
    const name = (c.name || '').toLowerCase()
    const pt = (c.pay_type || '').toLowerCase()
    return idStr.includes(s) || name.includes(s) || pt.includes(s)
  })
})

const visibleCount = computed(() => filteredChannels.value.length)
</script>
