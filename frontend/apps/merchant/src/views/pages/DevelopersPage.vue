<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">参数配置</div>
      <div class="mt-4 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">merchant_id</span>
          <input v-model.trim="merchantId" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">app_secret</span>
          <input v-model.trim="apiSecret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">IP 白名单</span>
          <input v-model.trim="ipWhitelist" class="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="例如：127.0.0.1,10.0.0.0/24" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">Notify URL</span>
          <input v-model.trim="notifyUrl" class="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="https://merchant.example.com/notify" />
        </label>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">下单联调</div>
      <div class="mt-2 text-sm text-slate-600">调用 /v1/pay/order 创建订单并跳转收银台。</div>

      <div class="mt-4 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">merchant_order_no</span>
          <input v-model.trim="merchantOrderNo" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">amount(分)</span>
          <input v-model.number="amount" type="number" min="1" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="loading || !merchantId || !apiSecret || !merchantOrderNo || amount <= 0"
          @click="createOrder"
        >
          {{ loading ? '创建中...' : '创建订单' }}
        </button>
        <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700" @click="regenOrderNo">
          重新生成订单号
        </button>
      </div>

      <div v-if="result" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-800">
        <div>order_no: {{ result.order_no }}</div>
        <div class="mt-2 flex flex-wrap gap-3">
          <a class="font-semibold underline" :href="result.checkout_url" target="_blank" rel="noreferrer">打开返回的 checkout_url</a>
          <a class="font-semibold underline" :href="localCheckoutUrl" target="_blank" rel="noreferrer">打开独立收银台</a>
        </div>
      </div>

      <div v-if="error" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
        {{ error }}
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">模拟上游回调（测试支付成功链路）</div>
      <div class="mt-2 text-sm text-slate-600">调用 /v1/callback/notify，触发订单成功、入账与异步通知。</div>

      <div class="mt-4 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">channel_id</span>
          <input v-model.number="mockChannelId" type="number" min="1" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">paid_amount(分)</span>
          <input v-model.number="mockPaidAmount" type="number" min="1" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1 md:col-span-4">
          <span class="text-xs font-medium text-slate-600">upstream_trade_no</span>
          <input v-model.trim="mockUpstreamTradeNo" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="col-span-12 grid gap-1">
          <span class="text-xs font-medium text-slate-600">channel_sign_secret</span>
          <input v-model.trim="mockChannelSecret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="默认 seed_demo.sql 为 channel_secret" />
        </label>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="mockLoading || !result?.order_no || mockChannelId <= 0 || mockPaidAmount <= 0 || !mockUpstreamTradeNo || !mockChannelSecret"
          @click="mockNotify"
        >
          {{ mockLoading ? '回调中...' : '触发支付成功' }}
        </button>
        <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700" @click="regenUpstreamTradeNo">
          重新生成上游单号
        </button>
      </div>

      <div v-if="mockOk" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-800">回调成功</div>
      <div v-if="mockError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
        {{ mockError }}
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">签名工具</div>
      <div class="mt-2 text-sm text-slate-600">按参数名排序后 MD5 签名（追加 key=secret）。</div>

      <div class="mt-4 grid grid-cols-12 gap-4">
        <label class="col-span-12 grid gap-1 md:col-span-6">
          <span class="text-xs font-medium text-slate-600">参数 JSON</span>
          <textarea v-model="signJson" rows="10" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-xs" />
        </label>
        <div class="col-span-12 grid gap-3 md:col-span-6">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">secret</span>
            <input v-model.trim="signSecret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <div class="rounded-xl bg-slate-50 p-4">
            <div class="text-xs font-semibold text-slate-700">签名结果</div>
            <div class="mt-2 break-all font-mono text-xs text-slate-800">{{ signOutput }}</div>
          </div>
          <div v-if="signError" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
            {{ signError }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import md5 from 'blueimp-md5'
import { computed, ref, watch } from 'vue'
import { loadMerchantAuth, saveMerchantAuth } from '../../lib/merchantApi'

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

const merchantOrderNo = ref(`MO-${Date.now()}`)
const amount = ref(100)

const loading = ref(false)
const error = ref('')
const result = ref<CreateOrderResp | null>(null)

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
      pay_type: 'mock',
      notify_url: notifyUrl.value,
    }
    const sign = md5Sign(params, apiSecret.value)
    const resp = await fetch('/v1/pay/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...params, sign }),
    })
    if (!resp.ok) {
      error.value = `创建失败(${resp.status})`
      return
    }
    result.value = (await resp.json()) as CreateOrderResp
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
    const resp = await fetch('/v1/callback/notify', {
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

const signJson = ref(
  JSON.stringify(
    {
      merchant_id: merchantId.value,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      pay_type: 'mock',
    },
    null,
    2,
  ),
)

watch([merchantId, merchantOrderNo, amount], () => {
  signJson.value = JSON.stringify(
    {
      merchant_id: merchantId.value,
      merchant_order_no: merchantOrderNo.value,
      amount: String(amount.value),
      currency: 'CNY',
      pay_type: 'mock',
    },
    null,
    2,
  )
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
