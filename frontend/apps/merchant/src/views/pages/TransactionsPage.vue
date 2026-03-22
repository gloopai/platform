<template>
  <div class="space-y-6">
    <PageHeader title="交易管理" description="查询订单、查看回调记录并重发通知" />

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <div class="text-sm font-semibold text-slate-900">筛选</div>
          <p class="mt-1 text-xs text-slate-500">支持订单号、商户订单号与状态</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <input
            v-model.trim="keyword"
            class="min-w-[12rem] flex-1 rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20 sm:max-w-xs"
            placeholder="订单号 / 商户订单号"
          />
          <select
            v-model="status"
            class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm shadow-inner focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20"
          >
            <option value="">全部状态</option>
            <option value="0">待支付</option>
            <option value="1">成功</option>
            <option value="2">失败</option>
            <option value="3">已关闭</option>
          </select>
          <button
            type="button"
            class="rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700"
            @click="reload"
          >
            搜索
          </button>
        </div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">订单号</th>
              <th class="whitespace-nowrap px-4 py-3">金额</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3">上游单号</th>
              <th class="whitespace-nowrap px-4 py-3">创建时间</th>
              <th class="whitespace-nowrap px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="6">加载中…</td>
            </tr>
            <tr v-else-if="orders.length === 0">
              <td class="px-4 py-12 text-center text-slate-500" colspan="6">
                <div class="mx-auto flex max-w-sm flex-col items-center gap-2">
                  <span class="rounded-full bg-slate-100 p-3 text-slate-400">
                    <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.25">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                  </span>
                  <span class="text-sm font-medium text-slate-700">暂无订单</span>
                  <span class="text-xs text-slate-500">调整筛选条件或稍后再试</span>
                </div>
              </td>
            </tr>
            <tr v-for="o in orders" :key="o.order_no" class="transition hover:bg-slate-50/80">
              <td class="px-4 py-3 align-top">
                <div class="font-medium text-slate-900">{{ o.order_no }}</div>
                <div class="mt-0.5 font-mono text-xs text-slate-500">{{ o.merchant_order_no }}</div>
              </td>
              <td class="px-4 py-3 align-top tabular-nums text-slate-800">{{ formatAmount(o.amount, o.currency) }}</td>
              <td class="px-4 py-3 align-top">
                <span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold" :class="statusBadgeClass(o.status)">
                  {{ statusLabel(o.status) }}
                </span>
              </td>
              <td class="px-4 py-3 align-top text-slate-700">{{ o.upstream_trade_no || '—' }}</td>
              <td class="px-4 py-3 align-top text-slate-600">{{ formatTime(o.created_at) }}</td>
              <td class="px-4 py-3 align-top">
                <div class="flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900"
                    @click="openDetail(o.order_no)"
                  >
                    详情
                  </button>
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-50 disabled:opacity-40"
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
    </div>

    <ErrorCallout v-if="error" :message="error" />

    <div v-if="detailOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-sm">
      <div class="max-h-[90vh] w-full max-w-3xl overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <div>
            <div class="text-sm font-semibold text-slate-900">订单详情</div>
            <div class="mt-0.5 font-mono text-xs text-slate-500">{{ detail?.order.order_no }}</div>
          </div>
          <button
            type="button"
            class="rounded-lg px-3 py-1.5 text-sm font-semibold text-slate-600 transition hover:bg-slate-100 hover:text-slate-900"
            @click="detailOpen = false"
          >
            关闭
          </button>
        </div>

        <div class="max-h-[calc(90vh-5rem)] overflow-y-auto px-5 py-4">
          <div class="grid gap-4">
            <div class="grid grid-cols-12 gap-3 rounded-2xl border border-slate-100 bg-slate-50/80 p-4 text-sm">
              <div class="col-span-12 md:col-span-6">
                <div class="text-xs font-medium text-slate-500">商户单号</div>
                <div class="mt-1 font-medium text-slate-900">{{ detail?.order.merchant_order_no }}</div>
              </div>
              <div class="col-span-12 md:col-span-3">
                <div class="text-xs font-medium text-slate-500">金额</div>
                <div class="mt-1 font-medium text-slate-900">{{ formatAmount(detail?.order.amount || 0, detail?.order.currency || 'CNY') }}</div>
              </div>
              <div class="col-span-12 md:col-span-3">
                <div class="text-xs font-medium text-slate-500">状态</div>
                <div class="mt-1 font-medium text-slate-900">{{ statusLabel(detail?.order.status || 0) }}</div>
              </div>
              <div class="col-span-12">
                <div class="text-xs font-medium text-slate-500">Notify URL</div>
                <div class="mt-1 break-all font-mono text-xs text-slate-800">{{ detail?.order.notify_url || '—' }}</div>
              </div>
            </div>

            <div class="overflow-hidden rounded-2xl border border-slate-200">
              <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">回调记录</div>
              <div class="max-h-80 overflow-auto">
                <table class="w-full text-left text-sm">
                  <thead class="sticky top-0 bg-white text-xs font-semibold text-slate-500">
                    <tr class="border-b border-slate-100">
                      <th class="px-4 py-2">时间</th>
                      <th class="px-4 py-2">URL</th>
                      <th class="px-4 py-2">状态</th>
                      <th class="px-4 py-2">响应/错误</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-if="detailLoading">
                      <td class="px-4 py-4 text-slate-500" colspan="4">加载中…</td>
                    </tr>
                    <tr v-else-if="(detail?.logs?.length || 0) === 0">
                      <td class="px-4 py-4 text-slate-500" colspan="4">暂无记录</td>
                    </tr>
                    <tr v-for="l in detail?.logs || []" :key="l.id">
                      <td class="px-4 py-2 align-top text-slate-700">{{ formatTime(l.created_at) }}</td>
                      <td class="px-4 py-2 align-top text-slate-700">
                        <div class="max-w-xs break-all text-xs">{{ l.notify_url }}</div>
                        <div class="mt-1 text-xs text-slate-400">attempt={{ l.attempt }}</div>
                      </td>
                      <td class="px-4 py-2 align-top text-slate-700">{{ l.http_status || '—' }}</td>
                      <td class="px-4 py-2 align-top text-slate-700">
                        <div class="max-w-sm break-all font-mono text-xs">{{ l.response_body || l.error_msg || '—' }}</div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div v-if="detailError" class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
              {{ detailError }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { fetchMerchantOrderDetail, fetchMerchantOrders, postRetryMerchantNotify } from '@/api/orders'
import type { MerchantOrderDetailResp, MerchantOrderItem } from '@/types/merchant.api'
import { formatCentsWithCurrency, formatUnixSeconds } from '@/utils/format'
import { orderStatusBadgeClass as statusBadgeClass, orderStatusLabel as statusLabel } from '@/utils/orderStatus'

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

function formatAmount(amount: number, currency: string) {
  return formatCentsWithCurrency(amount, currency)
}

function formatTime(ts: number) {
  return formatUnixSeconds(ts)
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchMerchantOrders({
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
    detail.value = await fetchMerchantOrderDetail(orderNo)
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
    await postRetryMerchantNotify(orderNo)
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
