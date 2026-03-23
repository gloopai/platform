<template>
  <div class="space-y-8">
    <PageHeader title="控制台" description="今日经营概况与账户状态一目了然" />

    <div class="space-y-4">
      <div class="grid gap-4 md:grid-cols-2">
        <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
          <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-400 opacity-15 blur-2xl transition group-hover:opacity-30" />
          <div class="relative">
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
                <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                </span>
                代收余额
              </div>
              <button
                type="button"
                class="rounded-lg border border-slate-200 bg-white px-2 py-1 text-[11px] font-semibold text-slate-700 hover:border-slate-300"
                @click="openTransferDialog"
              >
                划转
              </button>
            </div>
            <div class="mt-3 text-3xl font-semibold tabular-nums tracking-tight text-slate-900">{{ collectBalanceText }}</div>
            <div class="mt-2 text-xs text-slate-500">可转入代付账户用于下发</div>
          </div>
        </div>

        <div class="group relative overflow-hidden rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm transition hover:border-slate-300/90 hover:shadow-md">
          <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-300 opacity-20 blur-2xl transition group-hover:opacity-35" />
          <div class="relative">
            <div class="flex items-center gap-2 text-xs font-medium text-slate-500">
              <span class="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                </svg>
              </span>
              代付余额
            </div>
            <div class="mt-3 text-3xl font-semibold tabular-nums tracking-tight text-slate-900">{{ payoutBalanceText }}</div>
            <div class="mt-2 text-xs text-slate-500">提交代付订单时优先扣减</div>
          </div>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="border-b border-slate-100 bg-slate-50/80 px-5 py-3">
            <div class="text-sm font-semibold text-slate-900">代收概览</div>
            <div class="mt-0.5 text-xs text-slate-500">按今日代收数据汇总</div>
          </div>
          <div class="grid gap-3 px-5 py-4 sm:grid-cols-2 xl:grid-cols-4">
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">代收笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-slate-900">{{ orderCountText }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">成功笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-emerald-700">{{ collectPaidCountText }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">失败笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-rose-700">{{ collectFailedCountText }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">成功率 / 金额</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-slate-900">{{ successRateText }}</div>
              <div class="mt-1 text-xs tabular-nums text-slate-500">{{ todayAmountText }}</div>
            </div>
          </div>
        </div>

        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="border-b border-slate-100 bg-slate-50/80 px-5 py-3">
            <div class="text-sm font-semibold text-slate-900">代付概览</div>
            <div class="mt-0.5 text-xs text-slate-500">按最近 200 笔代付订单汇总</div>
          </div>
          <div class="grid gap-3 px-5 py-4 sm:grid-cols-2 xl:grid-cols-4">
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">代付笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-slate-900">{{ payoutOverview.count }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">成功笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-emerald-700">{{ payoutOverview.successCount }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">失败笔数</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-rose-700">{{ payoutOverview.failedCount }}</div>
            </div>
            <div class="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
              <div class="text-xs text-slate-500">成功率 / 金额</div>
              <div class="mt-1 text-lg font-semibold tabular-nums text-slate-900">{{ payoutSuccessRateText }}</div>
              <div class="mt-1 text-xs tabular-nums text-slate-500">{{ payoutAmountText }}</div>
            </div>
          </div>
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

    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="transferDialogOpen" class="fixed inset-0 z-[1000] bg-slate-900/50 p-4 backdrop-blur-sm">
          <div class="flex h-full items-center justify-center">
            <Transition
              enter-active-class="transition duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-[0.98]"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-[0.98]"
            >
              <div v-if="transferDialogOpen" class="w-full max-w-md rounded-2xl border border-slate-200 bg-white shadow-xl">
                <div class="border-b border-slate-200 px-5 py-4">
                  <div class="text-base font-semibold text-slate-900">余额划转（代收 -> 代付）</div>
                </div>
                <div class="space-y-3 px-5 py-4">
                  <div class="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
                    <div>当前代收：{{ collectBalanceText }}</div>
                    <div class="mt-1">当前代付：{{ payoutBalanceText }}</div>
                  </div>
                  <label class="grid gap-1">
                    <span class="text-xs text-slate-500">划转金额（{{ transferCurrencyCode }}）</span>
                    <input v-model.number="transferAmount" type="number" min="1" step="1" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
                  </label>
                  <p v-if="transferMsg" class="text-xs text-slate-600">{{ transferMsg }}</p>
                </div>
                <div class="flex justify-end gap-2 border-t border-slate-200 px-5 py-4">
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
                    :disabled="transferLoading"
                    @click="closeTransferDialog"
                  >
                    取消
                  </button>
                  <button
                    type="button"
                    class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
                    :disabled="transferLoading || transferAmount <= 0"
                    @click="submitTransfer"
                  >
                    {{ transferLoading ? '划转中…' : '确认划转' }}
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { transferCollectToPayout } from '@/api/finance'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { useMerchantSummary } from '@/composables/useMerchantSummary'
import { merchantDisplaySettings } from '@/lib/displaySettings'
import { formatYuanLabel } from '@/utils/format'

const { summary, byProduct, payoutOverview, error, loading, load } = useMerchantSummary()

const todayAmountText = computed(() => formatYuanLabel(summary.value?.today_amount ?? 0))

const collectBalanceText = computed(() => formatYuanLabel(summary.value?.collect_balance ?? summary.value?.balance ?? 0))
const payoutBalanceText = computed(() => formatYuanLabel(summary.value?.payout_balance ?? 0))
const transferCurrencyCode = computed(() => merchantDisplaySettings.value.currency_code || 'CNY')
const transferAmount = ref(0)
const transferLoading = ref(false)
const transferMsg = ref('')
const transferDialogOpen = ref(false)
const payoutSuccessRateText = computed(() => `${(payoutOverview.value.successRate * 100).toFixed(2)}%`)
const payoutAmountText = computed(() => formatYuanLabel(payoutOverview.value.amount))

const successRateText = computed(() => {
  const v = summary.value?.success_rate
  if (v === undefined || v === null) return '—'
  return `${(v * 100).toFixed(2)}%`
})
const collectPaidCountText = computed(() => {
  let total = 0
  for (const x of byProduct.value?.items || []) total += x.paid_count || 0
  return String(total)
})
const collectFailedCountText = computed(() => {
  let total = 0
  for (const x of byProduct.value?.items || []) total += x.failed_count || 0
  return String(total)
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

async function submitTransfer() {
  if (transferAmount.value <= 0) return
  transferLoading.value = true
  transferMsg.value = ''
  try {
    const amountCent = Math.floor(transferAmount.value) * 100
    const r = await transferCollectToPayout(amountCent)
    transferMsg.value = `划转成功：代收 ${formatYuanLabel(r.collect_balance)}，代付 ${formatYuanLabel(r.payout_balance)}`
    transferAmount.value = 0
    await load()
    closeTransferDialog()
  } catch {
    transferMsg.value = '划转失败：请确认代收余额是否充足。'
  } finally {
    transferLoading.value = false
  }
}

function openTransferDialog() {
  transferDialogOpen.value = true
  transferAmount.value = 0
  transferMsg.value = ''
}

function closeTransferDialog() {
  transferDialogOpen.value = false
  transferAmount.value = 0
  transferMsg.value = ''
}
</script>
