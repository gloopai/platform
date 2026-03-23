<template>
  <div class="space-y-8">
    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700 text-xs font-bold text-white">1</span>
        <h2 class="text-sm font-semibold text-slate-900">创建代收联调</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">
        调用 <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">/v1/payin/order</code> 创建订单并跳转收银台。
      </p>

      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">pay_type（支付产品）</span>
          <select v-model="payTypePreset" class="input-merchant">
            <option v-for="p in payProductOptions" :key="p.code" :value="p.code">{{ p.label }}（{{ p.code }}）</option>
            <option value="__custom__">自定义编码…</option>
          </select>
        </label>
        <label v-if="payTypePreset === '__custom__'" class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">自定义 pay_type</span>
          <input v-model.trim="payTypeCustom" class="input-merchant font-mono text-xs" placeholder="与后端 payin_products.code 一致" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">merchant_order_no</span>
          <input v-model.trim="merchantOrderNo" class="input-merchant font-mono text-xs" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">amount（分）</span>
          <input v-model.number="amount" type="number" min="1" class="input-merchant tabular-nums" />
        </label>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="btn-primary"
          :disabled="loading || !merchantId || !apiSecret || !merchantOrderNo || amount <= 0 || !resolvedPayType"
          @click="createOrder"
        >
          {{ loading ? '创建中…' : '创建订单' }}
        </button>
        <button type="button" class="btn-secondary" @click="regenOrderNo">重新生成订单号</button>
      </div>

      <div v-if="result" class="mt-4 rounded-2xl border border-slate-200 bg-slate-50/90 px-4 py-4 text-sm text-slate-900">
        <div class="font-mono text-xs">order_no: <span class="font-semibold">{{ result.order_no }}</span></div>
        <div class="mt-3 flex flex-wrap gap-3">
          <a class="doc-link" :href="result.checkout_url" target="_blank" rel="noreferrer">打开 checkout_url</a>
          <a class="doc-link" :href="localCheckoutUrl" target="_blank" rel="noreferrer">打开独立收银台</a>
        </div>
      </div>

      <div v-if="error" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{{ error }}</div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700 text-xs font-bold text-white">2</span>
        <h2 class="text-sm font-semibold text-slate-900">查询联调（代收/代付）</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">调用开放查询接口核对订单状态。</p>
      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">查询类型</span>
          <select v-model="queryMode" class="input-merchant">
            <option value="payin">代收（/v1/payin/query）</option>
            <option value="payout">代付（/v1/payout/query）</option>
          </select>
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-8">
          <span class="text-xs font-medium text-slate-600">order_no（平台订单号）</span>
          <input v-model.trim="queryOrderNo" class="input-merchant font-mono text-xs" />
        </label>
      </div>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button type="button" class="btn-primary" :disabled="queryLoading || !merchantId || !apiSecret || !queryOrderNo" @click="queryOpenOrder">
          {{ queryLoading ? '查询中…' : '查询订单' }}
        </button>
      </div>
      <pre v-if="queryResultText" class="mt-4 max-h-64 overflow-auto rounded-2xl border border-slate-200 bg-slate-50 p-4 font-mono text-xs leading-relaxed text-slate-900">{{ queryResultText }}</pre>
      <div v-if="queryError" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{{ queryError }}</div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-600 text-xs font-bold text-white">3</span>
        <h2 class="text-sm font-semibold text-slate-900">模拟上游回调</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">调用 <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">/v1/callback/notify</code>，验证支付成功回调。</p>

      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">channel_id</span>
          <input v-model.number="mockChannelId" type="number" min="1" class="input-merchant tabular-nums" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">paid_amount（分）</span>
          <input v-model.number="mockPaidAmount" type="number" min="1" class="input-merchant tabular-nums" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">upstream_trade_no</span>
          <input v-model.trim="mockUpstreamTradeNo" class="input-merchant font-mono text-xs" />
        </label>
        <label class="col-span-12 grid gap-1.5">
          <span class="text-xs font-medium text-slate-600">channel_sign_secret</span>
          <input v-model.trim="mockChannelSecret" class="input-merchant font-mono text-xs" />
        </label>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="btn-primary"
          :disabled="mockLoading || !result?.order_no || mockChannelId <= 0 || mockPaidAmount <= 0 || !mockUpstreamTradeNo || !mockChannelSecret"
          @click="mockNotify"
        >
          {{ mockLoading ? '回调中…' : '触发支付成功' }}
        </button>
        <button type="button" class="btn-secondary" @click="regenUpstreamTradeNo">重新生成上游单号</button>
      </div>

      <details class="mt-4 rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 text-sm text-slate-700">
        <summary class="cursor-pointer select-none font-medium text-slate-800">reason_code 快速说明</summary>
        <div class="mt-3 overflow-x-auto">
          <table class="min-w-[680px] text-left text-xs">
            <thead class="text-slate-500">
              <tr>
                <th class="px-2 py-1.5 font-semibold">reason_code</th>
                <th class="px-2 py-1.5 font-semibold">含义</th>
                <th class="px-2 py-1.5 font-semibold">处理建议</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 text-slate-700">
              <tr v-for="x in notifyReasonRows" :key="x.code">
                <td class="px-2 py-1.5 font-mono">{{ x.code }}</td>
                <td class="px-2 py-1.5">{{ x.meaning }}</td>
                <td class="px-2 py-1.5">{{ x.action }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </details>

      <div v-if="mockOk" class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-900">回调成功</div>
      <div v-if="mockError" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{{ mockError }}</div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-slate-900/[0.03] p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-800 text-xs font-bold text-white">4</span>
        <h2 class="text-sm font-semibold text-slate-900">签名工具</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">参数名排序后拼接，追加 <code class="rounded bg-white px-1.5 py-0.5 font-mono text-xs text-slate-800">key=secret</code> 再 MD5。</p>
      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">参数 JSON</span>
          <textarea v-model="signJson" rows="10" class="textarea-merchant" />
        </label>
        <div class="col-span-12 grid gap-3 md:col-span-6">
          <label class="grid gap-1.5">
            <span class="text-xs font-medium text-slate-600">secret</span>
            <input v-model.trim="signSecret" class="input-merchant font-mono text-xs" />
          </label>
          <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-inner">
            <div class="text-xs font-semibold text-slate-700">签名结果</div>
            <div class="mt-2 break-all font-mono text-xs leading-relaxed text-slate-900">{{ signOutput }}</div>
          </div>
          <div v-if="signError" class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{{ signError }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import md5 from 'blueimp-md5'
import { computed, ref, watch } from 'vue'
import { OPEN_API } from '@/api/endpoints'
import { NOTIFY_REASON_ROWS } from '@/config/notifyReasonCodes'
import { DEMO_PAY_PRODUCT_OPTIONS } from '@/config/payProducts'

type CreateOrderResp = {
  order_no: string
  status: number
  channel_id: number
  checkout_url: string
}

const props = defineProps<{
  merchantId: string
  apiSecret: string
  notifyUrl: string
}>()

const payProductOptions = DEMO_PAY_PRODUCT_OPTIONS
const payTypePreset = ref<(typeof DEMO_PAY_PRODUCT_OPTIONS)[number]['code'] | '__custom__'>('mock')
const payTypeCustom = ref('')
const resolvedPayType = computed(() => (payTypePreset.value === '__custom__' ? payTypeCustom.value.trim() : payTypePreset.value))

const merchantOrderNo = ref(`MO-${Date.now()}`)
const amount = ref(100)
const loading = ref(false)
const error = ref('')
const result = ref<CreateOrderResp | null>(null)

const queryMode = ref<'payin' | 'payout'>('payin')
const queryOrderNo = ref('')
const queryLoading = ref(false)
const queryError = ref('')
const queryResultText = ref('')

const mockChannelId = ref(1)
const mockPaidAmount = ref(amount.value)
const mockUpstreamTradeNo = ref(`UP-${Date.now()}`)
const mockChannelSecret = ref('channel_secret')
const mockLoading = ref(false)
const mockError = ref('')
const mockOk = ref(false)

const signJson = ref('')
const signSecret = ref(props.apiSecret)
const signError = ref('')
const notifyReasonRows = NOTIFY_REASON_ROWS

function md5Sign(params: Record<string, string>, secret: string): string {
  const keys = Object.keys(params).map((k) => k.toLowerCase()).filter((k) => k !== 'sign').sort()
  const parts: string[] = []
  for (const k of keys) {
    const v = params[k]
    if (!v) continue
    parts.push(`${k}=${v}`)
  }
  parts.push(`key=${secret}`)
  return md5(parts.join('&'))
}

function formatOpenApiError(status: number, bodyText: string): string {
  try {
    const j = JSON.parse(bodyText) as { code?: string; message?: string }
    if (j?.code) return `${j.code}: ${j.message ?? ''}`
  } catch {
  }
  return bodyText.trim() || `HTTP ${status}`
}

function regenOrderNo() {
  merchantOrderNo.value = `MO-${Date.now()}`
}

function regenUpstreamTradeNo() {
  mockUpstreamTradeNo.value = `UP-${Date.now()}`
}

async function createOrder() {
  loading.value = true
  error.value = ''
  result.value = null
  try {
    const params: Record<string, string> = {
      merchant_id: props.merchantId,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      payin_type: resolvedPayType.value,
      notify_url: props.notifyUrl,
    }
    const sign = md5Sign(params, props.apiSecret)
    const resp = await fetch(OPEN_API.payinOrder, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...params, sign }),
    })
    const text = await resp.text()
    if (!resp.ok) {
      error.value = `创建失败 — ${formatOpenApiError(resp.status, text)}`
      return
    }
    result.value = JSON.parse(text) as CreateOrderResp
    queryOrderNo.value = result.value.order_no
    mockChannelId.value = result.value.channel_id || mockChannelId.value
    mockPaidAmount.value = amount.value
    mockOk.value = false
    mockError.value = ''
  } catch {
    error.value = '网络错误'
  } finally {
    loading.value = false
  }
}

async function queryOpenOrder() {
  queryLoading.value = true
  queryError.value = ''
  queryResultText.value = ''
  try {
    const params: Record<string, string> = {
      merchant_id: props.merchantId,
      order_no: queryOrderNo.value.trim(),
      timestamp: String(Math.floor(Date.now() / 1000)),
    }
    const sign = md5Sign(params, props.apiSecret)
    const endpoint = queryMode.value === 'payout' ? OPEN_API.queryPayoutOrder : OPEN_API.queryPayinOrder
    const resp = await fetch(`${endpoint}?${new URLSearchParams({ ...params, sign }).toString()}`)
    const text = await resp.text()
    if (!resp.ok) {
      queryError.value = formatOpenApiError(resp.status, text)
      return
    }
    try {
      queryResultText.value = JSON.stringify(JSON.parse(text), null, 2)
    } catch {
      queryResultText.value = text
    }
  } catch {
    queryError.value = '网络错误'
  } finally {
    queryLoading.value = false
  }
}

async function mockNotify() {
  if (!result.value?.order_no) return
  mockLoading.value = true
  mockError.value = ''
  mockOk.value = false
  try {
    const params: Record<string, string> = {
      order_no: result.value.order_no,
      paid_amount: String(mockPaidAmount.value),
      upstream_trade_no: mockUpstreamTradeNo.value,
      channel_id: String(mockChannelId.value),
    }
    const sign = md5Sign(params, mockChannelSecret.value)
    const resp = await fetch(OPEN_API.callbackNotify, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...params, sign }),
    })
    if (!resp.ok) {
      mockError.value = `回调失败(${resp.status})`
      return
    }
    const data = (await resp.json()) as { ok: boolean; reason_code?: string; reason?: string }
    if (!data.ok) {
      mockError.value = `回调返回 ok=false: ${data.reason_code || 'UNKNOWN'}${data.reason ? ` (${data.reason})` : ''}`
      return
    }
    mockOk.value = true
  } catch {
    mockError.value = '网络错误'
  } finally {
    mockLoading.value = false
  }
}

const localCheckoutUrl = computed(() => (result.value?.order_no ? `http://127.0.0.1:5174/?order_no=${encodeURIComponent(result.value.order_no)}` : ''))
const signOutput = computed(() => {
  signError.value = ''
  try {
    const obj = JSON.parse(signJson.value) as Record<string, unknown>
    const params: Record<string, string> = {}
    for (const [k, v] of Object.entries(obj)) {
      if (v === null || v === undefined) continue
      params[String(k).toLowerCase()] = String(v)
    }
    return md5Sign(params, signSecret.value)
  } catch {
    signError.value = '参数 JSON 解析失败'
    return ''
  }
})

function syncSignJsonFromForm() {
  signJson.value = JSON.stringify(
    {
      merchant_id: props.merchantId,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      payin_type: resolvedPayType.value,
    },
    null,
    2,
  )
}

syncSignJsonFromForm()
watch([() => props.merchantId, () => props.apiSecret, merchantOrderNo, amount, payTypePreset, payTypeCustom], () => {
  signSecret.value = props.apiSecret
  syncSignJsonFromForm()
})
</script>

<style scoped>
.input-merchant {
  @apply w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
.textarea-merchant {
  @apply w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5 font-mono text-xs text-slate-900 shadow-inner focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
.btn-primary {
  @apply rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40;
}
.btn-secondary {
  @apply rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50;
}
.doc-link {
  @apply font-semibold text-slate-800 underline decoration-slate-400/90 underline-offset-2 hover:text-slate-950;
}
</style>

