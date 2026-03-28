<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">{{ title }}</h1>
      <p class="mt-1 text-sm text-slate-600">
        {{ description }}
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
      <div class="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-end">
        <label class="grid min-w-[10rem] gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">关键词</span>
          <input
            v-model.trim="keyword"
            type="search"
            placeholder="订单号 / 商户单号 / 商户 ID"
            class="rounded-lg border border-slate-200 px-3 py-2 text-sm"
            @keydown.enter.prevent="reload"
          />
        </label>
        <label class="grid min-w-[8rem] gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">商户 ID</span>
          <input
            v-model.trim="merchantId"
            type="text"
            placeholder="可选"
            class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
            @keydown.enter.prevent="reload"
          />
        </label>
        <label class="grid min-w-[8rem] gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">状态</span>
          <select v-model="status" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
            <option value="">全部</option>
            <option value="0">待支付</option>
            <option value="1">成功</option>
            <option value="2">失败</option>
            <option value="3">已关闭</option>
          </select>
        </label>
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800"
          @click="runSearch"
        >
          查询
        </button>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1400px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">平台订单号</th>
              <th class="whitespace-nowrap px-4 py-3">商户</th>
              <th class="whitespace-nowrap px-4 py-3">金额</th>
              <th class="whitespace-nowrap px-4 py-3">手续费</th>
              <th class="whitespace-nowrap px-4 py-3">净额</th>
              <th class="whitespace-nowrap px-4 py-3">计费模式</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3">支付产品</th>
              <th class="whitespace-nowrap px-4 py-3">通道</th>
              <th class="whitespace-nowrap px-4 py-3">上游单号</th>
              <th class="whitespace-nowrap px-4 py-3">创建时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="11">加载中…</td>
            </tr>
            <tr v-else-if="!rows.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="11">暂无数据</td>
            </tr>
            <tr v-for="o in rows" v-else :key="o.order_no" class="hover:bg-slate-50/80">
              <td class="max-w-[14rem] px-4 py-3">
                <div class="truncate font-mono text-xs text-slate-900" :title="o.order_no">{{ o.order_no }}</div>
              </td>
              <td class="max-w-[13rem] px-4 py-3">
                <div
                  class="truncate font-mono text-xs font-medium text-slate-800"
                  :title="o.merchant_id"
                >
                  {{ o.merchant_id }}
                </div>
                <div
                  class="mt-0.5 truncate font-mono text-[11px] text-slate-500"
                  :title="o.merchant_order_no"
                >
                  {{ o.merchant_order_no }}
                </div>
              </td>
              <td class="px-4 py-3 tabular-nums text-slate-800">{{ formatYuan(o.amount) }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ formatYuan(o.fee_amount || 0) }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ formatYuan(o.net_amount || 0) }}</td>
              <td class="max-w-[8rem] px-4 py-3">
                <div class="truncate text-xs text-slate-600" :title="feeModeOptionLabel(o.fee_mode)">
                  {{ feeModeOptionLabel(o.fee_mode) }}
                </div>
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="statusClass(o.status)"
                >
                  {{ statusLabel(o.status) }}
                </span>
              </td>
              <td class="max-w-[9rem] px-4 py-3">
                <div
                  class="truncate font-mono text-xs text-slate-700"
                  :title="o.payin_product_code || ''"
                >
                  {{ o.payin_product_code || '—' }}
                </div>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">#{{ o.channel_id }}</td>
              <td class="max-w-[12rem] px-4 py-3">
                <div
                  class="truncate text-xs text-slate-600"
                  :title="o.upstream_trade_no || ''"
                >
                  {{ o.upstream_trade_no || '—' }}
                </div>
              </td>
              <td class="max-w-[11rem] px-4 py-3">
                <div class="truncate text-slate-600" :title="formatTime(o.created_at)">
                  {{ formatTime(o.created_at) }}
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <AdminPaginationBar
        v-if="!loading && (total > 0 || rows.length > 0)"
        v-model:page="page"
        v-model:pageSize="pageSize"
        :total="total"
        :page-count="pageCount"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'

import { feeModeOptionLabel } from '../../../lib/feeSemantics'
import { adminGet } from '../../../lib/adminApi'
import { formatAdminMoney } from '../../../lib/displaySettings'

import type { AdminOrderRow, AdminOrdersResp } from './types'

const props = withDefaults(defineProps<{
  title?: string
  description?: string
  endpoint?: string
}>(), {
  title: '全站订单',
  description: '跨商户检索平台订单（只读，MVP）；关键词匹配平台单号、商户单号或商户 ID（精确）。',
  endpoint: '/v1/admin/payin_orders',
})

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const keyword = ref('')
const merchantId = ref('')
const status = ref('')

const loading = ref(false)
const error = ref('')
const rows = ref<AdminOrderRow[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

function formatYuan(cents: number) {
  return formatAdminMoney(cents)
}

function formatTime(ts: number) {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime())
    ? '—'
    : d.toLocaleString('zh-CN', { hour12: false, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function statusLabel(s: number) {
  if (s === 0) return '待支付'
  if (s === 1) return '成功'
  if (s === 2) return '失败'
  if (s === 3) return '已关闭'
  return String(s)
}

function statusClass(s: number) {
  if (s === 1) return 'bg-emerald-100 text-emerald-800'
  if (s === 0) return 'bg-amber-100 text-amber-900'
  if (s === 2 || s === 3) return 'bg-rose-100 text-rose-800'
  return 'bg-slate-100 text-slate-700'
}

function buildQuery(): string {
  const q = new URLSearchParams()
  if (keyword.value) q.set('keyword', keyword.value)
  if (merchantId.value) q.set('merchant_id', merchantId.value)
  if (status.value !== '') q.set('status', status.value)
  q.set('limit', String(pageSize.value))
  q.set('offset', String((page.value - 1) * pageSize.value))
  const s = q.toString()
  return s ? `${props.endpoint}?${s}` : props.endpoint
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminGet<AdminOrdersResp>(buildQuery())
    rows.value = res.orders || []
    const rowN = rows.value.length
    const t = typeof res.total === 'number' && Number.isFinite(res.total) ? res.total : 0
    if (t > 0) {
      total.value = t
    } else if (rowN < pageSize.value) {
      total.value = (page.value - 1) * pageSize.value + rowN
    } else {
      total.value = page.value * pageSize.value + 1
    }
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    rows.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function runSearch() {
  if (page.value !== 1) page.value = 1
  else void reload()
}

watch(pageSize, () => {
  page.value = 1
})

watch([page, pageSize], () => void reload(), { immediate: true })

let unregister: (() => void) | null = null
onMounted(() => {
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
