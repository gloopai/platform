<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <div class="text-sm font-semibold text-slate-900">订单列表</div>
          <div class="mt-1 text-sm text-slate-600">支持按订单号、状态搜索。</div>
        </div>
        <div class="flex flex-wrap gap-2">
          <input v-model.trim="keyword" class="w-56 rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="订单号 / 商户订单号" />
          <select v-model="status" class="rounded-md border border-slate-200 px-3 py-2 text-sm">
            <option value="">全部状态</option>
            <option value="0">待支付</option>
            <option value="1">成功</option>
            <option value="2">失败</option>
            <option value="3">已关闭</option>
          </select>
          <button class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white" @click="reload">搜索</button>
        </div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <table class="w-full text-left text-sm">
        <thead class="bg-slate-50 text-xs font-semibold text-slate-600">
          <tr>
            <th class="px-4 py-3">订单号</th>
            <th class="px-4 py-3">金额</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">上游单号</th>
            <th class="px-4 py-3">创建时间</th>
            <th class="px-4 py-3">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-if="loading">
            <td class="px-4 py-3 text-slate-600" colspan="6">加载中...</td>
          </tr>
          <tr v-else-if="orders.length === 0">
            <td class="px-4 py-3 text-slate-600" colspan="6">暂无数据</td>
          </tr>
          <tr v-for="o in orders" :key="o.order_no">
            <td class="px-4 py-3">
              <div class="font-medium text-slate-900">{{ o.order_no }}</div>
              <div class="mt-1 text-xs text-slate-500">{{ o.merchant_order_no }}</div>
            </td>
            <td class="px-4 py-3 text-slate-700">{{ formatAmount(o.amount, o.currency) }}</td>
            <td class="px-4 py-3">
              <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="statusBadgeClass(o.status)">
                {{ statusLabel(o.status) }}
              </span>
            </td>
            <td class="px-4 py-3 text-slate-700">{{ o.upstream_trade_no || '-' }}</td>
            <td class="px-4 py-3 text-slate-700">{{ formatTime(o.created_at) }}</td>
            <td class="px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <button
                  class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50"
                  @click="openDetail(o.order_no)"
                >
                  详情
                </button>
                <button
                  class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-40"
                  :disabled="retrying"
                  @click="retryNotify(o.order_no)"
                >
                  重发通知
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
      {{ error }}
    </div>

    <div v-if="detailOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="w-full max-w-3xl rounded-2xl bg-white shadow-xl">
        <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <div>
            <div class="text-sm font-semibold text-slate-900">订单详情</div>
            <div class="mt-1 text-xs text-slate-500">{{ detail?.order.order_no }}</div>
          </div>
          <button class="text-sm font-semibold text-slate-600 hover:text-slate-900" @click="detailOpen = false">关闭</button>
        </div>

        <div class="grid gap-4 px-5 py-4">
          <div class="grid grid-cols-12 gap-3 rounded-xl bg-slate-50 p-4 text-sm">
            <div class="col-span-12 md:col-span-6">
              <div class="text-xs text-slate-500">商户单号</div>
              <div class="mt-1 font-medium text-slate-900">{{ detail?.order.merchant_order_no }}</div>
            </div>
            <div class="col-span-12 md:col-span-3">
              <div class="text-xs text-slate-500">金额</div>
              <div class="mt-1 font-medium text-slate-900">{{ formatAmount(detail?.order.amount || 0, detail?.order.currency || 'CNY') }}</div>
            </div>
            <div class="col-span-12 md:col-span-3">
              <div class="text-xs text-slate-500">状态</div>
              <div class="mt-1 font-medium text-slate-900">{{ statusLabel(detail?.order.status || 0) }}</div>
            </div>
            <div class="col-span-12">
              <div class="text-xs text-slate-500">Notify URL</div>
              <div class="mt-1 break-all font-medium text-slate-900">{{ detail?.order.notify_url || '-' }}</div>
            </div>
          </div>

          <div class="rounded-xl border border-slate-200">
            <div class="border-b border-slate-200 px-4 py-3 text-sm font-semibold text-slate-900">回调记录</div>
            <div class="max-h-80 overflow-auto">
              <table class="w-full text-left text-sm">
                <thead class="bg-slate-50 text-xs font-semibold text-slate-600">
                  <tr>
                    <th class="px-4 py-3">时间</th>
                    <th class="px-4 py-3">URL</th>
                    <th class="px-4 py-3">状态</th>
                    <th class="px-4 py-3">响应/错误</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-200">
                  <tr v-if="detailLoading">
                    <td class="px-4 py-3 text-slate-600" colspan="4">加载中...</td>
                  </tr>
                  <tr v-else-if="(detail?.logs?.length || 0) === 0">
                    <td class="px-4 py-3 text-slate-600" colspan="4">暂无记录</td>
                  </tr>
                  <tr v-for="l in detail?.logs || []" :key="l.id">
                    <td class="px-4 py-3 text-slate-700">{{ formatTime(l.created_at) }}</td>
                    <td class="px-4 py-3 text-slate-700">
                      <div class="max-w-xs break-all">{{ l.notify_url }}</div>
                      <div class="mt-1 text-xs text-slate-500">attempt={{ l.attempt }}</div>
                    </td>
                    <td class="px-4 py-3 text-slate-700">{{ l.http_status || '-' }}</td>
                    <td class="px-4 py-3 text-slate-700">
                      <div class="max-w-sm break-all font-mono text-xs">{{ l.response_body || l.error_msg || '-' }}</div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-if="detailError" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
            {{ detailError }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { merchantConsoleGet, merchantConsolePost } from '../../lib/merchantApi'

type MerchantOrderItem = {
  order_no: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  channel_id: number
  paid_amount: number
  upstream_trade_no: string
  created_at: number
}

type MerchantNotifyLogItem = {
  id: number
  notify_url: string
  attempt: number
  http_status: number
  response_body: string
  error_msg: string
  created_at: number
}

type MerchantOrderDetailResp = {
  order: {
    order_no: string
    merchant_id: string
    merchant_order_no: string
    amount: number
    currency: string
    status: number
    channel_id: number
    return_url: string
    notify_url: string
    upstream_trade_no: string
  }
  logs: MerchantNotifyLogItem[]
}

const keyword = ref('')
const status = ref('')
const orders = ref<MerchantOrderItem[]>([])
const loading = ref(false)
const error = ref('')
const retrying = ref(false)

const detailOpen = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref<MerchantOrderDetailResp | null>(null)

function statusLabel(v: number): string {
  if (v === 0) return '待支付'
  if (v === 1) return '成功'
  if (v === 2) return '失败'
  if (v === 3) return '已关闭'
  return `未知(${v})`
}

function statusBadgeClass(v: number): string {
  if (v === 1) return 'bg-emerald-100 text-emerald-700'
  if (v === 2) return 'bg-rose-100 text-rose-700'
  if (v === 3) return 'bg-slate-100 text-slate-600'
  return 'bg-amber-100 text-amber-700'
}

function formatAmount(amount: number, currency: string) {
  return `${(amount / 100).toFixed(2)} ${currency || 'CNY'}`
}

function formatTime(ts: number) {
  const d = new Date(ts * 1000)
  return d.toLocaleString()
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await merchantConsoleGet<{ orders: MerchantOrderItem[] }>('/v1/merchant/orders', {
      order_no: keyword.value,
      status: status.value,
      limit: 50,
    })
    orders.value = res.orders || []
  } catch {
    error.value = '加载失败：请确认已登录且网关已启动。'
  } finally {
    loading.value = false
  }
}

async function openDetail(orderNo: string) {
  detailOpen.value = true
  detailLoading.value = true
  detailError.value = ''
  detail.value = null
  try {
    detail.value = await merchantConsoleGet<MerchantOrderDetailResp>('/v1/merchant/order/detail', { order_no: orderNo })
  } catch {
    detailError.value = '加载详情失败'
  } finally {
    detailLoading.value = false
  }
}

async function retryNotify(orderNo: string) {
  retrying.value = true
  error.value = ''
  try {
    await merchantConsolePost<{ ok: boolean }>('/v1/merchant/order/retry_notify', { order_no: orderNo })
  } catch {
    error.value = '重发通知失败'
  } finally {
    retrying.value = false
  }
}

onMounted(() => {
  void reload()
})
</script>
