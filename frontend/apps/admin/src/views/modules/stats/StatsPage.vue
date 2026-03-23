<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统概览</h1>
      <p class="mt-1 text-sm text-slate-600">
        今日核心指标与按 <strong>支付产品</strong> / <strong>上游通道</strong> 拆解（自然日，与订单创建时间一致）。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div v-if="loading" class="text-sm text-slate-500">加载中…</div>
    <template v-else-if="data">
      <StatsKpiRow :totals="data.totals" />
      <StatsStatusStrip
        :totals="data.totals"
        :enabled-channels="data.enabled_channels"
        :fused-channels="data.fused_channels"
      />
      <StatsBreakdownTables :products="data.by_payin_product" :channels="data.by_channel" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'

import StatsBreakdownTables from './StatsBreakdownTables.vue'
import StatsKpiRow from './StatsKpiRow.vue'
import StatsStatusStrip from './StatsStatusStrip.vue'
import type { StatsOverview } from './types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const loading = ref(true)
const error = ref('')
const data = ref<StatsOverview | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await adminGet<StatsOverview>('/v1/admin/stats/overview')
  } catch {
    error.value = '加载统计失败，请检查登录态与网关'
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
