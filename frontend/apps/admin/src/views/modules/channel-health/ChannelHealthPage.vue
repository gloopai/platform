<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">通道监控</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：展示上游通道的<strong>启用</strong>与<strong>熔断</strong>状态及路由侧汇总；成功率、延迟、时序曲线等需对接指标服务，见
        <router-link to="/routing" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          路由策略
        </router-link>
        页「后续规划」。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <section>
      <h2 class="mb-3 text-sm font-semibold text-slate-800">路由与配置汇总</h2>
      <RoutingStatGrid :summary="summary" :loading="summaryLoading" />
    </section>

    <section>
      <div class="mb-3 flex flex-wrap items-end justify-between gap-3">
        <h2 class="text-sm font-semibold text-slate-800">上游通道状态</h2>
        <router-link
          to="/channels"
          class="text-sm font-medium text-slate-700 underline decoration-slate-300 underline-offset-2 hover:text-slate-950"
        >
          去通道管理编辑
        </router-link>
      </div>

      <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full min-w-[720px] text-left text-sm">
            <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th class="whitespace-nowrap px-4 py-3">ID</th>
                <th class="whitespace-nowrap px-4 py-3">名称</th>
                <th class="whitespace-nowrap px-4 py-3">DriverKey</th>
                <th class="whitespace-nowrap px-4 py-3">代收</th>
                <th class="whitespace-nowrap px-4 py-3">代付</th>
                <th class="whitespace-nowrap px-4 py-3">启用</th>
                <th class="whitespace-nowrap px-4 py-3">熔断跳过</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="channelsLoading">
                <td class="px-4 py-8 text-center text-slate-500" colspan="7">加载中…</td>
              </tr>
              <tr v-else-if="!channels.length">
                <td class="px-4 py-10 text-center text-slate-500" colspan="7">暂无通道数据</td>
              </tr>
              <tr v-for="c in channels" v-else :key="c.id" class="hover:bg-slate-50/80">
                <td class="px-4 py-3 font-mono text-slate-800">#{{ c.id }}</td>
                <td class="px-4 py-3 font-medium text-slate-900">{{ c.name }}</td>
                <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ c.driver_key || '—' }}</td>
                <td class="px-4 py-3">
                  <span :class="c.supports_payin ? 'text-emerald-700' : 'text-slate-400'">
                    {{ c.supports_payin ? '是' : '否' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span :class="c.supports_payout ? 'text-emerald-700' : 'text-slate-400'">
                    {{ c.supports_payout ? '是' : '否' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                    :class="c.enabled ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                  >
                    {{ c.enabled ? '启用' : '停用' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                    :class="c.fuse_enabled ? 'bg-amber-100 text-amber-900' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ c.fuse_enabled ? '是（路由跳过）' : '否' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'
import RoutingStatGrid from '../routing/RoutingStatGrid.vue'
import type { AdminChannel } from '../channels/types'
import type { RoutingSummary } from '../routing/types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const summaryLoading = ref(false)
const summary = ref<RoutingSummary | null>(null)

const channelsLoading = ref(false)
const channels = ref<AdminChannel[]>([])

const error = ref('')

async function loadSummary() {
  summaryLoading.value = true
  try {
    summary.value = await adminGet<RoutingSummary>('/v1/admin/routing/summary')
  } catch {
    error.value = '加载路由汇总失败'
    summary.value = null
  } finally {
    summaryLoading.value = false
  }
}

async function loadChannels() {
  channelsLoading.value = true
  try {
    const data = await adminGet<{ channels: AdminChannel[] }>('/v1/admin/channels')
    channels.value = data.channels || []
  } catch {
    if (!error.value) error.value = '加载通道列表失败'
    channels.value = []
  } finally {
    channelsLoading.value = false
  }
}

async function load() {
  error.value = ''
  await Promise.all([loadSummary(), loadChannels()])
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
