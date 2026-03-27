<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">提现申请列表</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">提现申请审核、驳回与手动打款确认。</p>
    </div>

    <div class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-2 lg:grid-cols-4">
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">商户（可选）</span>
        <input
          v-model.trim="merchantId"
          type="text"
          placeholder="merchant_id"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @keyup.enter="loadWithdrawals"
        />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">状态</span>
        <select
          v-model="statusFilter"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
        >
          <option value="">全部</option>
          <option value="0">待审核</option>
          <option value="1">已驳回</option>
          <option value="2">待打款</option>
          <option value="3">打款中</option>
          <option value="4">成功</option>
          <option value="5">失败</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">提现单号（模糊）</span>
        <input
          v-model.trim="withdrawNoKeyword"
          type="text"
          placeholder="输入单号片段筛选"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
        />
      </label>
      <div class="flex items-end gap-2">
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
          @click="loadWithdrawals"
        >
          加载
        </button>
      </div>
    </div>
    <p class="text-xs text-slate-500">列表为接口返回的最近 200 条；状态与单号筛选在本地对已加载数据生效。</p>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[860px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">提现单号</th>
              <th class="px-4 py-3">商户</th>
              <th class="px-4 py-3">申请金额</th>
              <th class="px-4 py-3">实付金额</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">申请时间</th>
              <th class="px-4 py-3">备注</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="withdrawLoading"><td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中...</td></tr>
            <tr v-else-if="!filteredWithdrawals.length">
              <td class="px-4 py-8 text-center text-slate-500" colspan="8">
                {{ withdrawals.length ? '当前筛选条件下暂无记录' : '暂无提现申请' }}
              </td>
            </tr>
            <template v-else>
              <tr v-for="w in filteredWithdrawals" :key="w.withdraw_no" class="hover:bg-slate-50/80">
                <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ w.withdraw_no }}</td>
                <td class="px-4 py-3 font-medium text-slate-900">{{ w.merchant_id }}</td>
                <td class="px-4 py-3 text-slate-700">
                  <div>{{ formatUsdt(w.apply_amount) }}</div>
                  <div v-if="w.fiat_debit_amount" class="text-[11px] text-slate-500">扣款 {{ formatFiat(w.fiat_debit_amount) }}</div>
                </td>
                <td class="px-4 py-3 text-slate-700">{{ formatUsdt(w.net_amount) }}</td>
                <td class="px-4 py-3"><span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="withdrawStatusClass(w.status)">{{ withdrawStatusText(w.status) }}</span></td>
                <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatTs(w.created_at) }}</td>
                <td class="px-4 py-3 text-slate-700">{{ w.apply_note || '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-wrap gap-2">
                    <button v-if="w.status === 0" type="button" class="rounded border border-emerald-200 bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-700" @click="reviewWithdrawal(w, true)">审核通过并扣款</button>
                    <button v-if="w.status === 0" type="button" class="rounded border border-rose-200 bg-rose-50 px-2 py-1 text-xs font-semibold text-rose-700" @click="reviewWithdrawal(w, false)">驳回</button>
                    <button v-if="w.status === 2 || w.status === 3" type="button" class="rounded border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs font-semibold text-indigo-700" @click="confirmPayout(w)">手动打款后确认完成</button>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPut } from '../../../lib/adminApi'

type WithdrawalItem = {
  withdraw_no: string
  merchant_id: string
  apply_amount: number
  fee_amount: number
  net_amount: number
  fiat_debit_amount?: number
  status: number
  receive_account: string
  receive_name: string
  bank_name: string
  apply_note: string
  created_at: number
}

type AdminDisplaySettings = { currency_symbol: string }

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const withdrawals = ref<WithdrawalItem[]>([])
const withdrawLoading = ref(false)
const merchantId = ref('')
const statusFilter = ref('')
const withdrawNoKeyword = ref('')

const filteredWithdrawals = computed(() => {
  let rows = withdrawals.value
  if (statusFilter.value !== '') {
    const n = Number(statusFilter.value)
    if (!Number.isNaN(n)) rows = rows.filter((w) => w.status === n)
  }
  const kw = withdrawNoKeyword.value.trim().toLowerCase()
  if (kw) rows = rows.filter((w) => (w.withdraw_no || '').toLowerCase().includes(kw))
  return rows
})
const fiatSymbol = ref('¥')
const toast = useUiToast()
const dialog = useUiDialog()

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN', { hour12: false })
}

function formatFiat(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const sym = fiatSymbol.value || '¥'
  return `${sign}${sym} ${(Math.abs(cents) / 100).toFixed(2)}`
}

function formatUsdt(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  return `${sign}${(Math.abs(cents) / 100).toFixed(2)} USDT`
}

function withdrawStatusText(s: number): string {
  if (s === 0) return '待审核'
  if (s === 1) return '已驳回'
  if (s === 2) return '待打款'
  if (s === 3) return '打款中'
  if (s === 4) return '成功'
  if (s === 5) return '失败'
  return '未知'
}

function withdrawStatusClass(s: number): string {
  if (s === 0) return 'bg-amber-100 text-amber-700'
  if (s === 1) return 'bg-rose-100 text-rose-700'
  if (s === 2 || s === 3) return 'bg-indigo-100 text-indigo-700'
  if (s === 4) return 'bg-emerald-100 text-emerald-700'
  if (s === 5) return 'bg-rose-100 text-rose-700'
  return 'bg-slate-100 text-slate-700'
}

async function loadDisplaySettings() {
  try {
    const ds = await adminGet<AdminDisplaySettings>('/v1/admin/display_settings')
    fiatSymbol.value = (ds.currency_symbol || '¥').trim() || '¥'
  } catch {
    fiatSymbol.value = '¥'
  }
}

async function loadWithdrawals() {
  withdrawLoading.value = true
  try {
    const q = new URLSearchParams()
    if (merchantId.value.trim()) q.set('merchant_id', merchantId.value.trim())
    q.set('limit', '200')
    const r = await adminGet<{ items: WithdrawalItem[] }>(`/v1/admin/settlement/withdrawals?${q.toString()}`)
    withdrawals.value = r.items ?? []
  } catch {
    withdrawals.value = []
    toast.error('提现申请加载失败')
  } finally {
    withdrawLoading.value = false
  }
}

async function reviewWithdrawal(w: WithdrawalItem, approved: boolean) {
  const ok = await dialog.confirm(approved ? `确认通过提现 ${w.withdraw_no} 并执行系统扣款吗？` : `确认驳回提现 ${w.withdraw_no} 吗？`, approved ? '审核确认' : '驳回确认')
  if (!ok) return
  try {
    await adminPut(`/v1/admin/settlement/withdrawals/${encodeURIComponent(w.withdraw_no)}/review`, { approved, review_note: '', operator: '' })
    toast.success(approved ? '审核通过，已扣款并进入待打款' : '已驳回提现申请')
    await loadWithdrawals()
  } catch (e) {
    toast.error(`审核处理失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

async function confirmPayout(w: WithdrawalItem) {
  const ok = await dialog.confirm(`请确认已在线下完成 USDT 转账，再点击“确认完成”。\n提现单号：${w.withdraw_no}`, '确认打款完成')
  if (!ok) return
  try {
    await adminPut(`/v1/admin/settlement/withdrawals/${encodeURIComponent(w.withdraw_no)}/payout`, { payout_note: '', operator: '' })
    toast.success('已确认打款完成')
    await loadWithdrawals()
  } catch (e) {
    toast.error(`确认打款失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void loadDisplaySettings()
  void loadWithdrawals()
  if (registerRefresh) unregister = registerRefresh(() => {
    void loadDisplaySettings()
    void loadWithdrawals()
  })
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
