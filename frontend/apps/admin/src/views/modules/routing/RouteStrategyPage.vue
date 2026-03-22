<template>
  <div class="grid gap-6">
    <header class="rounded-xl border border-slate-200 bg-gradient-to-br from-slate-50 to-white px-6 py-5 shadow-sm">
      <h1 class="text-lg font-semibold text-slate-900">路由策略</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        当前版本将「策略」落实在数据与交易服务中：<strong>产品内多通道加权</strong>、<strong>商户可用产品</strong>、<strong>通道熔断与限额</strong>。以下为算法说明、实时汇总与配置入口；规则引擎、按成功率动态调度等仍在规划中。
      </p>
      <div v-if="summary" class="mt-4 rounded-lg border border-slate-200 bg-white/80 px-4 py-3">
        <div class="text-xs font-medium uppercase tracking-wide text-slate-500">当前选用算法</div>
        <div class="mt-1 font-mono text-sm text-slate-800">{{ summary.algorithm_key }}</div>
        <div class="mt-1 text-sm text-slate-600">{{ summary.algorithm_label }}</div>
      </div>
    </header>

    <section>
      <h2 class="mb-3 text-sm font-semibold text-slate-800">配置与数据概览</h2>
      <RoutingStatGrid :summary="summary" :loading="loading" />
      <p v-if="error" class="mt-3 text-sm text-rose-600">{{ error }}</p>
    </section>

    <section>
      <h2 class="mb-3 text-sm font-semibold text-slate-800">策略落在哪里</h2>
      <RoutingConfigLinks />
    </section>

    <section class="rounded-xl border border-dashed border-slate-200 bg-slate-50/80 px-5 py-4 text-sm text-slate-600">
      <div class="font-medium text-slate-800">后续规划（尚未提供单独配置页）</div>
      <ul class="mt-2 list-inside list-disc space-y-1">
        <li>按商户 / 金额段 / 时段覆盖权重或指定主备通道</li>
        <li>按成功率、延迟自动降权或熔断恢复策略</li>
        <li>与「通道监控」联动的告警与人工切换流水</li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'

import RoutingConfigLinks from './RoutingConfigLinks.vue'
import RoutingStatGrid from './RoutingStatGrid.vue'
import type { RoutingSummary } from './types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const loading = ref(false)
const error = ref('')
const summary = ref<RoutingSummary | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await adminGet<RoutingSummary>('/v1/admin/routing/summary')
  } catch {
    error.value = '加载路由概览失败，请检查登录态与网关'
    summary.value = null
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
