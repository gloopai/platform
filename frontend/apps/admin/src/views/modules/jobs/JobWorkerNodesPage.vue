<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">Job 节点</h1>
      <p class="mt-1 text-sm text-slate-600">
        展示已向数据库上报心跳的 <code class="rounded bg-slate-100 px-1 font-mono text-xs">job-worker</code> 进程；<strong class="font-medium text-slate-800">执行中</strong>为当前
        <code class="font-mono text-xs">running</code> 任务数，<strong class="font-medium text-slate-800">近1h成功</strong>为过去一小时内该节点完成的任务数。
        <span class="block mt-1 text-slate-500">全局 <code class="font-mono">queued</code> 排队数表示尚未被任意 worker 认领的任务条数。</span>
      </p>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
      <div class="text-sm text-slate-600">
        队列中（全局）：
        <span class="font-mono font-semibold text-slate-900">{{ queuedTotal }}</span>
        条
      </div>
      <label class="flex cursor-pointer items-center gap-2 text-xs text-slate-600">
        <input v-model="autoRefresh" type="checkbox" class="rounded border-slate-300" />
        每 10 秒自动刷新
      </label>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[900px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">节点 ID（Worker.ID）</th>
              <th class="px-4 py-3">主机名</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">最近心跳</th>
              <th class="px-4 py-3">执行中</th>
              <th class="px-4 py-3">近1h成功</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td class="px-4 py-8 text-center text-slate-500" colspan="6">加载中...</td></tr>
            <tr v-else-if="!nodes.length"><td class="px-4 py-8 text-center text-slate-500" colspan="6">暂无心跳记录（请确认 job-worker 已启动且已执行过心跳）</td></tr>
            <tr v-for="n in nodes" :key="n.worker_id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs text-slate-900">{{ n.worker_id }}</td>
              <td class="px-4 py-3 text-slate-700">{{ n.hostname || '—' }}</td>
              <td class="px-4 py-3">
                <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="onlineClass(n)">{{ onlineLabel(n) }}</span>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatTs(n.last_heartbeat_at) }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ n.running_tasks }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ n.success_last_hour }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet } from '../../../lib/adminApi'

type Node = {
  worker_id: string
  hostname: string
  last_heartbeat_at: number
  running_tasks: number
  success_last_hour: number
}

const toast = useUiToast()
const loading = ref(false)
const nodes = ref<Node[]>([])
const queuedTotal = ref(0)
const autoRefresh = ref(true)

const STALE_SEC = 45

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN', { hour12: false })
}

function onlineLabel(n: Node): string {
  if (!n.last_heartbeat_at) return '未知'
  const age = Date.now() / 1000 - n.last_heartbeat_at
  return age <= STALE_SEC ? '在线' : '疑似离线'
}

function onlineClass(n: Node): string {
  if (!n.last_heartbeat_at) return 'bg-slate-100 text-slate-600'
  const age = Date.now() / 1000 - n.last_heartbeat_at
  return age <= STALE_SEC ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'
}

async function load() {
  loading.value = true
  try {
    const r = await adminGet<{ nodes: Node[]; queued_total: number }>('/v1/admin/job_workers')
    nodes.value = r.nodes || []
    queuedTotal.value = typeof r.queued_total === 'number' ? r.queued_total : 0
  } catch (e) {
    toast.error(`加载失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    loading.value = false
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

function startPoll() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (autoRefresh.value && document.visibilityState === 'visible') void load()
  }, 10000)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

watch(autoRefresh, (on) => {
  if (on) startPoll()
  else stopPoll()
})

onMounted(() => {
  void load()
  startPoll()
})
onUnmounted(() => stopPoll())
</script>
