<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">运维监控</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：展示网关进程<strong>存活探活</strong>（<code class="rounded bg-slate-100 px-1 py-0.5 font-mono text-xs">GET /health</code>）。QPS、错误率、RPC/队列/DB 拓扑等需对接可观测性平台，见下方说明。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="text-sm font-semibold text-slate-900">网关健康</div>
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-800 shadow-sm hover:bg-slate-50"
          @click="loadHealth"
        >
          重新检测
        </button>
      </div>

      <p v-if="error" class="mt-3 text-sm text-rose-600">{{ error }}</p>

      <div v-else class="mt-4 space-y-2 text-sm">
        <div v-if="loading" class="text-slate-500">检测中…</div>
        <template v-else-if="health">
          <div class="flex flex-wrap items-center gap-2">
            <span
              class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold"
              :class="health.status === 'ok' ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
            >
              {{ health.status === 'ok' ? '正常' : health.status }}
            </span>
            <span class="text-slate-600">服务：<span class="font-mono text-slate-800">{{ health.service }}</span></span>
          </div>
          <p v-if="health.timestamp_ms != null" class="font-mono text-xs text-slate-500">
            时间戳（ms）：{{ health.timestamp_ms }}
          </p>
          <pre class="mt-3 max-h-48 overflow-auto rounded-xl bg-slate-50 p-3 font-mono text-xs text-slate-800">{{ rawJson }}</pre>
        </template>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>各 RPC 子服务、NSQ、MySQL 的聚合健康与延迟</li>
        <li>错误率、慢查询与 Trace ID 检索</li>
      </ul>
      <p class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 font-mono text-xs text-amber-900">
        待接入：可观测性平台或统一 metrics API
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

type HealthResp = {
  status?: string
  service?: string
  timestamp_ms?: number
}

const loading = ref(false)
const error = ref('')
const health = ref<HealthResp | null>(null)
const rawBody = ref('')

const rawJson = computed(() => {
  if (!rawBody.value.trim()) return '—'
  try {
    return JSON.stringify(JSON.parse(rawBody.value), null, 2)
  } catch {
    return rawBody.value
  }
})

async function loadHealth() {
  loading.value = true
  error.value = ''
  health.value = null
  rawBody.value = ''
  try {
    const resp = await fetch('/health', { method: 'GET' })
    rawBody.value = await resp.text()
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`)
    }
    health.value = JSON.parse(rawBody.value) as HealthResp
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadHealth()
})
</script>
