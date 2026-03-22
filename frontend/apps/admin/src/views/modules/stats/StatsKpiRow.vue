<template>
  <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-xs font-medium text-slate-500">今日成功交易额（元）</div>
      <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">{{ yuan(totals.paid_amount) }}</div>
      <div class="mt-1 text-[10px] text-slate-400">成功订单金额汇总</div>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-xs font-medium text-slate-500">今日创建订单笔数</div>
      <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">{{ totals.order_count }}</div>
      <div class="mt-1 text-[10px] text-slate-400">待支付 / 成功 / 失败 / 关单见下表</div>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-xs font-medium text-slate-500">成交率（成功/今日创建）</div>
      <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">{{ pct(totals.conversion_rate_pct) }}</div>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-xs font-medium text-slate-500">支付成功率（成功/成功+失败）</div>
      <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">{{ pct(totals.terminal_success_rate_pct) }}</div>
      <div class="mt-1 text-[10px] text-slate-400">不含待支付、关单</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatPct, formatYuan } from './format'
import type { StatsTotals } from './types'

defineProps<{
  totals: StatsTotals
}>()

function yuan(cents: number) {
  return formatYuan(cents)
}
function pct(n: number) {
  return formatPct(n)
}
</script>
