<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">全站订单</h1>
      <p class="mt-1 text-sm text-slate-600">
        跨商户检索平台订单（只读，MVP）；关键词匹配平台单号、商户单号或商户 ID（精确）。
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
        <label class="grid w-24 gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">条数</span>
          <select v-model.number="limit" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
            <option :value="30">30</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </label>
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800"
          @click="reload"
        >
          查询
        </button>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[960px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">平台订单号</th>
              <th class="whitespace-nowrap px-4 py-3">商户</th>
              <th class="whitespace-nowrap px-4 py-3">金额</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3">支付产品</th>
              <th class="whitespace-nowrap px-4 py-3">通道</th>
              <th class="whitespace-nowrap px-4 py-3">上游单号</th>
              <th class="whitespace-nowrap px-4 py-3">创建时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中…</td>
            </tr>
            <tr v-else-if="!rows.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="8">暂无数据</td>
            </tr>
            <tr v-for="o in rows" v-else :key="o.order_no" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs text-slate-900">{{ o.order_no }}</td>
              <td class="px-4 py-3">
                <div class="font-mono text-xs font-medium text-slate-800">{{ o.merchant_id }}</div>
                <div class="mt-0.5 font-mono text-[11px] text-slate-500">{{ o.merchant_order_no }}</div>
              </td>
              <td class="px-4 py-3 tabular-nums text-slate-800">{{ formatYuan(o.amount) }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="statusClass(o.status)"
                >
                  {{ statusLabel(o.status) }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ o.pay_product_code || '—' }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">#{{ o.channel_id }}</td>
              <td class="px-4 py-3 text-xs text-slate-600">{{ o.upstream_trade_no || '—' }}</td>
              <td class="px-4 py-3 text-slate-600">{{ formatTime(o.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'

import type { AdminOrderRow, AdminOrdersResp } from './types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const keyword = ref('')
const merchantId = ref('')
const status = ref('')
const limit = ref(50)

const loading = ref(false)
const error = ref('')
const rows = ref<AdminOrderRow[]>([])

function formatYuan(cents: number) {
  return `¥ ${(cents / 100).toFixed(2)}`
}

function formatTime(ts: number) {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return d.toLocaleString()
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
  if (limit.value > 0) q.set('limit', String(limit.value))
  const s = q.toString()
  return s ? `/v1/admin/orders?${s}` : '/v1/admin/orders'
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminGet<AdminOrdersResp>(buildQuery())
    rows.value = res.orders || []
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    rows.value = []
  } finally {
    loading.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void reload()
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
