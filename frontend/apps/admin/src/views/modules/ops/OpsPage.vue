<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">运维监控</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        展示服务健康与实例（节点）状态：Consul 注册状态 + 后端就绪探测（部分服务）。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="text-sm font-semibold text-slate-900">服务状态</div>
        <div class="flex flex-wrap items-center gap-3">
          <label class="inline-flex items-center gap-2 text-sm text-slate-700">
            <input v-model="autoRefresh" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
            自动刷新（10s）
          </label>
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
            @click="load(false)"
          >
            重新检测
          </button>
        </div>
      </div>

      <p v-if="error" class="mt-3 text-sm text-rose-600">{{ error }}</p>

      <div v-else class="mt-4 space-y-2 text-sm">
        <div v-if="loading" class="text-slate-500">检测中…</div>
        <template v-else>
          <div v-if="refreshing" class="text-xs text-slate-400">后台刷新中…</div>
          <div class="flex flex-wrap items-center gap-2">
            <span
              class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold"
              :class="data?.ok ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
            >
              {{ data?.ok ? '整体正常' : '存在异常' }}
            </span>
            <span class="text-slate-600">接口：<span class="font-mono text-slate-800">GET /v1/admin/ops/services</span></span>
          </div>

          <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="s in data?.services || []"
              :key="s.service_name"
              class="cursor-pointer rounded-2xl border bg-white p-4 shadow-sm transition"
              :class="
                selectedServiceName === s.service_name
                  ? 'border-indigo-300 ring-2 ring-indigo-100'
                  : 'border-slate-200 hover:border-slate-300'
              "
              @click="selectedServiceName = s.service_name"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="truncate text-sm font-semibold text-slate-900">{{ s.service_name }}</div>
                  <div class="mt-1 text-xs text-slate-500">
                    实例：{{ s.total }}（pass {{ s.passing }} / warn {{ s.warning }} / crit {{ s.critical }}）
                  </div>
                </div>
                <span
                  class="inline-flex shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="s.ok ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
                >
                  {{ s.ok ? 'OK' : 'BAD' }}
                </span>
              </div>
            </div>
          </div>

          <div v-if="selectedService" class="mt-6">
            <div class="mb-2 text-sm font-semibold text-slate-800">{{ selectedService.service_name }} 实例</div>
            <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
              <div class="overflow-x-auto">
                <table class="w-full min-w-[720px] text-left text-sm">
                  <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    <tr>
                      <th class="whitespace-nowrap px-4 py-3">状态</th>
                      <th class="whitespace-nowrap px-4 py-3">Node</th>
                      <th class="whitespace-nowrap px-4 py-3">ServiceID</th>
                      <th class="whitespace-nowrap px-4 py-3">地址</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-if="!selectedService.nodes?.length">
                      <td class="px-4 py-8 text-center text-slate-500" colspan="4">无实例（Consul 未发现）</td>
                    </tr>
                    <tr v-for="n in selectedService.nodes || []" v-else :key="n.service_id" class="hover:bg-slate-50/80">
                      <td class="px-4 py-3">
                        <span
                          class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                          :class="
                            n.status === 'passing'
                              ? 'bg-emerald-100 text-emerald-800'
                              : n.status === 'warning'
                                ? 'bg-amber-100 text-amber-900'
                                : 'bg-rose-100 text-rose-800'
                          "
                        >
                          {{ n.status }}
                        </span>
                      </td>
                      <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ n.node || '—' }}</td>
                      <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ n.service_id }}</td>
                      <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ n.address }}:{{ n.port }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
          <div v-else class="mt-6 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-500">暂无服务数据</div>
        </template>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>接入 metrics：QPS、错误率、延迟分位、队列积压</li>
        <li>实例维度：就绪检查（ready）/ 版本号 / 启动时间</li>
      </ul>
      <p class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 font-mono text-xs text-amber-900">
        待接入：可观测性平台或统一 metrics API
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'

const loading = ref(false)
const refreshing = ref(false)
const error = ref('')
const autoRefresh = ref(true)
const data = ref<any | null>(null)
const selectedServiceName = ref('')

const selectedService = computed(() => {
  const services = (data.value?.services || []) as Array<any>
  if (!services.length) return null
  return services.find((s) => s.service_name === selectedServiceName.value) || services[0]
})

async function load(silent = false) {
  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
    error.value = ''
  }
  try {
    const next = await adminGet<any>('/v1/admin/ops/services')
    data.value = next
    const names = (next?.services || []).map((s: any) => s.service_name)
    if (!selectedServiceName.value || !names.includes(selectedServiceName.value)) {
      selectedServiceName.value = names[0] || ''
    }
  } catch (e) {
    if (!silent) {
      error.value = e instanceof Error ? e.message : String(e)
      data.value = null
      selectedServiceName.value = ''
    }
  } finally {
    if (silent) {
      refreshing.value = false
    } else {
      loading.value = false
    }
  }
}

let timer: number | null = null
onMounted(() => {
  void load(false)
  timer = window.setInterval(() => {
    if (!autoRefresh.value) return
    void load(true)
  }, 10_000)
})

onUnmounted(() => {
  if (timer != null) window.clearInterval(timer)
})
</script>
