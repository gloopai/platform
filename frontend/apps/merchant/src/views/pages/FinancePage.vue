<template>
  <div class="space-y-8">
    <PageHeader title="财务中心" description="资金流水、提现与对账（部分能力为占位）" />

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <div class="text-sm font-semibold text-slate-900">结算与流水</div>
          <p class="mt-1 text-xs text-slate-500">展示余额变更记录（fund_logs）</p>
        </div>
      </div>
      <div class="mt-4 overflow-hidden rounded-xl border border-slate-100">
        <div class="overflow-x-auto">
          <table class="w-full min-w-[640px] text-left text-sm">
            <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th class="whitespace-nowrap px-4 py-3">时间</th>
                <th class="whitespace-nowrap px-4 py-3">类型</th>
                <th class="whitespace-nowrap px-4 py-3">订单号</th>
                <th class="whitespace-nowrap px-4 py-3">变更</th>
                <th class="whitespace-nowrap px-4 py-3">余额</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td class="px-4 py-8 text-center text-slate-500" colspan="5">加载中…</td>
              </tr>
              <tr v-else-if="logs.length === 0">
                <td class="px-4 py-12 text-center text-slate-500" colspan="5">暂无流水记录</td>
              </tr>
              <tr v-for="l in logs" :key="l.id" class="transition hover:bg-slate-50/80">
                <td class="whitespace-nowrap px-4 py-3 text-slate-700">{{ formatTime(l.created_at) }}</td>
                <td class="px-4 py-3 text-slate-800">{{ l.change_type }}</td>
                <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ l.order_no || '—' }}</td>
                <td class="px-4 py-3 tabular-nums text-slate-800">{{ formatAmount(l.amount) }}</td>
                <td class="px-4 py-3 tabular-nums font-medium text-slate-900">{{ formatAmount(l.balance_after) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <div class="rounded-2xl border border-slate-200/90 bg-gradient-to-br from-white to-slate-50/50 p-6 shadow-sm">
        <div class="flex items-start gap-3">
          <span class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-200/80 text-slate-700">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-1a2 2 0 00-2-2h-1" />
            </svg>
          </span>
          <div>
            <div class="text-sm font-semibold text-slate-900">提现申请</div>
            <p class="mt-1 text-sm text-slate-600">发起余额提现至绑定银行卡（占位能力）</p>
            <button
              type="button"
              class="mt-4 rounded-xl border border-slate-300 bg-white px-4 py-2.5 text-sm font-semibold text-slate-800 shadow-sm transition hover:bg-slate-50"
            >
              发起提现
            </button>
          </div>
        </div>
      </div>
      <div class="rounded-2xl border border-slate-200/90 bg-gradient-to-br from-white to-slate-50/50 p-6 shadow-sm">
        <div class="flex items-start gap-3">
          <span class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-200/80 text-slate-700">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
          </span>
          <div>
            <div class="text-sm font-semibold text-slate-900">对账下载</div>
            <p class="mt-1 text-sm text-slate-600">按日导出 CSV / Excel（占位能力）</p>
            <div class="mt-4 flex flex-wrap gap-2">
              <button
                type="button"
                class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
              >
                下载 CSV
              </button>
              <button
                type="button"
                class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
              >
                下载 Excel
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ErrorCallout v-if="error" :message="error" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { fetchMerchantFundLogs } from '@/api/finance'
import type { MerchantFundLogItem } from '@/types/merchant.api'
import { formatUnixSeconds, formatYuanLabel } from '@/utils/format'

const logs = ref<MerchantFundLogItem[]>([])
const loading = ref(false)
const error = ref('')

function formatAmount(v: number) {
  return formatYuanLabel(v)
}

function formatTime(ts: number) {
  return formatUnixSeconds(ts)
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchMerchantFundLogs(50)
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
