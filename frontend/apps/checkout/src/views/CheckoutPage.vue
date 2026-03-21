<template>
  <div class="min-h-full bg-slate-100">
    <div class="mx-auto max-w-2xl px-4 py-10">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="text-sm text-slate-500">订单信息</div>
            <div class="mt-1 text-lg font-semibold text-slate-900">{{ merchantName }}</div>
          </div>
          <div class="text-right">
            <div class="text-xs text-slate-500">倒计时</div>
            <div class="mt-1 text-sm font-semibold text-slate-900">{{ countdownText }}</div>
          </div>
        </div>

        <div class="mt-5 grid gap-2 rounded-xl bg-slate-50 p-4">
          <div class="flex items-center justify-between">
            <div class="text-sm text-slate-500">金额</div>
            <div class="text-xl font-semibold text-slate-900">{{ amountText }}</div>
          </div>
          <div class="flex items-center justify-between">
            <div class="text-sm text-slate-500">订单号</div>
            <div class="break-all text-xs font-medium text-slate-700">{{ orderNo || '-' }}</div>
          </div>
          <div class="flex items-center justify-between">
            <div class="text-sm text-slate-500">状态</div>
            <div class="text-sm font-semibold" :class="statusClass">{{ statusText }}</div>
          </div>
        </div>

        <div class="mt-6">
          <div class="text-sm font-semibold text-slate-900">选择支付方式</div>
          <div class="mt-3 grid gap-2">
            <button
              v-for="m in methodsSorted"
              :key="m.key"
              class="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-3 text-left hover:bg-slate-50"
              :class="selectedMethod === m.key ? 'border-slate-900 bg-slate-50' : ''"
              @click="selectedMethod = m.key"
            >
              <div>
                <div class="text-sm font-semibold text-slate-900">{{ m.name }}</div>
                <div class="mt-1 text-xs text-slate-500">{{ m.desc }}</div>
              </div>
              <div v-if="selectedMethod === m.key" class="text-xs font-semibold text-slate-700">已选择</div>
            </button>
          </div>
        </div>

        <div class="mt-6 grid gap-3">
          <button
            class="w-full rounded-lg bg-slate-900 px-4 py-3 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="!orderNo || status !== 0"
            @click="payNow"
          >
            {{ payButtonText }}
          </button>

          <button
            class="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm font-semibold text-slate-700"
            @click="refresh"
          >
            刷新状态
          </button>
        </div>

        <div v-if="status === 0" class="mt-5 flex items-center gap-3 rounded-xl bg-slate-50 p-4">
          <div class="h-5 w-5 animate-spin rounded-full border-2 border-slate-300 border-t-slate-900"></div>
          <div class="text-sm text-slate-600">支付中，正在轮询支付状态…</div>
        </div>

        <div v-if="status === 1" class="mt-5 rounded-xl border border-emerald-200 bg-emerald-50 p-4">
          <div class="text-sm font-semibold text-emerald-800">支付成功</div>
          <div v-if="redirectText" class="mt-2 text-sm text-emerald-700">{{ redirectText }}</div>
        </div>

        <div v-if="status === 2 || status === 3" class="mt-5 rounded-xl border border-rose-200 bg-rose-50 p-4">
          <div class="text-sm font-semibold text-rose-800">{{ status === 2 ? '支付失败' : '订单已关闭' }}</div>
          <div class="mt-2 text-sm text-rose-700">请返回商户页面重新发起支付。</div>
        </div>

        <div v-if="error" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
          {{ error }}
        </div>
      </div>
    </div>

    <div v-if="showQrModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl">
        <div class="flex items-center justify-between">
          <div class="text-sm font-semibold text-slate-900">扫码支付</div>
          <button class="text-sm font-semibold text-slate-500 hover:text-slate-700" @click="showQrModal = false">
            关闭
          </button>
        </div>
        <div class="mt-4 grid place-items-center rounded-xl bg-slate-50 p-8">
          <div class="h-48 w-48 rounded-lg border border-dashed border-slate-300 bg-white"></div>
          <div class="mt-3 text-xs text-slate-500">这里可展示二维码（对接上游后替换）</div>
        </div>
        <div class="mt-4 grid grid-cols-2 gap-3">
          <button class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white" @click="refresh">已支付</button>
          <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700" @click="showQrModal = false">
            重新获取
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'

type OrderInfo = {
  order_no: string
  merchant_id: string
  amount: number
  currency: string
  status: number
  return_url: string
}

type MethodKey = 'wechat' | 'alipay' | 'unionpay' | 'bank' | 'crypto'

const route = useRoute()
const orderNo = computed(() => String(route.query.order_no || ''))

const merchantName = computed(() => '聚合支付')
const amount = ref<number>(0)
const currency = ref<string>('CNY')
const status = ref<number>(0)
const returnUrl = ref<string>('')

const error = ref('')
const showQrModal = ref(false)
const selectedMethod = ref<MethodKey>('wechat')

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

const countdownText = computed(() => {
  const s = secondsLeft.value
  const mm = String(Math.floor(s / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  return `${mm}:${ss}`
})

const amountText = computed(() => `${(amount.value / 100).toFixed(2)} ${currency.value || 'CNY'}`)

const statusText = computed(() => {
  if (!orderNo.value) return '缺少 order_no'
  if (status.value === 0) return '待支付'
  if (status.value === 1) return '支付成功'
  if (status.value === 2) return '支付失败'
  if (status.value === 3) return '已关闭'
  return `未知(${status.value})`
})

const statusClass = computed(() => {
  if (status.value === 1) return 'text-emerald-700'
  if (status.value === 2) return 'text-rose-700'
  if (status.value === 3) return 'text-slate-600'
  return 'text-amber-700'
})

const methods = [
  { key: 'wechat' as const, name: '微信支付', desc: '微信环境优先展示' },
  { key: 'alipay' as const, name: '支付宝', desc: '支持 H5 / 扫码' },
  { key: 'unionpay' as const, name: '云闪付', desc: '支持扫码/快捷' },
  { key: 'bank' as const, name: '网银', desc: '适用于 PC 场景' },
  { key: 'crypto' as const, name: '数字货币', desc: '可选通道（示例）' },
]

const isWeChat = computed(() => /micromessenger/i.test(navigator.userAgent))
const methodsSorted = computed(() => {
  if (!isWeChat.value) return methods
  return [...methods].sort((a, b) => (a.key === 'wechat' ? -1 : b.key === 'wechat' ? 1 : 0))
})

const isMobile = computed(() => /iphone|ipad|android/i.test(navigator.userAgent))

const payButtonText = computed(() => {
  if (isMobile.value) return '唤起 App 支付'
  return '扫码支付'
})

const redirectText = computed(() => {
  if (!returnUrl.value || redirectIn.value <= 0) return ''
  return `${redirectIn.value}s 后自动跳转`
})

async function load() {
  if (!orderNo.value) return
  const res = await fetch(`/v1/terminal/order?order_no=${encodeURIComponent(orderNo.value)}`)
  if (!res.ok) {
    throw new Error(String(res.status))
  }
  const data = (await res.json()) as { order: OrderInfo }
  amount.value = data.order.amount
  currency.value = data.order.currency
  status.value = data.order.status
  returnUrl.value = data.order.return_url || ''
}

async function refresh() {
  error.value = ''
  try {
    await load()
    if (status.value === 1 && returnUrl.value) {
      startRedirect()
    }
  } catch {
    error.value = '查询订单失败'
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

function payNow() {
  if (!orderNo.value) return
  if (isMobile.value) {
    error.value = '移动端唤起 App 需要对接上游 Schema（此处为占位）。'
    return
  }
  showQrModal.value = true
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
