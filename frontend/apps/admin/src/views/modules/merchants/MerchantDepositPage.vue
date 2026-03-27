<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">资金存入</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        向商户<strong>可用余额</strong>入账（法币分）。法币直接按金额入账；USDT 按系统展示汇率换算为法币后入账，并记入资金流水（<code class="rounded bg-slate-100 px-1 font-mono text-xs">AVAILABLE_DEPOSIT</code>）。
      </p>
    </div>

    <div class="mx-auto w-full max-w-xl space-y-6">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <MerchantIdPicker v-model="form.merchant_id" :merchants="allMerchants" />
        <p v-if="selectedMerchant" class="mt-2 text-xs text-slate-600">
          当前可用余额：<span class="font-semibold text-slate-900">{{ formatFiat(selectedMerchant.available_balance ?? 0) }}</span>
        </p>

        <div class="mt-4 grid gap-3">
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            存入方式
            <select v-model="form.mode" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <option value="fiat">法币（{{ fiatCode }}）</option>
              <option value="usdt">USDT（按汇率换算为法币入账）</option>
            </select>
          </label>

          <template v-if="form.mode === 'fiat'">
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              存入金额（{{ fiatCode }}）
              <input
                v-model.number="form.fiat_yuan"
                type="number"
                min="0"
                step="0.01"
                class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              />
            </label>
          </template>
          <template v-else>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              存入金额（USDT）
              <input
                v-model.number="form.usdt_yuan"
                type="number"
                min="0"
                step="0.01"
                class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              />
            </label>
            <div class="rounded-lg border border-indigo-100 bg-indigo-50/80 px-3 py-2 text-sm text-indigo-900">
              <div class="text-xs font-medium text-indigo-800">换算后入账（法币）</div>
              <div class="mt-0.5 font-mono font-semibold">{{ formatFiat(previewFiatCentsFromUsdt) }}</div>
              <p class="mt-1 text-[11px] text-indigo-700/90">汇率：1 USDT = {{ fiatToUsdtRate.toFixed(4) }} {{ fiatCode }}（与提现页一致：<code class="font-mono">floor(usdt分 × 汇率)</code> 法币分）</p>
            </div>
          </template>

          <label class="grid gap-1 text-xs font-medium text-slate-600">
            备注（可选）
            <input v-model.trim="form.note" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" placeholder="线下凭证号等" />
          </label>
        </div>

        <div class="mt-4 flex flex-wrap gap-2">
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
            :disabled="saving || !canSubmit"
            @click="submit"
          >
            {{ saving ? '提交中…' : '确认存入' }}
          </button>
          <RouterLink
            to="/settlement/logs"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50"
          >
            查看资金流水
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import MerchantIdPicker from '../../../components/MerchantIdPicker.vue'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPost } from '../../../lib/adminApi'

type AdminMerchantInfo = { merchant_id: string; payin_balance: number; available_balance: number }
type AdminDisplaySettings = { country_code: string; currency_code: string; currency_symbol: string; fiat_to_usdt_rate: number }

const route = useRoute()
const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const toast = useUiToast()
const saving = ref(false)
const fiatToUsdtRate = ref(7.2)
const fiatSymbol = ref('¥')
const fiatCode = ref('CNY')
const allMerchants = ref<AdminMerchantInfo[]>([])
const form = ref({
  merchant_id: '',
  mode: 'fiat' as 'fiat' | 'usdt',
  fiat_yuan: 0,
  usdt_yuan: 0,
  note: '',
})

const selectedMerchant = computed(() => {
  const id = form.value.merchant_id.trim()
  if (!id) return null
  return allMerchants.value.find((x) => x.merchant_id === id) ?? null
})

const usdtAmountCents = computed(() => Math.max(0, Math.round(form.value.usdt_yuan * 100)))

const previewFiatCentsFromUsdt = computed(() => {
  if (fiatToUsdtRate.value <= 0) return 0
  return Math.floor(usdtAmountCents.value * fiatToUsdtRate.value)
})

const fiatDepositCents = computed(() => Math.max(0, Math.round(form.value.fiat_yuan * 100)))

const canSubmit = computed(() => {
  if (!form.value.merchant_id.trim()) return false
  if (form.value.mode === 'fiat') return fiatDepositCents.value > 0
  return usdtAmountCents.value > 0 && previewFiatCentsFromUsdt.value > 0
})

function formatFiat(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const sym = fiatSymbol.value || '¥'
  return `${sign}${sym} ${(Math.abs(cents) / 100).toFixed(2)}`
}

function merchantIdFromRouteQuery(): string {
  const q = route.query.merchant_id
  if (typeof q === 'string') return q.trim()
  if (Array.isArray(q) && q[0]) return String(q[0]).trim()
  return ''
}

function applyMerchantFromRouteQuery() {
  const mid = merchantIdFromRouteQuery()
  if (mid) form.value.merchant_id = mid
}

async function loadMerchants() {
  try {
    const mr = await adminGet<{ merchants: AdminMerchantInfo[] }>('/v1/admin/merchants')
    allMerchants.value = mr.merchants ?? []
    applyMerchantFromRouteQuery()
  } catch {
    allMerchants.value = []
  }
}

watch(
  () => route.query.merchant_id,
  () => {
    applyMerchantFromRouteQuery()
  },
)

async function loadRates() {
  try {
    const ds = await adminGet<AdminDisplaySettings>('/v1/admin/display_settings')
    fiatToUsdtRate.value = ds.fiat_to_usdt_rate > 0 ? ds.fiat_to_usdt_rate : 7.2
    fiatSymbol.value = (ds.currency_symbol || '¥').trim() || '¥'
    fiatCode.value = (ds.currency_code || 'CNY').trim() || 'CNY'
  } catch {
    fiatToUsdtRate.value = 7.2
  }
}

async function submit() {
  if (!canSubmit.value) return
  const mid = form.value.merchant_id.trim()
  saving.value = true
  try {
    const body: Record<string, unknown> = {
      merchant_id: mid,
      mode: form.value.mode,
      note: form.value.note.trim(),
    }
    if (form.value.mode === 'fiat') {
      body.fiat_amount_cents = fiatDepositCents.value
    } else {
      body.usdt_amount_cents = usdtAmountCents.value
    }
    const r = await adminPost<{
      order_no: string
      available_balance: number
      fiat_credited_cents: number
      mode: string
    }>('/v1/admin/settlement/deposit', body)
    toast.success(`已入账，流水单号 ${r.order_no}，当前可用余额 ${formatFiat(r.available_balance)}`)
    form.value.fiat_yuan = 0
    form.value.usdt_yuan = 0
    form.value.note = ''
    await loadMerchants()
  } catch (e) {
    toast.error(`存入失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    saving.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  applyMerchantFromRouteQuery()
  void loadRates()
  void loadMerchants()
  if (registerRefresh)
    unregister = registerRefresh(() => {
      void loadRates()
      void loadMerchants()
    })
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
