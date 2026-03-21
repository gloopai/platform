<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">今日看板</div>
      <div class="mt-4 grid gap-3 md:grid-cols-4">
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">今日流水</div>
          <div class="mt-2 text-xl font-semibold text-slate-900">{{ todayAmountText }}</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">订单数</div>
          <div class="mt-2 text-xl font-semibold text-slate-900">{{ summary?.today_count ?? '-' }}</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">成功率</div>
          <div class="mt-2 text-xl font-semibold text-slate-900">{{ successRateText }}</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">待结算余额</div>
          <div class="mt-2 text-xl font-semibold text-slate-900">{{ balanceText }}</div>
        </div>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">提示</div>
      <div class="mt-2 text-sm text-slate-600">{{ tipText }}</div>
      <div v-if="error" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
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
  if (v === undefined || v === null) return '-'
  return `${(v * 100).toFixed(2)}%`
})

const tipText = computed(() => {
  if (!summary.value) return '请先在“开发配置”填写 merchant_id/app_secret，然后刷新页面。'
  return '可在“开发配置”里创建订单并联调回调链路，或在“交易管理”查看订单列表。'
})

async function load() {
  error.value = ''
  try {
    summary.value = await merchantConsoleGet<MerchantSummaryResp>('/v1/merchant/summary')
  } catch {
    error.value = '加载失败：请确认已登录且网关已启动。'
  }
}

onMounted(() => {
  void load()
})
</script>
