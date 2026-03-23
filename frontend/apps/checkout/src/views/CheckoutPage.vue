<template>
  <div
    class="min-h-full bg-gradient-to-b from-slate-100/95 via-white to-slate-50/90 text-slate-900 antialiased"
  >
    <div class="mx-auto max-w-lg px-4 pb-10 pt-6 sm:px-5 sm:pt-10">
      <header class="mb-5 flex items-center justify-between gap-3 sm:mb-6">
        <div class="flex items-center gap-2.5">
          <div
            class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-slate-600 to-slate-800 text-sm font-bold text-white shadow-md shadow-slate-900/15"
            aria-hidden="true"
          >
            P
          </div>
          <div>
            <h1 class="text-base font-semibold tracking-tight text-slate-900">收银台</h1>
            <p class="text-[11px] text-slate-500">聚合支付 · 安全收银</p>
          </div>
        </div>
        <div
          class="hidden items-center gap-1.5 rounded-full border border-slate-200/90 bg-slate-50/90 px-2.5 py-1 text-[10px] font-medium text-slate-700 sm:flex"
          role="status"
        >
          <svg class="h-3.5 w-3.5 text-slate-600" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d="M12 2 4 6v6c0 5 3.5 9.5 8 10.5 4.5-1 8-5.5 8-10.5V6l-8-4Z"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linejoin="round"
            />
            <path d="m9.5 12.5 1.8 1.8 3.7-4.3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
          加密传输
        </div>
      </header>

      <!-- 缺少订单参数 -->
      <div
        v-if="!orderNo"
        class="rounded-2xl border border-amber-200/90 bg-amber-50/90 p-6 text-center shadow-sm"
      >
        <p class="text-sm font-semibold text-amber-900">未找到订单</p>
        <p class="mt-2 text-xs leading-relaxed text-amber-800/90">
          请通过商户跳转链接进入，或在地址后附加参数
          <code class="rounded bg-amber-100/80 px-1 py-0.5 font-mono text-[11px]">?order_no=</code>
        </p>
      </div>

      <div v-else class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-lg shadow-slate-900/[0.06]">
        <div class="border-b border-slate-100 bg-gradient-to-r from-slate-50/80 to-white px-5 py-4 sm:px-6 sm:py-5">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p class="text-[11px] font-medium uppercase tracking-wider text-slate-400">订单信息</p>
              <p class="mt-1 truncate text-lg font-semibold text-slate-900">{{ orderTitle }}</p>
              <p v-if="merchantIdDisplay" class="mt-0.5 truncate font-mono text-[11px] text-slate-500">
                {{ merchantIdDisplay }}
              </p>
            </div>
            <div class="shrink-0 text-right">
              <p class="text-[11px] font-medium text-slate-400">剩余时间</p>
              <p
                class="mt-1 tabular-nums text-sm font-semibold"
                :class="countdownUrgent ? 'text-amber-700' : isExpired ? 'text-rose-600' : 'text-slate-900'"
              >
                {{ isExpired ? '已超时' : countdownText }}
              </p>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-slate-100 bg-white p-4">
            <div class="flex items-end justify-between gap-3">
              <span class="text-sm text-slate-500">应付金额</span>
              <span class="text-2xl font-bold tabular-nums tracking-tight text-slate-900 sm:text-3xl">
                {{ amountText }}
              </span>
            </div>
            <div class="mt-4 flex flex-col gap-2 border-t border-slate-100 pt-4 text-sm">
              <div class="flex items-start justify-between gap-2">
                <span class="shrink-0 text-slate-500">平台订单号</span>
                <div class="flex min-w-0 items-center justify-end gap-1.5">
                  <span class="break-all text-right font-mono text-xs font-medium text-slate-700">{{ orderNo }}</span>
                  <button
                    type="button"
                    class="shrink-0 rounded-lg border border-slate-200 bg-white px-2 py-0.5 text-[10px] font-semibold text-slate-600 transition hover:border-slate-300 hover:bg-slate-50"
                    @click="copyOrderNo"
                  >
                    {{ copiedOrderNo ? '已复制' : '复制' }}
                  </button>
                </div>
              </div>
              <div v-if="merchantOrderNoDisplay" class="flex items-start justify-between gap-2">
                <span class="shrink-0 text-slate-500">商户订单号</span>
                <span class="break-all text-right font-mono text-xs text-slate-700">{{ merchantOrderNoDisplay }}</span>
              </div>
              <div class="flex items-center justify-between gap-2">
                <span class="text-slate-500">状态</span>
                <span class="font-semibold" :class="statusClass">{{ statusText }}</span>
              </div>
              <div v-if="payinProductCodeOnOrder" class="flex items-center justify-between gap-2">
                <span class="text-slate-500">下单支付产品</span>
                <span class="font-mono text-xs font-medium text-slate-800">{{ payinProductCodeOnOrder }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="px-5 py-5 sm:px-6">
          <p class="text-sm font-semibold text-slate-900">选择支付方式</p>
          <p v-if="channelLocked" class="mt-1 text-xs text-slate-500">商户已指定支付通道，不可更换其他方式。</p>
          <div class="mt-3 grid gap-2" role="radiogroup" aria-label="支付方式">
            <button
              v-for="m in methodsSorted"
              :key="m.key"
              type="button"
              role="radio"
              :aria-checked="selectedMethod === m.key"
              class="flex w-full items-center gap-3 rounded-xl border px-4 py-3.5 text-left transition focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-400/45"
              :class="
                selectedMethod === m.key
                  ? 'border-slate-400 bg-slate-50 ring-1 ring-slate-300/60'
                  : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50/80'
              "
              @click="selectedMethod = m.key"
            >
              <span
                class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2"
                :class="selectedMethod === m.key ? 'border-slate-700 bg-slate-700' : 'border-slate-300'"
                aria-hidden="true"
              >
                <span v-if="selectedMethod === m.key" class="h-2 w-2 rounded-full bg-white" />
              </span>
              <div class="min-w-0 flex-1">
                <div class="text-sm font-semibold text-slate-900">{{ m.name }}</div>
                <div class="mt-0.5 text-xs text-slate-500">{{ m.desc }}</div>
              </div>
            </button>
          </div>

          <div class="mt-6 grid gap-3">
            <button
              type="button"
              class="w-full rounded-xl bg-slate-800 px-4 py-3.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40 disabled:shadow-none"
              :disabled="payDisabled"
              @click="payNow"
            >
              {{ paying ? '处理中…' : payButtonText }}
            </button>

            <button
              type="button"
              class="w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="refresh"
            >
              刷新状态
            </button>
          </div>

          <div
            v-if="status === 0 && !isExpired"
            class="mt-5 flex items-start gap-3 rounded-xl border border-slate-100 bg-slate-50/90 p-4"
          >
            <span
              class="mt-0.5 inline-flex h-2 w-2 shrink-0 animate-pulse rounded-full bg-slate-500"
              aria-hidden="true"
            />
            <p class="text-sm leading-relaxed text-slate-600">
              待支付：页面将自动刷新订单状态，完成支付后请勿关闭窗口。
            </p>
          </div>

          <div
            v-if="status === 0 && isExpired"
            class="mt-5 rounded-xl border border-rose-200 bg-rose-50 p-4"
            role="alert"
          >
            <p class="text-sm font-semibold text-rose-900">支付已超时</p>
            <p class="mt-1 text-sm text-rose-800/90">请返回商户重新下单或联系商户处理。</p>
          </div>

          <div v-if="status === 1" class="mt-5 rounded-xl border border-slate-200 bg-slate-50/95 p-4">
            <div class="text-sm font-semibold text-slate-900">支付成功</div>
            <div v-if="redirectText" class="mt-2 text-sm text-slate-600">{{ redirectText }}</div>
          </div>

          <div v-if="status === 2 || status === 3" class="mt-5 rounded-xl border border-rose-200 bg-rose-50 p-4">
            <div class="text-sm font-semibold text-rose-900">{{ status === 2 ? '支付失败' : '订单已关闭' }}</div>
            <div class="mt-2 text-sm text-rose-800/90">请返回商户页面重新发起支付。</div>
          </div>

          <div v-if="error" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
            {{ error }}
          </div>
        </div>
      </div>

      <p class="mt-6 text-center text-[11px] text-slate-400">支付遇到问题？请返回商户应用或联系商户客服。</p>
    </div>

    <div
      v-if="showQrModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-labelledby="qr-modal-title"
    >
      <div class="w-full max-w-sm rounded-2xl border border-slate-200/90 bg-white p-6 shadow-2xl">
        <div class="flex items-center justify-between gap-2">
          <div id="qr-modal-title" class="text-sm font-semibold text-slate-900">扫码支付</div>
          <button
            type="button"
            class="rounded-lg px-2 py-1 text-sm font-semibold text-slate-500 transition hover:bg-slate-100 hover:text-slate-800"
            @click="showQrModal = false"
          >
            关闭
          </button>
        </div>
        <div class="mt-4 grid place-items-center rounded-xl bg-slate-50 p-8">
          <img
            v-if="prepayPayload?.qr_payload"
            :src="qrImgSrc(prepayPayload.qr_payload)"
            alt="支付二维码"
            class="h-48 w-48 rounded-lg border border-slate-200 bg-white object-contain p-2 shadow-inner"
          />
          <div v-else class="h-48 w-48 rounded-lg border border-dashed border-slate-300 bg-white shadow-inner" />
          <div class="mt-3 max-w-xs text-center text-xs leading-relaxed text-slate-500">
            内容由 <code class="rounded bg-slate-100 px-1 font-mono text-[10px]">POST /v1/terminal/pay</code> 返回；联调支付完成后可用仓库
            <code class="font-mono text-[10px]">simulate_upstream</code> 模拟回调。
          </div>
        </div>
        <div class="mt-4 grid grid-cols-2 gap-3">
          <button
            type="button"
            class="rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800"
            @click="refresh"
          >
            我已完成支付
          </button>
          <button
            type="button"
            class="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50"
            @click="showQrModal = false"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

type OrderInfo = {
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  return_url: string
  payin_product_code?: string
  /** 1 = 商户下单已指定通道，不可改支付方式 */
  channel_locked?: number
}

type PayProductItem = { code: string; name: string }

type TerminalOrderPayload = {
  order: OrderInfo
  payin_products?: PayProductItem[]
}

type MethodRow = { key: string; name: string; desc: string }

/** 与网关 payin_products / 产品编码一致；展示名以服务端为准 */
const DESC_BY_CODE: Record<string, string> = {
  mock: '联调占位，平台路由至上游',
  wechat: '微信内优先使用',
  alipay: '支持 H5 / 扫码',
  unionpay: '支持扫码或快捷支付',
  bank: '适用于 PC 场景',
  crypto: '可选通道（示例）',
}

const FALLBACK_METHODS: MethodRow[] = [
  { key: 'wechat', name: '微信支付', desc: DESC_BY_CODE.wechat },
  { key: 'alipay', name: '支付宝', desc: DESC_BY_CODE.alipay },
  { key: 'unionpay', name: '云闪付', desc: DESC_BY_CODE.unionpay },
  { key: 'bank', name: '网银', desc: DESC_BY_CODE.bank },
  { key: 'crypto', name: '数字货币', desc: DESC_BY_CODE.crypto },
]

const route = useRoute()
const orderNo = computed(() => String(route.query.order_no || '').trim())

const merchantIdDisplay = ref('')
const merchantOrderNoDisplay = ref('')
const amount = ref<number>(0)
const currency = ref<string>('CNY')
const status = ref<number>(0)
const returnUrl = ref<string>('')
const payinProductCodeOnOrder = ref('')
const channelLocked = ref(false)

const error = ref('')
const showQrModal = ref(false)
const prepayPayload = ref<{ pay_url: string; qr_payload: string; pay_mode: string } | null>(null)
const paying = ref(false)
const selectedMethod = ref('')
const serverPayProducts = ref<PayProductItem[] | null>(null)
const copiedOrderNo = ref(false)

const startedAt = Date.now()
const ttlSeconds = 15 * 60
const now = ref(Date.now())
let timer: number | null = null
let poller: number | null = null
let redirectTimer: number | null = null
const redirectIn = ref<number>(0)

const secondsLeft = computed(() => {
  const passed = Math.floor((now.value - startedAt) / 1000)
  return Math.max(0, ttlSeconds - passed)
})

const isExpired = computed(() => secondsLeft.value <= 0)

const countdownUrgent = computed(() => !isExpired.value && secondsLeft.value <= 5 * 60)

const countdownText = computed(() => {
  const s = secondsLeft.value
  const mm = String(Math.floor(s / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  return `${mm}:${ss}`
})

const amountText = computed(() => `${(amount.value / 100).toFixed(2)} ${currency.value || 'CNY'}`)

const orderTitle = computed(() => {
  if (merchantOrderNoDisplay.value) return merchantOrderNoDisplay.value
  if (orderNo.value) return '订单支付'
  return '—'
})

const statusText = computed(() => {
  if (!orderNo.value) return '—'
  if (status.value === 0) return '待支付'
  if (status.value === 1) return '支付成功'
  if (status.value === 2) return '支付失败'
  if (status.value === 3) return '已关闭'
  return `未知(${status.value})`
})

const statusClass = computed(() => {
  if (status.value === 1) return 'text-slate-800'
  if (status.value === 2) return 'text-rose-700'
  if (status.value === 3) return 'text-slate-600'
  return 'text-amber-700'
})

const isWeChat = computed(() => /micromessenger/i.test(navigator.userAgent))

const methodsSorted = computed((): MethodRow[] => {
  const fromServer = serverPayProducts.value
  const base: MethodRow[] =
    fromServer && fromServer.length > 0
      ? fromServer.map((p) => ({
          key: p.code,
          name: p.name,
          desc: DESC_BY_CODE[p.code] || '由平台路由至对应上游通道',
        }))
      : FALLBACK_METHODS
  if (!isWeChat.value) return base
  return [...base].sort((a, b) => (a.key === 'wechat' ? -1 : b.key === 'wechat' ? 1 : 0))
})

watch(
  methodsSorted,
  (list) => {
    if (list.length === 0) return
    if (!list.some((m) => m.key === selectedMethod.value)) {
      selectedMethod.value = list[0].key
    }
  },
  { immediate: true },
)

const isMobile = computed(() => /iphone|ipad|android/i.test(navigator.userAgent))

const payButtonText = computed(() => {
  if (isExpired.value) return '支付已超时'
  if (isMobile.value) return '发起支付'
  return '确认支付'
})

const payDisabled = computed(
  () => !orderNo.value || status.value !== 0 || isExpired.value || paying.value || !selectedMethod.value,
)

const redirectText = computed(() => {
  if (!returnUrl.value || redirectIn.value <= 0) return ''
  return `${redirectIn.value} 秒后自动跳转`
})

async function openApiErrorText(res: Response): Promise<string> {
  const t = await res.text()
  try {
    const j = JSON.parse(t) as { code?: string; message?: string }
    if (j?.code) return `${j.code}: ${j.message ?? ''}`
  } catch {
    /* ignore */
  }
  return t.trim() || `HTTP ${res.status}`
}

async function copyOrderNo() {
  if (!orderNo.value) return
  try {
    await navigator.clipboard.writeText(orderNo.value)
    copiedOrderNo.value = true
    window.setTimeout(() => {
      copiedOrderNo.value = false
    }, 2000)
  } catch {
    error.value = '复制失败，请长按订单号手动复制'
    window.setTimeout(() => {
      if (error.value === '复制失败，请长按订单号手动复制') error.value = ''
    }, 3000)
  }
}

async function load() {
  if (!orderNo.value) return
  const res = await fetch(`/v1/terminal/order?order_no=${encodeURIComponent(orderNo.value)}`)
  if (!res.ok) {
    throw new Error(await openApiErrorText(res))
  }
  const data = (await res.json()) as TerminalOrderPayload
  amount.value = data.order.amount
  currency.value = data.order.currency
  status.value = data.order.status
  returnUrl.value = data.order.return_url || ''
  merchantIdDisplay.value = data.order.merchant_id || ''
  merchantOrderNoDisplay.value = data.order.merchant_order_no || ''
  payinProductCodeOnOrder.value = data.order.payin_product_code || ''
  channelLocked.value = Number(data.order.channel_locked) === 1
  serverPayProducts.value = data.payin_products && data.payin_products.length > 0 ? data.payin_products : null
}

async function refresh() {
  error.value = ''
  try {
    await load()
    if (status.value === 1 && returnUrl.value) {
      startRedirect()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '查询订单失败'
  }
}

function startRedirect() {
  if (redirectTimer != null) return
  redirectIn.value = 3
  redirectTimer = window.setInterval(() => {
    redirectIn.value -= 1
    if (redirectIn.value <= 0) {
      if (redirectTimer != null) window.clearInterval(redirectTimer)
      redirectTimer = null
      window.location.href = returnUrl.value
    }
  }, 1000)
}

function qrImgSrc(payload: string) {
  return `https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encodeURIComponent(payload)}`
}

async function payNow() {
  if (!orderNo.value || isExpired.value || !selectedMethod.value) return
  paying.value = true
  error.value = ''
  try {
    const res = await fetch('/v1/terminal/pay', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        order_no: orderNo.value,
        payin_product_code: selectedMethod.value,
      }),
    })
    if (!res.ok) {
      error.value = await openApiErrorText(res)
      return
    }
    const data = (await res.json()) as {
      pay_url: string
      qr_payload: string
      pay_mode: string
    }
    prepayPayload.value = data
    const httpUrl = data.pay_url.startsWith('http://') || data.pay_url.startsWith('https://')
    if (isMobile.value && httpUrl) {
      window.location.href = data.pay_url
      return
    }
    showQrModal.value = true
  } catch {
    error.value = '网络错误'
  } finally {
    paying.value = false
  }
}

onMounted(async () => {
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 250)
  await refresh()
  poller = window.setInterval(() => {
    if (status.value === 0 && secondsLeft.value > 0) void refresh()
  }, 2000)
})

onUnmounted(() => {
  if (timer != null) window.clearInterval(timer)
  if (poller != null) window.clearInterval(poller)
  if (redirectTimer != null) window.clearInterval(redirectTimer)
})
</script>
