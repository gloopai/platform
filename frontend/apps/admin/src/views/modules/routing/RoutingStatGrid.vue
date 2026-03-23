<template>
  <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
    <div
      v-for="cell in cells"
      :key="cell.label"
      class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm"
    >
      <div class="text-xs font-medium text-slate-500">{{ cell.label }}</div>
      <div class="mt-1 text-2xl font-semibold tabular-nums text-slate-900">{{ cell.value }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { RoutingSummary } from './types'

const props = defineProps<{
  summary: RoutingSummary | null
  loading: boolean
}>()

const cells = computed(() => {
  const s = props.summary
  const dash = props.loading ? '…' : '—'
  return [
    { label: '代收产品（启用）', value: s ? s.enabled_payin_products : dash },
    { label: '代付产品（启用）', value: s ? s.enabled_payout_products : dash },
    { label: '上游通道（启用）', value: s ? s.enabled_channels : dash },
    { label: '代收绑定（启用）', value: s ? s.active_bindings : dash },
    { label: '代付绑定（启用）', value: s ? s.active_payout_bindings : dash },
    { label: '代收白名单商户', value: s ? s.merchants_with_payin_whitelist : dash },
    { label: '代付白名单商户', value: s ? s.merchants_with_payout_whitelist : dash },
    { label: '熔断中的通道', value: s ? s.fused_channels : dash },
  ]
})
</script>
