<template>
  <div class="flex flex-wrap gap-3 text-sm text-slate-600">
    <span class="rounded-full bg-slate-100 px-3 py-1">待支付 <strong class="text-slate-900">{{ totals.pending_count }}</strong></span>
    <span class="rounded-full bg-emerald-50 px-3 py-1 text-emerald-900"
      >成功 <strong>{{ totals.paid_count }}</strong></span
    >
    <span class="rounded-full bg-rose-50 px-3 py-1 text-rose-900"
      >失败 <strong>{{ totals.failed_count }}</strong></span
    >
    <span class="rounded-full bg-slate-100 px-3 py-1">关单 <strong class="text-slate-900">{{ totals.closed_count }}</strong></span>
    <span class="rounded-full bg-slate-100 px-3 py-1"
      >启用通道 <strong class="text-slate-900">{{ enabledChannels }}</strong> · 熔断
      <strong class="text-amber-800">{{ fusedChannels }}</strong></span
    >
  </div>
</template>

<script setup lang="ts">
import type { StatsTotals } from './types'

defineProps<{
  totals: StatsTotals
  enabledChannels: number
  fusedChannels: number
}>()
</script>
