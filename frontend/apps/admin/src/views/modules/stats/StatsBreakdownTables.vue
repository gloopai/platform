<template>
  <div class="grid gap-6 lg:grid-cols-2">
    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="border-b border-slate-100 px-5 py-3">
        <div class="text-sm font-semibold text-slate-900">按支付产品</div>
        <p class="mt-0.5 text-xs text-slate-500">对外产品维度：笔数、成交额、成交率、支付成功率</p>
      </div>
      <div class="max-h-[min(70vh,420px)] overflow-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="sticky top-0 bg-slate-50 text-xs font-medium text-slate-500">
            <tr>
              <th class="px-4 py-2">产品</th>
              <th class="px-2 py-2 text-right">创建</th>
              <th class="px-2 py-2 text-right">成交额</th>
              <th class="px-2 py-2 text-right">成交率</th>
              <th class="px-4 py-2 text-right">支付成功率</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="row in products" :key="row.product_code" class="hover:bg-slate-50/80">
              <td class="px-4 py-2">
                <div class="font-medium text-slate-900">{{ row.product_name }}</div>
                <div class="font-mono text-[11px] text-slate-500">{{ row.product_code }}</div>
              </td>
              <td class="px-2 py-2 text-right tabular-nums">{{ row.order_count }}</td>
              <td class="px-2 py-2 text-right tabular-nums">{{ yuan(row.paid_amount) }}</td>
              <td class="px-2 py-2 text-right text-xs">{{ pct(row.conversion_rate_pct) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ pct(row.terminal_success_rate_pct) }}</td>
            </tr>
            <tr v-if="products.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-sm text-slate-500">今日暂无订单数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="border-b border-slate-100 px-5 py-3">
        <div class="text-sm font-semibold text-slate-900">按上游通道</div>
        <p class="mt-0.5 text-xs text-slate-500">路由落库后的通道维度（未路由订单在 channel_id=0）</p>
      </div>
      <div class="max-h-[min(70vh,420px)] overflow-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="sticky top-0 bg-slate-50 text-xs font-medium text-slate-500">
            <tr>
              <th class="px-4 py-2">通道</th>
              <th class="px-2 py-2 text-right">创建</th>
              <th class="px-2 py-2 text-right">成交额</th>
              <th class="px-2 py-2 text-right">成交率</th>
              <th class="px-4 py-2 text-right">支付成功率</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="row in channels" :key="row.channel_id" class="hover:bg-slate-50/80">
              <td class="px-4 py-2">
                <div class="font-medium text-slate-900">{{ row.channel_name }}</div>
                <div class="font-mono text-[11px] text-slate-500">id={{ row.channel_id }}</div>
              </td>
              <td class="px-2 py-2 text-right tabular-nums">{{ row.order_count }}</td>
              <td class="px-2 py-2 text-right tabular-nums">{{ yuan(row.paid_amount) }}</td>
              <td class="px-2 py-2 text-right text-xs">{{ pct(row.conversion_rate_pct) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ pct(row.terminal_success_rate_pct) }}</td>
            </tr>
            <tr v-if="channels.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-sm text-slate-500">今日暂无订单数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatPct, formatYuan } from './format'
import type { StatsChannelRow, StatsProductRow } from './types'

defineProps<{
  products: StatsProductRow[]
  channels: StatsChannelRow[]
}>()

function yuan(cents: number) {
  return formatYuan(cents)
}
function pct(n: number) {
  return formatPct(n)
}
</script>
