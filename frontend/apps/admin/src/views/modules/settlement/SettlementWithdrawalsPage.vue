<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">提现申请</h1>
        <p class="mt-1 max-w-3xl text-sm text-slate-600">
          使用法币余额发起提现，按系统配置汇率换算为 USDT；USDT 收款地址取自商户资料，请在商户管理中维护。
        </p>
      </div>
      <RouterLink
        to="/settlement/withdrawals/list"
        class="shrink-0 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm hover:bg-slate-50"
      >
        提现申请列表
      </RouterLink>
    </div>

    <div class="mx-auto w-full space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">说明</div>
      <p class="mt-2 text-sm leading-relaxed text-slate-600">
        选择商户与<strong>提现余额来源</strong>后，输入<strong>申请金额（法币）</strong>；系统将按汇率换算为 USDT 并写入提现单。手续费按 USDT 填写；打款前请在列表中完成审核与确认。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-800">发起提现申请</div>

      <div class="mt-4 grid gap-3 sm:grid-cols-2">
        <div class="sm:col-span-2">
          <MerchantIdPicker v-model="withdrawForm.merchant_id" :merchants="allMerchants" />
        </div>

        <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
          提现余额来源
          <select v-model="withdrawForm.balance_source" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
            <option value="available">可用余额（{{ formatFiat(merchantAvailableBalance) }}）</option>
            <option value="payin">代收余额（{{ formatFiat(merchantPayinBalance) }}）</option>
          </select>
        </label>

        <label class="grid gap-1 text-xs font-medium text-slate-600">
          申请金额（{{ fiatCode }}）
          <input
            v-model.number="withdrawForm.apply_fiat_yuan"
            type="number"
            min="0"
            step="0.01"
            class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
          />
        </label>

        <div class="rounded-lg border border-slate-200 bg-slate-50/90 px-3 py-2">
          <div class="text-xs font-medium text-slate-600">约合 USDT</div>
          <div class="mt-1 font-mono text-sm font-semibold text-slate-900">{{ applyUsdtDisplay }}</div>
          <p v-if="applyUsdtCents <= 0 && withdrawForm.apply_fiat_yuan > 0" class="mt-1 text-xs text-amber-800">
            不足 0.01 USDT，请提高法币金额。
          </p>
        </div>

        <p class="text-xs leading-relaxed text-slate-500 sm:col-span-2">
          汇率：1 USDT = {{ fiatToUsdtRate.toFixed(4) }} {{ fiatCode }}（系统配置）。可提上限：{{ formatFiat(maxSourceFiatCents) }}（约 {{ maxWithdrawUsdtText }}）。
        </p>

        <label class="grid gap-1 text-xs font-medium text-slate-600">
          手续费（USDT）
          <input
            v-model.number="withdrawForm.fee_amount_yuan"
            type="number"
            min="0"
            step="0.01"
            class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
          />
        </label>

        <label class="grid gap-1 text-xs font-medium text-slate-600">
          网络 / 链（如 TRC20）
          <input v-model.trim="withdrawForm.bank_name" type="text" placeholder="TRC20" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </label>

        <div class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
          <span>USDT 收款地址（商户配置，只读）</span>
          <div
            class="flex flex-col gap-2 rounded-lg border px-3 py-2 sm:flex-row sm:items-start sm:justify-between"
            :class="merchantUsdtAddress ? 'border-slate-200 bg-slate-50' : 'border-amber-200 bg-amber-50'"
          >
            <p class="min-w-0 flex-1 break-all font-mono text-sm text-slate-800">
              {{ merchantUsdtAddress || '未配置，请先在商户管理中填写 USDT 收款地址。' }}
            </p>
            <button
              v-if="merchantUsdtAddress"
              type="button"
              class="shrink-0 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50"
              @click="copyUsdtAddress"
            >
              复制
            </button>
          </div>
        </div>

        <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
          备注（可选）
          <textarea
            v-model.trim="withdrawForm.apply_note"
            rows="3"
            placeholder="补充说明"
            class="resize-y rounded-lg border border-slate-200 px-3 py-2 text-sm"
          />
        </label>
      </div>

      <p v-if="!canSubmit && withdrawForm.merchant_id" class="mt-3 text-xs text-amber-800">
        {{ submitBlockedReason }}
      </p>

      <div class="mt-4 flex flex-wrap gap-2">
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
          :disabled="withdrawSaving || !canSubmit"
          @click="createWithdrawal"
        >
          {{ withdrawSaving ? '提交中…' : '提交申请' }}
        </button>
      </div>
    </div>

    <div class="rounded-xl border border-slate-100 bg-slate-50/80 p-4 text-xs text-slate-600">
      <div class="font-semibold text-slate-700">提示</div>
      <ul class="mt-2 list-disc space-y-1 pl-4 leading-relaxed">
        <li>收款地址与商户管理中「USDT 收款地址」一致，此处不可编辑。</li>
        <li>审核、打款确认请在「提现申请列表」中操作。</li>
      </ul>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import MerchantIdPicker from '../../../components/MerchantIdPicker.vue'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPost } from '../../../lib/adminApi'

type AdminMerchantInfo = { merchant_id: string; payin_balance: number; available_balance: number; withdraw_usdt_address?: string }
type AdminDisplaySettings = { country_code: string; currency_code: string; currency_symbol: string; fiat_to_usdt_rate: number }

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const withdrawSaving = ref(false)
const fiatToUsdtRate = ref(7.2)
const fiatSymbol = ref('¥')
const fiatCode = ref('CNY')
const merchantAvailableBalance = ref(0)
const merchantPayinBalance = ref(0)
const allMerchants = ref<AdminMerchantInfo[]>([])
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

const canSubmit = computed(() => {
  if (!withdrawForm.value.merchant_id.trim()) return false
  if (!merchantUsdtAddress.value) return false
  if (withdrawForm.value.apply_fiat_yuan <= 0) return false
  if (applyUsdtCents.value <= 0) return false
  return true
})

const submitBlockedReason = computed(() => {
  if (!withdrawForm.value.merchant_id.trim()) return ''
  if (!merchantUsdtAddress.value) return '请先在商户管理中配置该商户的 USDT 收款地址。'
  if (withdrawForm.value.apply_fiat_yuan <= 0) return '请输入申请金额。'
  if (applyUsdtCents.value <= 0) return '换算后 USDT 为 0，请调整金额。'
  return ''
})

function formatFiat(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const sym = fiatSymbol.value || '¥'
  return `${sign}${sym} ${(Math.abs(cents) / 100).toFixed(2)}`
}

async function copyUsdtAddress() {
  const addr = merchantUsdtAddress.value
  if (!addr) return
  try {
    await navigator.clipboard.writeText(addr)
    toast.success('已复制收款地址')
  } catch {
    toast.error('复制失败，请手动复制')
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
  if (!addr) return toast.error('该商户未配置 USDT 收款地址')
  if (withdrawForm.value.apply_fiat_yuan <= 0) return
  if (applyAmount <= 0) return toast.error('按当前汇率换算后 USDT 为 0，请调整法币金额')
  if (feeAmount < 0 || feeAmount > applyAmount) return toast.error('手续费不能大于申请金额（USDT）')
  if (applyAmount > maxWithdrawUsdtCents.value)
    return toast.error(`超过可提现金额上限：${maxWithdrawUsdtText.value}（${formatFiat(maxSourceFiatCents.value)}）`)
  const ok = await dialog.confirm(
    `确认提交提现申请？\n商户：${withdrawForm.value.merchant_id.trim()}\n余额来源：${withdrawForm.value.balance_source === 'payin' ? '代收余额' : '可用余额'}\n申请法币：${formatFiat(applyFiatCents.value)}\n约合：${(applyAmount / 100).toFixed(2)} USDT\n收款地址：${addr}`,
    '提交确认',
  )
  if (!ok) return
  withdrawSaving.value = true
  try {
    await adminPost('/v1/admin/settlement/withdrawals', {
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
    await loadWithdrawBaseData()
  } catch (e) {
    toast.error(`创建提现申请失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    withdrawSaving.value = false
  }
}

watch(() => withdrawForm.value.merchant_id, () => {
  void loadWithdrawContext()
})

let unregister: (() => void) | null = null
onMounted(() => {
  void loadWithdrawBaseData()
  if (registerRefresh) unregister = registerRefresh(() => {
    void loadWithdrawBaseData()
  })
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
