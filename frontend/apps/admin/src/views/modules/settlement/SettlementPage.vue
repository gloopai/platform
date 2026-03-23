<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">结算与提现</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：先提供平台侧<strong>资金流水视图</strong>（按商户可筛选），用于核对入账；结算单、提现审核、打款流程后续迭代。当前可配合
        <router-link to="/orders" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          全站订单
        </router-link>
        与
        <router-link to="/reconcile" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          对账中心
        </router-link>
        核对平台侧数据口径。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="flex flex-wrap items-end gap-3">
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">商户（可选）</span>
        <input
          v-model.trim="merchantId"
          type="text"
          placeholder="merchant_id"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @keyup.enter="load"
        />
      </label>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
        @click="load"
      >
        加载
      </button>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[980px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">时间</th>
              <th class="whitespace-nowrap px-4 py-3">商户</th>
              <th class="whitespace-nowrap px-4 py-3">订单号</th>
              <th class="whitespace-nowrap px-4 py-3">类型</th>
              <th class="whitespace-nowrap px-4 py-3">变动金额</th>
              <th class="whitespace-nowrap px-4 py-3">变动前</th>
              <th class="whitespace-nowrap px-4 py-3">变动后</th>
              <th class="whitespace-nowrap px-4 py-3">原因</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中…</td>
            </tr>
            <tr v-else-if="!logs.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="8">暂无资金流水</td>
            </tr>
            <tr v-for="x in pagedLogs" v-else :key="x.id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatTs(x.created_at) }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ x.merchant_id }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ x.order_no }}</td>
              <td class="px-4 py-3">{{ x.change_type }}</td>
              <td class="px-4 py-3 font-semibold" :class="x.amount >= 0 ? 'text-emerald-700' : 'text-rose-700'">{{ formatAmount(x.amount) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatAmount(x.balance_before) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatAmount(x.balance_after) }}</td>
              <td class="px-4 py-3 text-slate-700">{{ x.reason || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <AdminPaginationBar
        v-if="!loading && logs.length > 0"
        v-model:page="page"
        v-model:pageSize="pageSize"
        :total="total"
        :page-count="pageCount"
      />
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划（未实现）</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>结算单生成、商户确认与 T+N / D+1 策略</li>
        <li>提现申请、审核与打款通道</li>
        <li>与对账中心差异批次联动</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { useClientPagination } from '../../../composables/useClientPagination'

import { adminGet } from '../../../lib/adminApi'

type SettlementLogItem = {
  id: number
  merchant_id: string
  order_no: string
  change_type: string
  amount: number
  balance_before: number
  balance_after: number
  reason: string
  created_at: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const merchantId = ref('')
const loading = ref(true)
const error = ref('')
const logs = ref<SettlementLogItem[]>([])
const { page, pageSize, total, pageCount, slice: pagedLogs } = useClientPagination(logs, 20)

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime())
    ? '—'
    : d.toLocaleString('zh-CN', { hour12: false, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatAmount(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const abs = Math.abs(cents)
  return `${sign}${(abs / 100).toFixed(2)}`
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    page.value = 1
    const q = new URLSearchParams()
    if (merchantId.value) q.set('merchant_id', merchantId.value)
    q.set('limit', '200')
    const r = await adminGet<{ logs: SettlementLogItem[] }>(`/v1/admin/settlement/logs?${q.toString()}`)
    logs.value = r.logs ?? []
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    logs.value = []
  } finally {
    loading.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void load()
  if (registerRefresh) unregister = registerRefresh(() => void load())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
