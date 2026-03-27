<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">结算与提现</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">资金流水金额为系统配置的法币；提现按法币余额申请并换算为 USDT，审核后手动打款确认。</p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="flex flex-wrap border-b border-slate-200">
      <button
        type="button"
        class="relative -mb-px border-b-2 px-3 pb-3 text-sm font-semibold transition md:px-4"
        :class="activeTab === 'logs' ? 'border-slate-900 text-slate-900' : 'border-transparent text-slate-500 hover:text-slate-800'"
        @click="activeTab = 'logs'"
      >
        资金流水
      </button>
      <button
        type="button"
        class="relative -mb-px border-b-2 px-3 pb-3 text-sm font-semibold transition md:px-4"
        :class="activeTab === 'withdraw' ? 'border-slate-900 text-slate-900' : 'border-transparent text-slate-500 hover:text-slate-800'"
        @click="activeTab = 'withdraw'"
      >
        提现申请（phase2）
      </button>
    </div>

    <div v-show="activeTab === 'logs'" class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-2 lg:grid-cols-6">
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
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
          @click="load"
        >
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

    <div v-show="activeTab === 'logs'" class="grid gap-3 md:grid-cols-4">
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">流水条数</div>
        <div class="mt-1 text-lg font-semibold text-slate-900">{{ filteredLogs.length }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">入账合计</div>
        <div class="mt-1 text-lg font-semibold text-emerald-700">{{ formatLogMoney(summary.inflow) }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">出账合计</div>
        <div class="mt-1 text-lg font-semibold text-rose-700">{{ formatLogMoney(-summary.outflowAbs) }}</div>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <div class="text-xs text-slate-500">净变化</div>
        <div class="mt-1 text-lg font-semibold" :class="summary.net >= 0 ? 'text-emerald-700' : 'text-rose-700'">
          {{ formatLogMoney(summary.net) }}
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'logs'" class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
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
              <td class="px-4 py-3">
                <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="changeTypeClass(x.change_type)">
                  {{ x.change_type }}
                </span>
              </td>
              <td class="px-4 py-3 text-slate-800">{{ accountTypeLabel(resolvedAccountType(x)) }}</td>
              <td class="px-4 py-3 font-semibold" :class="x.amount >= 0 ? 'text-emerald-700' : 'text-rose-700'">{{ formatLogMoney(x.amount) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatLogMoney(x.balance_before) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatLogMoney(x.balance_after) }}</td>
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

    <div v-show="activeTab === 'withdraw'" class="grid gap-4 lg:grid-cols-12">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm lg:col-span-4">
        <div class="text-sm font-semibold text-slate-900">创建提现申请</div>
        <div class="mt-4 grid gap-3">
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            商户（可搜索）
            <input
              v-model.trim="withdrawForm.merchant_id"
              type="text"
              list="withdraw-merchant-options"
              placeholder="输入商户ID搜索"
              class="rounded-lg border border-slate-200 px-3 py-2 text-sm"
            />
            <datalist id="withdraw-merchant-options">
              <option v-for="m in allMerchants" :key="m.merchant_id" :value="m.merchant_id" />
            </datalist>
          </label>
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-3 text-[12px] text-slate-700">
            <div>当前代收余额：{{ formatFiat(merchantPayinBalance) }}</div>
            <div class="mt-1">当前可用余额：{{ formatFiat(merchantAvailableBalance) }}</div>
          </div>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            提现余额来源
            <select v-model="withdrawForm.balance_source" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <option value="available">可用余额</option>
              <option value="payin">代收余额</option>
            </select>
          </label>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            申请金额（{{ fiatCode }}，法币）
            <input v-model.number="withdrawForm.apply_fiat_yuan" type="number" min="0" step="0.01" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            <span class="text-[11px] text-slate-500">系统汇率：1 USDT = {{ fiatToUsdtRate.toFixed(4) }} {{ fiatCode }}（来自系统配置）</span>
            <span class="text-[11px] text-slate-500">可提现法币上限：{{ formatFiat(maxSourceFiatCents) }}（约合 {{ maxWithdrawUsdtText }}）</span>
          </label>
          <div class="rounded-lg border border-indigo-100 bg-indigo-50/80 px-3 py-2 text-[12px] text-indigo-900">
            <div class="font-medium">约合 USDT</div>
            <div class="mt-0.5 font-mono text-sm">{{ applyUsdtDisplay }}</div>
            <p v-if="applyUsdtCents <= 0 && withdrawForm.apply_fiat_yuan > 0" class="mt-1 text-[11px] text-amber-800">按当前汇率换算后 USDT 为 0，请提高法币金额。</p>
          </div>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            手续费（USDT）
            <input v-model.number="withdrawForm.fee_amount_yuan" type="number" min="0" step="0.01" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <div class="grid gap-1 text-xs font-medium text-slate-600">
            <span>USDT 收款地址（商户配置，只读）</span>
            <div
              class="rounded-lg border px-3 py-2 text-sm leading-relaxed"
              :class="merchantUsdtAddress ? 'border-slate-200 bg-slate-50 font-mono text-slate-800' : 'border-amber-200 bg-amber-50 text-amber-900'"
            >
              {{ merchantUsdtAddress || '未配置，请先在商户管理中填写 USDT 收款地址' }}
            </div>
          </div>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            链名称（如 TRC20）
            <input v-model.trim="withdrawForm.bank_name" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            申请备注
            <textarea v-model.trim="withdrawForm.apply_note" rows="3" class="rounded-lg border border-slate-200 px-3 py-2 text-sm"></textarea>
          </label>
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
            :disabled="withdrawSaving || !withdrawForm.merchant_id || !merchantUsdtAddress || withdrawForm.apply_fiat_yuan <= 0 || applyUsdtCents <= 0"
            @click="createWithdrawal"
          >
            {{ withdrawSaving ? '提交中...' : '提交申请' }}
          </button>
        </div>
      </div>

      <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm lg:col-span-8">
        <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3">
          <div class="text-sm font-semibold text-slate-900">提现申请列表（USDT 手动流程）</div>
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50"
            @click="loadWithdrawals"
          >
            刷新
          </button>
        </div>
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
              <tr v-if="withdrawLoading">
                <td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中...</td>
              </tr>
              <tr v-else-if="!withdrawals.length">
                <td class="px-4 py-8 text-center text-slate-500" colspan="8">暂无提现申请</td>
              </tr>
              <tr v-for="w in withdrawals" v-else :key="w.withdraw_no" class="hover:bg-slate-50/80">
                <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ w.withdraw_no }}</td>
                <td class="px-4 py-3 font-medium text-slate-900">{{ w.merchant_id }}</td>
                <td class="px-4 py-3 text-slate-700">
                  <div>{{ formatUsdt(w.apply_amount) }}</div>
                  <div v-if="w.fiat_debit_amount" class="text-[11px] text-slate-500">扣款 {{ formatFiat(w.fiat_debit_amount) }}</div>
                </td>
                <td class="px-4 py-3 text-slate-700">{{ formatUsdt(w.net_amount) }}</td>
                <td class="px-4 py-3">
                  <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="withdrawStatusClass(w.status)">
                    {{ withdrawStatusText(w.status) }}
                  </span>
                </td>
                <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatTs(w.created_at) }}</td>
                <td class="px-4 py-3 text-slate-700">{{ w.apply_note || '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-wrap gap-2">
                    <button
                      v-if="w.status === 0"
                      type="button"
                      class="rounded border border-emerald-200 bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-700"
                      @click="reviewWithdrawal(w, true)"
                    >
                      审核通过并扣款
                    </button>
                    <button
                      v-if="w.status === 0"
                      type="button"
                      class="rounded border border-rose-200 bg-rose-50 px-2 py-1 text-xs font-semibold text-rose-700"
                      @click="reviewWithdrawal(w, false)"
                    >
                      驳回
                    </button>
                    <button
                      v-if="w.status === 2 || w.status === 3"
                      type="button"
                      class="rounded border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs font-semibold text-indigo-700"
                      @click="confirmPayout(w)"
                    >
                      手动打款后确认完成
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">提现流程（下阶段）</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>提现申请单：申请金额、手续费、实付金额、收款地址信息</li>
        <li>平台审核：通过/驳回（通过时系统扣减可用余额）</li>
        <li>人工在链上完成 USDT 转账后，后台点击“确认完成”回写成功</li>
        <li>与对账中心差异批次联动</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { useClientPagination } from '../../../composables/useClientPagination'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'
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
type AdminMerchantInfo = {
  merchant_id: string
  payin_balance: number
  available_balance: number
  withdraw_usdt_address?: string
}
type AdminDisplaySettings = {
  country_code: string
  currency_code: string
  currency_symbol: string
  fiat_to_usdt_rate: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const activeTab = ref<'logs' | 'withdraw'>('logs')
const merchantId = ref('')
const changeType = ref('')
const accountType = ref('')
const keyword = ref('')
const loading = ref(true)
const error = ref('')
const logs = ref<SettlementLogItem[]>([])
const withdrawals = ref<WithdrawalItem[]>([])
const withdrawLoading = ref(false)
const withdrawSaving = ref(false)
const fiatToUsdtRate = ref(7.2)
const fiatSymbol = ref('¥')
const fiatCode = ref('CNY')
const merchantAvailableBalance = ref(0)
const toast = useUiToast()
const dialog = useUiDialog()
const withdrawForm = ref({
  merchant_id: '',
  balance_source: 'available' as 'available' | 'payin',
  apply_fiat_yuan: 0,
  fee_amount_yuan: 0,
  bank_name: '',
  apply_note: '',
})

const merchantUsdtAddress = computed(() => {
  const id = withdrawForm.value.merchant_id.trim()
  if (!id) return ''
  const row = allMerchants.value.find((x) => x.merchant_id === id)
  return (row?.withdraw_usdt_address || '').trim()
})
const allMerchants = ref<AdminMerchantInfo[]>([])
const merchantPayinBalance = ref(0)
const maxSourceFiatCents = computed(() =>
  withdrawForm.value.balance_source === 'payin' ? merchantPayinBalance.value : merchantAvailableBalance.value,
)

const maxWithdrawUsdtCents = computed(() => {
  if (fiatToUsdtRate.value <= 0) return 0
  return Math.floor(maxSourceFiatCents.value / fiatToUsdtRate.value)
})
const maxWithdrawUsdtText = computed(() => `${(maxWithdrawUsdtCents.value / 100).toFixed(2)} USDT`)

const applyFiatCents = computed(() => Math.round(withdrawForm.value.apply_fiat_yuan * 100))
const applyUsdtCents = computed(() => {
  if (fiatToUsdtRate.value <= 0) return 0
  return Math.floor(applyFiatCents.value / fiatToUsdtRate.value)
})
const applyUsdtDisplay = computed(() => {
  if (fiatToUsdtRate.value <= 0) return '—'
  if (applyUsdtCents.value <= 0) return '0.00 USDT'
  return `${(applyUsdtCents.value / 100).toFixed(2)} USDT`
})

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
  return Number.isNaN(d.getTime())
    ? '—'
    : d.toLocaleString('zh-CN', {
        hour12: false,
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
}

function formatLogMoney(cents: number): string {
  return formatAdminMoney(cents)
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

function changeTypeClass(t: string): string {
  if (t === 'ORDER_PAID') return 'bg-emerald-100 text-emerald-700'
  if (t === 'PAYIN_TO_PAYOUT') return 'bg-indigo-100 text-indigo-700'
  if (t === 'PAYOUT_DEBIT') return 'bg-rose-100 text-rose-700'
  if (t === 'AVAILABLE_DEPOSIT') return 'bg-sky-100 text-sky-800'
  return 'bg-slate-100 text-slate-700'
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

function csvEscape(v: string) {
  const s = (v ?? '').replaceAll('"', '""')
  return `"${s}"`
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

async function loadWithdrawals() {
  withdrawLoading.value = true
  try {
    const q = new URLSearchParams()
    if (merchantId.value) q.set('merchant_id', merchantId.value)
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

async function loadWithdrawContext() {
  const merchant = withdrawForm.value.merchant_id.trim()
  if (!merchant) {
    merchantPayinBalance.value = 0
    merchantAvailableBalance.value = 0
    return
  }
  const row = allMerchants.value.find((x) => x.merchant_id === merchant)
  merchantPayinBalance.value = row?.payin_balance ?? 0
  merchantAvailableBalance.value = row?.available_balance ?? 0
}

async function loadWithdrawBaseData() {
  try {
    const ds = await adminGet<AdminDisplaySettings>('/v1/admin/display_settings')
    fiatToUsdtRate.value = ds.fiat_to_usdt_rate > 0 ? ds.fiat_to_usdt_rate : 7.2
    fiatSymbol.value = (ds.currency_symbol || '¥').trim() || '¥'
    fiatCode.value = (ds.currency_code || 'CNY').trim() || 'CNY'
    const mr = await adminGet<{ merchants: AdminMerchantInfo[] }>('/v1/admin/merchants')
    allMerchants.value = mr.merchants ?? []
    await loadWithdrawContext()
  } catch {
    merchantPayinBalance.value = 0
    merchantAvailableBalance.value = 0
  }
}

async function createWithdrawal() {
  const applyAmount = applyUsdtCents.value
  const feeAmount = Math.floor(withdrawForm.value.fee_amount_yuan * 100)
  const addr = merchantUsdtAddress.value
  if (!addr) {
    toast.error('该商户未配置 USDT 收款地址')
    return
  }
  if (withdrawForm.value.apply_fiat_yuan <= 0) return
  if (applyAmount <= 0) {
    toast.error('按当前汇率换算后 USDT 为 0，请调整法币金额')
    return
  }
  if (feeAmount < 0 || feeAmount > applyAmount) {
    toast.error('手续费不能大于申请金额（USDT）')
    return
  }
  if (applyAmount > maxWithdrawUsdtCents.value) {
    toast.error(`超过可提现金额上限：${maxWithdrawUsdtText.value}（${formatFiat(maxSourceFiatCents.value)}）`)
    return
  }
  const ok = await dialog.confirm(
    `确认提交提现申请？\n商户：${withdrawForm.value.merchant_id.trim()}\n余额来源：${withdrawForm.value.balance_source === 'payin' ? '代收余额' : '可用余额'}\n申请法币：${formatFiat(applyFiatCents.value)}\n约合：${(applyAmount / 100).toFixed(2)} USDT\n收款地址：${addr}`,
    '提交确认',
  )
  if (!ok) return
  withdrawSaving.value = true
  try {
    await adminPost<{ item: WithdrawalItem }>('/v1/admin/settlement/withdrawals', {
      merchant_id: withdrawForm.value.merchant_id.trim(),
      balance_source: withdrawForm.value.balance_source,
      apply_amount: applyAmount,
      fee_amount: feeAmount,
      receive_account: addr,
      receive_name: '',
      bank_name: withdrawForm.value.bank_name.trim(),
      apply_note: withdrawForm.value.apply_note.trim(),
    })
    toast.success('提现申请已创建，待审核')
    withdrawForm.value = {
      merchant_id: withdrawForm.value.merchant_id,
      balance_source: withdrawForm.value.balance_source,
      apply_fiat_yuan: 0,
      fee_amount_yuan: 0,
      bank_name: '',
      apply_note: '',
    }
    await loadWithdrawals()
    await load()
    await loadWithdrawBaseData()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`创建提现申请失败：${msg}`)
  } finally {
    withdrawSaving.value = false
  }
}

async function reviewWithdrawal(w: WithdrawalItem, approved: boolean) {
  const ok = await dialog.confirm(
    approved
      ? `确认通过提现 ${w.withdraw_no} 并执行系统扣款吗？`
      : `确认驳回提现 ${w.withdraw_no} 吗？`,
    approved ? '审核确认' : '驳回确认',
  )
  if (!ok) return
  try {
    await adminPut(`/v1/admin/settlement/withdrawals/${encodeURIComponent(w.withdraw_no)}/review`, {
      approved,
      review_note: '',
      operator: '',
    })
    toast.success(approved ? '审核通过，已扣款并进入待打款' : '已驳回提现申请')
    await loadWithdrawals()
    await load()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`审核处理失败：${msg}`)
  }
}

async function confirmPayout(w: WithdrawalItem) {
  const ok = await dialog.confirm(
    `请确认已在线下完成 USDT 转账，再点击“确认完成”。\n提现单号：${w.withdraw_no}`,
    '确认打款完成',
  )
  if (!ok) return
  try {
    await adminPut(`/v1/admin/settlement/withdrawals/${encodeURIComponent(w.withdraw_no)}/payout`, {
      payout_note: '',
      operator: '',
    })
    toast.success('已确认打款完成')
    await loadWithdrawals()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`确认打款失败：${msg}`)
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void loadAdminDisplaySettings().then(() => void load())
  void loadWithdrawals()
  void loadWithdrawBaseData()
  if (registerRefresh)
    unregister = registerRefresh(() => {
      void loadAdminDisplaySettings().then(() => void load())
      void loadWithdrawals()
      void loadWithdrawBaseData()
    })
})
onUnmounted(() => {
  if (unregister) unregister()
})

watch(
  () => withdrawForm.value.merchant_id,
  () => {
    void loadWithdrawContext()
  },
)
</script>
