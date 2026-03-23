<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">开发配置</h1>
      <p class="mt-1 text-sm text-slate-600">管理接入参数、联调下单与签名调试</p>
    </div>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-900 text-xs font-bold text-white">1</span>
        <h2 class="text-sm font-semibold text-slate-900">参数配置</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">将用于本地保存与请求签名，修改后会写入浏览器存储。</p>
      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">merchant_id</span>
          <input v-model.trim="merchantId" class="input-merchant" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">app_secret</span>
          <input v-model.trim="apiSecret" class="input-merchant" type="password" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">IP 白名单</span>
          <input v-model.trim="ipWhitelist" class="input-merchant" placeholder="例如：127.0.0.1,10.0.0.0/24" />
        </label>
        <label class="col-span-12 grid gap-1.5 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">Notify URL</span>
          <input v-model.trim="notifyUrl" class="input-merchant" placeholder="https://merchant.example.com/notify" />
        </label>
      </div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700 text-xs font-bold text-white">2</span>
        <h2 class="text-sm font-semibold text-slate-900">下单联调</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">
        调用 <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-800">/v1/pay/order</code> 创建订单并跳转收银台。请求体中的
        <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">pay_type</code> 表示<strong>支付产品编码</strong>（如 mock、wechat），平台在内部再路由到具体上游通道；商户侧<strong>不要</strong>传内部
        <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">channel_id</code> 参与选路。
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
          <input v-model.trim="payTypeCustom" class="input-merchant font-mono text-xs" placeholder="与后端 pay_products.code 一致" />
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
          class="rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="loading || !merchantId || !apiSecret || !merchantOrderNo || amount <= 0 || !resolvedPayType"
          @click="createOrder"
        >
          {{ loading ? '创建中…' : '创建订单' }}
        </button>
        <button
          type="button"
          class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
          @click="regenOrderNo"
        >
          重新生成订单号
        </button>
      </div>

      <div v-if="result" class="mt-4 rounded-2xl border border-slate-200 bg-slate-50/90 px-4 py-4 text-sm text-slate-900">
        <div class="font-mono text-xs">
          order_no: <span class="font-semibold">{{ result.order_no }}</span>
        </div>
        <div class="mt-3 flex flex-wrap gap-3">
          <a class="font-semibold text-slate-800 underline decoration-slate-400/90 underline-offset-2 hover:text-slate-950" :href="result.checkout_url" target="_blank" rel="noreferrer">打开 checkout_url</a>
          <a class="font-semibold text-slate-800 underline decoration-slate-400/90 underline-offset-2 hover:text-slate-950" :href="localCheckoutUrl" target="_blank" rel="noreferrer">打开独立收银台</a>
        </div>
      </div>

      <div v-if="error" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
        {{ error }}
      </div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700 text-xs font-bold text-white">2b</span>
        <h2 class="text-sm font-semibold text-slate-900">查单联调（开放接口）</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">
        <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">GET /v1/pay/query</code>，参数与签名规则同开放平台；用于核对订单状态（MVP 仅展示原始 JSON）。
      </p>
      <div class="mt-6 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1.5 md:col-span-8">
          <span class="text-xs font-medium text-slate-600">order_no（平台订单号）</span>
          <input v-model.trim="queryOrderNo" class="input-merchant font-mono text-xs" placeholder="创建订单返回的 order_no" />
        </label>
      </div>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="queryLoading || !merchantId || !apiSecret || !queryOrderNo"
          @click="queryOpenOrder"
        >
          {{ queryLoading ? '查询中…' : '查询订单' }}
        </button>
      </div>
      <pre
        v-if="queryResultText"
        class="mt-4 max-h-64 overflow-auto rounded-2xl border border-slate-200 bg-slate-50 p-4 font-mono text-xs leading-relaxed text-slate-900"
        >{{ queryResultText }}</pre
      >
      <div v-if="queryError" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
        {{ queryError }}
      </div>
    </section>

    <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-slate-600 text-xs font-bold text-white">3</span>
        <h2 class="text-sm font-semibold text-slate-900">模拟上游回调</h2>
      </div>
      <p class="mt-2 text-sm text-slate-600">调用 <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs">/v1/callback/notify</code>，验证支付成功、入账与异步通知。</p>

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
          <input v-model.trim="mockChannelSecret" class="input-merchant font-mono text-xs" placeholder="默认 seed_demo.sql 为 channel_secret" />
        </label>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="mockLoading || !result?.order_no || mockChannelId <= 0 || mockPaidAmount <= 0 || !mockUpstreamTradeNo || !mockChannelSecret"
          @click="mockNotify"
        >
          {{ mockLoading ? '回调中…' : '触发支付成功' }}
        </button>
        <button
          type="button"
          class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
          @click="regenUpstreamTradeNo"
        >
          重新生成上游单号
        </button>
      </div>

      <div v-if="mockOk" class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-900">回调成功</div>
      <div v-if="mockError" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
        {{ mockError }}
      </div>
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
          <div v-if="signError" class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
            {{ signError }}
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import md5 from 'blueimp-md5'
import { computed, ref, watch } from 'vue'
import { OPEN_API } from '@/api/endpoints'
import { DEMO_PAY_PRODUCT_OPTIONS } from '@/config/payProducts'
import { loadMerchantAuth, saveMerchantAuth } from '@/lib/merchantApi'

type CreateOrderResp = {
  order_no: string
  status: number
  channel_id: number
  checkout_url: string
}

const auth = loadMerchantAuth()
const merchantId = ref(auth.merchantId)
const apiSecret = ref(auth.apiSecret)
const ipWhitelist = ref('127.0.0.1')
const notifyUrl = ref('')

const payProductOptions = DEMO_PAY_PRODUCT_OPTIONS
const payTypePreset = ref<(typeof DEMO_PAY_PRODUCT_OPTIONS)[number]['code'] | '__custom__'>('mock')
const payTypeCustom = ref('')
const resolvedPayType = computed(() =>
  payTypePreset.value === '__custom__' ? payTypeCustom.value.trim() : payTypePreset.value,
)

const merchantOrderNo = ref(`MO-${Date.now()}`)
const amount = ref(100)

const loading = ref(false)
const error = ref('')
const result = ref<CreateOrderResp | null>(null)

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

function regenOrderNo() {
  merchantOrderNo.value = `MO-${Date.now()}`
}

function regenUpstreamTradeNo() {
  mockUpstreamTradeNo.value = `UP-${Date.now()}`
}

function md5Sign(params: Record<string, string>, secret: string): string {
  const keys = Object.keys(params)
    .map((k) => k.toLowerCase())
    .filter((k) => k !== 'sign')
    .sort()
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
    /* ignore */
  }
  return bodyText.trim() || `HTTP ${status}`
}

async function createOrder() {
  loading.value = true
  error.value = ''
  result.value = null
  try {
    const params: Record<string, string> = {
      merchant_id: merchantId.value,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      pay_type: resolvedPayType.value,
      notify_url: notifyUrl.value,
    }
    const sign = md5Sign(params, apiSecret.value)
    const resp = await fetch(OPEN_API.payOrder, {
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
    const orderNo = queryOrderNo.value.trim()
    if (!orderNo) {
      queryError.value = '请填写 order_no'
      return
    }
    const params: Record<string, string> = {
      merchant_id: merchantId.value,
      order_no: orderNo,
      timestamp: String(Math.floor(Date.now() / 1000)),
    }
    const sign = md5Sign(params, apiSecret.value)
    const qs = new URLSearchParams({ ...params, sign })
    const resp = await fetch(`${OPEN_API.queryOrder}?${qs.toString()}`)
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
    const data = (await resp.json()) as { ok: boolean }
    if (!data.ok) {
      mockError.value = '回调返回 ok=false'
      return
    }
    mockOk.value = true
  } catch {
    mockError.value = '网络错误'
  } finally {
    mockLoading.value = false
  }
}

const localCheckoutUrl = computed(() => {
  if (!result.value?.order_no) return ''
  return `http://127.0.0.1:5174/?order_no=${encodeURIComponent(result.value.order_no)}`
})

const signJson = ref('')

function syncSignJsonFromForm() {
  signJson.value = JSON.stringify(
    {
      merchant_id: merchantId.value,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      pay_type: resolvedPayType.value,
    },
    null,
    2,
  )
}

syncSignJsonFromForm()

watch([merchantId, merchantOrderNo, amount, payTypePreset, payTypeCustom], () => {
  syncSignJsonFromForm()
})

const signSecret = ref(apiSecret.value)
watch(apiSecret, (v) => {
  signSecret.value = v
})

watch([merchantId, apiSecret], () => {
  saveMerchantAuth({ merchantId: merchantId.value, apiSecret: apiSecret.value })
})

const signError = ref('')
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
</script>

<style scoped>
.input-merchant {
  @apply w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
.textarea-merchant {
  @apply w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5 font-mono text-xs text-slate-900 shadow-inner focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
</style>
