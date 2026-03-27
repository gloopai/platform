<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">资金流水</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">商户资金变动明细（入账、划转、提现扣减）；金额单位为系统配置的法币。</p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-2 lg:grid-cols-6">
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
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">变动类型</span>
        <select v-model="changeType" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm">
          <option value="">全部</option>
          <option value="ORDER_PAID">ORDER_PAID</option>
          <option value="PAYIN_TO_PAYOUT">PAYIN_TO_PAYOUT</option>
          <option value="PAYOUT_DEBIT">PAYOUT_DEBIT</option>
          <option value="AVAILABLE_DEPOSIT">AVAILABLE_DEPOSIT</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">账户类型</span>
        <select v-model="accountType" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm">
          <option value="">全部</option>
          <option value="payin">代收余额</option>
          <option value="available">可用余额</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">关键字</span>
        <input
          v-model.trim="keyword"
          type="text"
          placeholder="订单号 / 原因"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @keyup.enter="load"
        />
      </label>
      <div class="flex items-end gap-2 lg:col-span-2">
        <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50" @click="load">
          加载
        </button>
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50 disabled:opacity-50"
          :disabled="!filteredLogs.length"
          @click="exportCsv"
        >
          导出 CSV
        </button>
      </div>
    </div>

    <div class="grid gap-3 md:grid-cols-4">
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">流水条数</div>
        <div class="mt-1 text-lg font-semibold text-slate-900">{{ filteredLogs.length }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">入账合计</div>
        <div class="mt-1 text-lg font-semibold text-emerald-700">{{ formatAmount(summary.inflow) }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">出账合计</div>
        <div class="mt-1 text-lg font-semibold text-rose-700">{{ formatAmount(-summary.outflowAbs) }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">净变化</div>
        <div class="mt-1 text-lg font-semibold" :class="summary.net >= 0 ? 'text-emerald-700' : 'text-rose-700'">{{ formatAmount(summary.net) }}</div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1060px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">时间</th>
              <th class="whitespace-nowrap px-4 py-3">商户</th>
              <th class="whitespace-nowrap px-4 py-3">订单号</th>
              <th class="whitespace-nowrap px-4 py-3">类型</th>
              <th class="whitespace-nowrap px-4 py-3">账户类型</th>
              <th class="whitespace-nowrap px-4 py-3">变动金额</th>
              <th class="whitespace-nowrap px-4 py-3">变动前</th>
              <th class="whitespace-nowrap px-4 py-3">变动后</th>
              <th class="whitespace-nowrap px-4 py-3">原因</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="9">加载中…</td>
            </tr>
            <tr v-else-if="!filteredLogs.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="9">暂无资金流水</td>
            </tr>
            <tr v-for="x in pagedLogs" v-else :key="x.id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatTs(x.created_at) }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ x.merchant_id }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ x.order_no }}</td>
              <td class="px-4 py-3"><span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="changeTypeClass(x.change_type)">{{ x.change_type }}</span></td>
              <td class="px-4 py-3 text-slate-800">{{ accountTypeLabel(resolvedAccountType(x)) }}</td>
              <td class="px-4 py-3 font-semibold" :class="x.amount >= 0 ? 'text-emerald-700' : 'text-rose-700'">{{ formatAmount(x.amount) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatAmount(x.balance_before) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatAmount(x.balance_after) }}</td>
              <td class="px-4 py-3 text-slate-700">{{ x.reason || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <AdminPaginationBar
        v-if="!loading && filteredLogs.length > 0"
        v-model:page="page"
        v-model:pageSize="pageSize"
        :total="total"
        :page-count="pageCount"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { useClientPagination } from '../../../composables/useClientPagination'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet } from '../../../lib/adminApi'
import { formatAdminMoney, loadAdminDisplaySettings } from '../../../lib/displaySettings'

type SettlementLogItem = {
  id: number
  merchant_id: string
  order_no: string
  change_type: string
  account_type?: string
  amount: number
  balance_before: number
  balance_after: number
  reason: string
  created_at: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const merchantId = ref('')
const changeType = ref('')
const accountType = ref('')
const keyword = ref('')
const loading = ref(true)
const error = ref('')
const logs = ref<SettlementLogItem[]>([])
const toast = useUiToast()

const filteredLogs = computed(() => {
  const t = changeType.value.trim().toUpperCase()
  const at = accountType.value.trim()
  const k = keyword.value.trim().toLowerCase()
  return logs.value.filter((x) => {
    if (t && x.change_type.toUpperCase() !== t) return false
    if (at && resolvedAccountType(x) !== at) return false
    if (!k) return true
    return `${x.order_no} ${x.reason}`.toLowerCase().includes(k)
  })
})

function resolvedAccountType(x: SettlementLogItem): string {
  const a = (x.account_type || '').trim()
  if (a === 'available' || a === 'payin') return a
  return x.change_type === 'PAYOUT_DEBIT' || x.change_type === 'AVAILABLE_DEPOSIT' ? 'available' : 'payin'
}

function accountTypeLabel(code: string): string {
  if (code === 'available') return '可用余额'
  return '代收余额'
}

const summary = computed(() => {
  let inflow = 0
  let outflowAbs = 0
  for (const x of filteredLogs.value) {
    if (x.amount >= 0) inflow += x.amount
    else outflowAbs += Math.abs(x.amount)
  }
  return { inflow, outflowAbs, net: inflow - outflowAbs }
})

const { page, pageSize, total, pageCount, slice: pagedLogs } = useClientPagination(filteredLogs, 20)

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN', { hour12: false })
}

function formatAmount(cents: number): string {
  return formatAdminMoney(cents)
}

function changeTypeClass(t: string): string {
  if (t === 'ORDER_PAID') return 'bg-emerald-100 text-emerald-700'
  if (t === 'PAYIN_TO_PAYOUT') return 'bg-indigo-100 text-indigo-700'
  if (t === 'PAYOUT_DEBIT') return 'bg-rose-100 text-rose-700'
  if (t === 'AVAILABLE_DEPOSIT') return 'bg-sky-100 text-sky-800'
  return 'bg-slate-100 text-slate-700'
}

function csvEscape(v: string) {
  return `"${(v ?? '').replaceAll('"', '""')}"`
}

function exportCsv() {
  if (!filteredLogs.value.length) return
  const header = ['created_at', 'merchant_id', 'order_no', 'change_type', 'account_type', 'amount', 'balance_before', 'balance_after', 'reason']
  const rows = filteredLogs.value.map((x) => [
    formatTs(x.created_at),
    x.merchant_id,
    x.order_no,
    x.change_type,
    accountTypeLabel(resolvedAccountType(x)),
    String(x.amount),
    String(x.balance_before),
    String(x.balance_after),
    x.reason || '',
  ])
  const csv = [header, ...rows].map((r) => r.map(csvEscape).join(',')).join('\n')
  const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `settlement_logs_${Date.now()}.csv`
  a.click()
  URL.revokeObjectURL(url)
  toast.success('CSV 已导出')
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
    toast.error('资金流水加载失败')
  } finally {
    loading.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void loadAdminDisplaySettings().then(() => void load())
  if (registerRefresh)
    unregister = registerRefresh(() => {
      void loadAdminDisplaySettings().then(() => void load())
    })
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
