<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">结算记录</div>
      <div class="mt-2 text-sm text-slate-600">展示余额变更流水（fund_logs）。</div>
      <div class="mt-4 overflow-hidden rounded-xl border border-slate-200">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs font-semibold text-slate-600">
            <tr>
              <th class="px-4 py-3">时间</th>
              <th class="px-4 py-3">类型</th>
              <th class="px-4 py-3">订单号</th>
              <th class="px-4 py-3">变更</th>
              <th class="px-4 py-3">余额</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr v-if="loading">
              <td class="px-4 py-3 text-slate-600" colspan="5">加载中...</td>
            </tr>
            <tr v-else-if="logs.length === 0">
              <td class="px-4 py-3 text-slate-600" colspan="5">暂无数据</td>
            </tr>
            <tr v-for="l in logs" :key="l.id">
              <td class="px-4 py-3 text-slate-700">{{ formatTime(l.created_at) }}</td>
              <td class="px-4 py-3 text-slate-700">{{ l.change_type }}</td>
              <td class="px-4 py-3 text-slate-700">{{ l.order_no || '-' }}</td>
              <td class="px-4 py-3 text-slate-700">{{ formatAmount(l.amount) }}</td>
              <td class="px-4 py-3 text-slate-700">{{ formatAmount(l.balance_after) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-sm font-semibold text-slate-900">提现申请</div>
        <div class="mt-2 text-sm text-slate-600">手动发起余额提现到绑定银行卡（占位）。</div>
        <button class="mt-4 rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white">发起提现</button>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-sm font-semibold text-slate-900">对账下载</div>
        <div class="mt-2 text-sm text-slate-600">按日下载 CSV/Excel 对账单（占位）。</div>
        <div class="mt-4 flex flex-wrap gap-2">
          <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">下载 CSV</button>
          <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">下载 Excel</button>
        </div>
      </div>
    </div>
    <div v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { merchantConsoleGet } from '../../lib/merchantApi'

type MerchantFundLogItem = {
  id: number
  order_no: string
  change_type: string
  amount: number
  balance_before: number
  balance_after: number
  reason: string
  created_at: number
}

const logs = ref<MerchantFundLogItem[]>([])
const loading = ref(false)
const error = ref('')

function formatAmount(v: number) {
  return `¥ ${(v / 100).toFixed(2)}`
}

function formatTime(ts: number) {
  const d = new Date(ts * 1000)
  return d.toLocaleString()
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await merchantConsoleGet<{ logs: MerchantFundLogItem[] }>('/v1/merchant/fund_logs', { limit: 50 })
    logs.value = res.logs || []
  } catch {
    error.value = '加载失败：请确认已登录且网关已启动。'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void reload()
})
</script>
