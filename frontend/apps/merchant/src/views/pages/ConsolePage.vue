<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">控制台</h1>
      <p class="mt-1 text-sm text-slate-600">今日经营概况与账户状态一目了然</p>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-emerald-200/80 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-emerald-400 opacity-40 blur-2xl transition group-hover:opacity-60" />
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

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-emerald-200/80 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-cyan-400 opacity-40 blur-2xl transition group-hover:opacity-60" />
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

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-emerald-200/80 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-teal-400 opacity-40 blur-2xl transition group-hover:opacity-60" />
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

      <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-emerald-200/80 hover:shadow-md">
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-amber-400 opacity-40 blur-2xl transition group-hover:opacity-60" />
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

    <div class="overflow-hidden rounded-2xl border border-emerald-200/60 bg-gradient-to-br from-emerald-50/90 to-white p-6 shadow-sm">
      <div class="flex flex-wrap items-start gap-3">
        <span class="mt-0.5 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-700">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </span>
        <div class="min-w-0 flex-1">
          <div class="text-sm font-semibold text-emerald-900">使用提示</div>
          <p class="mt-1 text-sm leading-relaxed text-emerald-900/80">{{ tipText }}</p>
        </div>
      </div>
      <div v-if="error" class="mt-4 rounded-xl border border-rose-200 bg-white px-4 py-3 text-sm text-rose-800">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { merchantConsoleGet } from '../../lib/merchantApi'

type MerchantSummaryResp = {
  today_amount: number
  today_count: number
  success_rate: number
  balance: number
  merchant_id: string
}

const summary = ref<MerchantSummaryResp | null>(null)
const error = ref('')

const todayAmountText = computed(() => {
  const v = summary.value?.today_amount ?? 0
  return `¥ ${(v / 100).toFixed(2)}`
})

const balanceText = computed(() => {
  const v = summary.value?.balance ?? 0
  return `¥ ${(v / 100).toFixed(2)}`
})

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

async function load() {
  error.value = ''
  try {
    summary.value = await merchantConsoleGet<MerchantSummaryResp>('/v1/merchant/summary')
  } catch {
    error.value = '数据加载失败：请确认已登录且网关服务正常运行。'
  }
}

onMounted(() => {
  void load()
})
</script>
