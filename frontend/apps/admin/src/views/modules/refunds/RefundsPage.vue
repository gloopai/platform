<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">退款与差错</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：先提供<strong>退款/差错候选订单</strong>只读列表（支付失败、已关闭），用于运营初筛；正式退款流、审核流与调账流程后续接入。需要核对全量订单请使用
        <router-link to="/payin-orders" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          全站订单
        </router-link>
        。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="flex flex-wrap items-end gap-3">
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">商户（可选）</span>
        <input
          v-model.trim="merchantId"
          type="text"
          placeholder="merchant_id"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @keyup.enter="load"
        />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">关键词（可选）</span>
        <input
          v-model.trim="keyword"
          type="text"
          placeholder="订单号 / 商户单号"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @keyup.enter="load"
        />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">状态</span>
        <select v-model="status" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm">
          <option value="all">全部候选</option>
          <option value="failed">支付失败</option>
          <option value="closed">已关闭</option>
        </select>
      </label>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
        @click="load"
      >
        加载
      </button>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[980px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">时间</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3">平台订单号</th>
              <th class="whitespace-nowrap px-4 py-3">商户</th>
              <th class="whitespace-nowrap px-4 py-3">商户单号</th>
              <th class="whitespace-nowrap px-4 py-3">金额</th>
              <th class="whitespace-nowrap px-4 py-3">产品</th>
              <th class="whitespace-nowrap px-4 py-3">通道</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中…</td>
            </tr>
            <tr v-else-if="!items.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="8">暂无候选订单</td>
            </tr>
            <tr v-for="x in pagedItems" v-else :key="x.order_no" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ formatTs(x.created_at) }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="x.status === 2 ? 'bg-rose-100 text-rose-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ x.status_label }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ x.order_no }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ x.merchant_id }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ x.merchant_order_no }}</td>
              <td class="px-4 py-3 font-semibold text-slate-900">{{ formatAmount(x.amount) }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ x.payin_product_code || '—' }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">#{{ x.channel_id }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <AdminPaginationBar
        v-if="!loading && items.length > 0"
        v-model:page="page"
        v-model:pageSize="pageSize"
        :total="total"
        :page-count="pageCount"
      />
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划（未实现）</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>退款申请与审核、原路退回与通道侧状态同步</li>
        <li>差错登记、长款挂账与人工调账</li>
        <li>与对账中心差异类型联动</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { useClientPagination } from '../../../composables/useClientPagination'

import { adminGet } from '../../../lib/adminApi'
import { formatAdminMoney } from '../../../lib/displaySettings'

type RefundItem = {
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: number
  status_label: string
  channel_id: number
  payin_product_code: string
  upstream_trade_no: string
  created_at: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const merchantId = ref('')
const keyword = ref('')
const status = ref<'all' | 'failed' | 'closed'>('all')
const loading = ref(true)
const error = ref('')
const items = ref<RefundItem[]>([])
const { page, pageSize, total, pageCount, slice: pagedItems } = useClientPagination(items, 20)

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime())
    ? '—'
    : d.toLocaleString('zh-CN', { hour12: false, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatAmount(cents: number): string {
  return formatAdminMoney(cents)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    page.value = 1
    const q = new URLSearchParams({ status: status.value, limit: '200' })
    if (merchantId.value) q.set('merchant_id', merchantId.value)
    if (keyword.value) q.set('keyword', keyword.value)
    const r = await adminGet<{ items: RefundItem[] }>(`/v1/admin/refunds?${q.toString()}`)
    items.value = r.items ?? []
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    items.value = []
  } finally {
    loading.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void load()
  if (registerRefresh) unregister = registerRefresh(() => void load())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
