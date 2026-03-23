<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">对账中心</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：按<strong>自然日</strong>展示平台侧订单聚合（与「系统概览」同源口径），作为与上游通道对账文件比对前的<strong>平台账</strong>快照。选择日期后加载；差异批次、文件与上游导入为后续迭代。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="flex flex-wrap items-end gap-3">
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-slate-700">对账日</span>
        <input
          v-model="dateStr"
          type="date"
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm"
          @change="load"
        />
      </label>
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
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
        @click="load"
      >
        加载
      </button>
      <router-link
        to="/orders"
        class="text-sm font-medium text-slate-700 underline decoration-slate-300 underline-offset-2 hover:text-slate-950"
      >
        全站订单明细
      </router-link>
    </div>

    <div v-if="loading" class="text-sm text-slate-500">加载中…</div>
    <template v-else-if="data">
      <p class="text-xs text-slate-500">
        数据日期：<span class="font-mono text-slate-700">{{ data.date }}</span>（按订单创建时间落在该日内）
      </p>
      <StatsKpiRow :totals="data.totals" />
      <StatsBreakdownTables :products="data.by_pay_product" :channels="data.by_channel" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'

import StatsBreakdownTables from '../stats/StatsBreakdownTables.vue'
import StatsKpiRow from '../stats/StatsKpiRow.vue'
import type { ReconcileDayOverview } from './types'

function todayLocalISODate(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const dateStr = ref(todayLocalISODate())
const merchantId = ref('')
const loading = ref(true)
const error = ref('')
const data = ref<ReconcileDayOverview | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const q = new URLSearchParams({ date: dateStr.value })
    if (merchantId.value) q.set('merchant_id', merchantId.value)
    data.value = await adminGet<ReconcileDayOverview>(`/v1/admin/reconcile/day?${q.toString()}`)
  } catch {
    error.value = '加载失败，请确认日期格式为 YYYY-MM-DD 且已登录'
    data.value = null
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
