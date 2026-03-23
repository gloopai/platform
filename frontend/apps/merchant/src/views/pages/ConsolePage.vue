<template>
  <div class="space-y-8">
    <PageHeader title="控制台" description="今日经营概况与账户状态一目了然" />

    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-300 opacity-25 blur-2xl transition group-hover:opacity-40" />
        <div class="relative">
          <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </span>
            今日流水
          </div>
          <div class="mt-3 text-2xl font-semibold tabular-nums tracking-tight text-slate-900 sm:text-3xl">{{ todayAmountText }}</div>
        </div>
      </div>

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-400 opacity-20 blur-2xl transition group-hover:opacity-40" />
        <div class="relative">
          <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </span>
            今日订单
          </div>
          <div class="mt-3 text-2xl font-semibold tabular-nums tracking-tight text-slate-900 sm:text-3xl">{{ orderCountText }}</div>
        </div>
      </div>

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-300 opacity-20 blur-2xl transition group-hover:opacity-40" />
        <div class="relative">
          <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </span>
            支付成功率
          </div>
          <div class="mt-3 text-2xl font-semibold tabular-nums tracking-tight text-slate-900 sm:text-3xl">{{ successRateText }}</div>
        </div>
      </div>

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-400 opacity-15 blur-2xl transition group-hover:opacity-30" />
        <div class="relative">
          <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
              </svg>
            </span>
            待结算余额
          </div>
          <div class="mt-3 text-2xl font-semibold tabular-nums tracking-tight text-slate-900 sm:text-3xl">{{ balanceText }}</div>
        </div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/80 bg-gradient-to-br from-slate-50/95 to-white p-6 shadow-sm">
      <div class="flex flex-wrap items-start gap-3">
        <span class="mt-0.5 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-slate-200/80 text-slate-700">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </span>
        <div class="min-w-0 flex-1">
          <div class="text-sm font-semibold text-slate-900">使用提示</div>
          <p class="mt-1 text-sm leading-relaxed text-slate-600">{{ tipText }}</p>
        </div>
      </div>
      <ErrorCallout v-if="error" class="mt-4" :message="error" />
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/80 px-5 py-3">
        <div class="text-sm font-semibold text-slate-900">今日按支付产品</div>
        <div class="mt-0.5 text-xs text-slate-500">按支付产品统计订单量、成交额与成功率</div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-white text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">支付产品</th>
              <th class="whitespace-nowrap px-4 py-3">订单数</th>
              <th class="whitespace-nowrap px-4 py-3">成功笔数</th>
              <th class="whitespace-nowrap px-4 py-3">失败笔数</th>
              <th class="whitespace-nowrap px-4 py-3">成交额</th>
              <th class="whitespace-nowrap px-4 py-3">成功率</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="6">加载中…</td>
            </tr>
            <tr v-else-if="(byProduct?.items?.length || 0) === 0">
              <td class="px-4 py-8 text-center text-slate-500" colspan="6">暂无今日支付产品数据</td>
            </tr>
            <tr v-for="x in byProduct?.items || []" v-else :key="x.pay_product_code" class="hover:bg-slate-50/80">
              <td class="px-4 py-3">
                <div class="font-medium text-slate-900">{{ x.pay_product_name || x.pay_product_code }}</div>
                <div class="mt-0.5 font-mono text-xs text-slate-500">{{ x.pay_product_code || '—' }}</div>
              </td>
              <td class="px-4 py-3 tabular-nums text-slate-800">{{ x.order_count }}</td>
              <td class="px-4 py-3 tabular-nums text-emerald-700">{{ x.paid_count }}</td>
              <td class="px-4 py-3 tabular-nums text-rose-700">{{ x.failed_count }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-900">{{ formatYuanLabel(x.paid_amount) }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ x.success_rate_pct.toFixed(2) }}%</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { useMerchantSummary } from '@/composables/useMerchantSummary'
import { formatYuanLabel } from '@/utils/format'

const { summary, byProduct, error, loading } = useMerchantSummary()

const todayAmountText = computed(() => formatYuanLabel(summary.value?.today_amount ?? 0))

const balanceText = computed(() => formatYuanLabel(summary.value?.balance ?? 0))

const successRateText = computed(() => {
  const v = summary.value?.success_rate
  if (v === undefined || v === null) return '—'
  return `${(v * 100).toFixed(2)}%`
})

const orderCountText = computed(() => {
  const v = summary.value?.today_count
  if (v === undefined || v === null) return '—'
  return String(v)
})

const tipText = computed(() => {
  if (!summary.value) return '连接数据后即可查看实时指标。请先在「开发配置」中确认商户参数，并确保网关服务已启动。'
  return '可在「开发配置」中联调下单与回调，在「交易管理」中查询订单与通知记录。'
})
</script>
