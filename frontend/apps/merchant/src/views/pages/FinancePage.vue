<template>
  <div class="space-y-8">
    <PageHeader title="财务中心" description="资金流水、提现与对账（部分能力为占位）" />

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <div class="text-sm font-semibold text-slate-900">结算与流水</div>
          <p class="mt-1 text-xs text-slate-500">展示余额变更记录（fund_logs）</p>
        </div>
        <div class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-700">
          代收余额：<span class="font-semibold tabular-nums text-slate-900">{{ formatAmount(summary?.payin_balance ?? 0) }}</span>
          <span class="mx-2 text-slate-400">|</span>
          可用余额：<span class="font-semibold tabular-nums text-slate-900">{{ formatAmount(summary?.available_balance ?? 0) }}</span>
        </div>
      </div>
      <div class="mt-4 overflow-hidden rounded-xl border border-slate-100">
        <div class="overflow-x-auto">
          <table class="w-full min-w-[760px] text-left text-sm">
            <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th class="whitespace-nowrap px-4 py-3">时间</th>
                <th class="whitespace-nowrap px-4 py-3">类型</th>
                <th class="whitespace-nowrap px-4 py-3">账户类型</th>
                <th class="whitespace-nowrap px-4 py-3">订单号</th>
                <th class="whitespace-nowrap px-4 py-3">变更</th>
                <th class="whitespace-nowrap px-4 py-3">余额</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td class="px-4 py-8 text-center text-slate-500" colspan="6">加载中…</td>
              </tr>
              <tr v-else-if="logs.length === 0">
                <td class="px-4 py-12 text-center text-slate-500" colspan="6">暂无流水记录</td>
              </tr>
              <tr v-for="l in pagedLogs" :key="l.id" class="transition hover:bg-slate-50/80">
                <td class="whitespace-nowrap px-4 py-3 text-slate-700">{{ formatTime(l.created_at) }}</td>
                <td class="px-4 py-3 text-slate-800">{{ l.change_type }}</td>
                <td class="px-4 py-3 text-slate-800">{{ accountTypeLabel(l) }}</td>
                <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ l.order_no || '—' }}</td>
                <td class="px-4 py-3 tabular-nums text-slate-800">{{ formatAmount(l.amount) }}</td>
                <td class="px-4 py-3 tabular-nums font-medium text-slate-900">{{ formatAmount(l.balance_after) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <MerchantPaginationBar
          v-if="!loading && logs.length > 0"
          v-model:page="page"
          v-model:pageSize="pageSize"
          :total="total"
          :page-count="pageCount"
        />
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
            <div class="text-sm font-semibold text-slate-900">提现申请（代收余额）</div>
            <p class="mt-1 text-sm text-slate-600">提现仅使用代收余额，可用余额不可提现（MVP 规则）</p>
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
            <div class="text-sm font-semibold text-slate-900">代收划转到代付</div>
            <p class="mt-1 text-sm text-slate-600">用于补足可用余额；提交代付订单前系统将校验可用余额</p>
            <div class="mt-4 flex flex-wrap items-end gap-2">
              <label class="grid gap-1.5">
                <span class="text-xs text-slate-500">划转金额（{{ transferCurrencyCode }}）</span>
                <input v-model.number="transferAmount" type="number" min="1" step="1" class="input-merchant w-40 tabular-nums" />
              </label>
              <button
                type="button"
                class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="transferLoading || transferAmount <= 0"
                @click="submitTransfer"
              >
                {{ transferLoading ? '划转中…' : '确认划转' }}
              </button>
            </div>
            <p v-if="transferMessage" class="mt-2 text-xs text-slate-600">{{ transferMessage }}</p>
          </div>
        </div>
      </div>
    </div>

    <ErrorCallout v-if="error" :message="error" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import MerchantPaginationBar from '@/components/ui/MerchantPaginationBar.vue'
import { useClientPagination } from '@/composables/useClientPagination'
import { fetchMerchantFundLogs, transferPayinToPayout } from '@/api/finance'
import { fetchMerchantSummary } from '@/api/console'
import { merchantDisplaySettings } from '@/lib/displaySettings'
import type { MerchantFundLogItem } from '@/types/merchant.api'
import { formatUnixSeconds, formatYuanLabel } from '@/utils/format'

const logs = ref<MerchantFundLogItem[]>([])
const { page, pageSize, total, pageCount, slice: pagedLogs } = useClientPagination(logs, 20)
const loading = ref(false)
const error = ref('')
const summary = ref<{ payin_balance?: number; available_balance?: number } | null>(null)
const transferCurrencyCode = computed(() => merchantDisplaySettings.value.currency_code || 'CNY')
const transferAmount = ref(0)
const transferLoading = ref(false)
const transferMessage = ref('')

function formatAmount(v: number) {
  return formatYuanLabel(v)
}

function formatTime(ts: number) {
  return formatUnixSeconds(ts)
}

function accountTypeLabel(l: MerchantFundLogItem): string {
  const a = (l.account_type || '').trim()
  if (a === 'available') return '可用余额'
  if (a === 'payin') return '代收余额'
  return l.change_type === 'PAYOUT_DEBIT' ? '可用余额' : '代收余额'
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    page.value = 1
    const [res, sum] = await Promise.all([fetchMerchantFundLogs(200), fetchMerchantSummary()])
    logs.value = res.logs || []
    summary.value = sum
    transferMessage.value = ''
  } catch {
    error.value = '加载失败：请确认已登录且网关已启动。'
  } finally {
    loading.value = false
  }
}

async function submitTransfer() {
  if (transferAmount.value <= 0) return
  transferLoading.value = true
  transferMessage.value = ''
  try {
    const amountCent = Math.floor(transferAmount.value) * 100
    const resp = await transferPayinToPayout(amountCent)
    transferMessage.value = `划转成功：代收余额 ${formatAmount(resp.payin_balance)}，可用余额 ${formatAmount(resp.available_balance)}`
    transferAmount.value = 0
    await reload()
  } catch {
    transferMessage.value = '划转失败：请确认代收余额充足。'
  } finally {
    transferLoading.value = false
  }
}

onMounted(() => {
  void reload()
})
</script>

<style scoped>
.input-merchant {
  @apply rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
</style>
