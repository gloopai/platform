<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">任务日志</h1>
      <p class="mt-1 text-sm text-slate-600">
        记录每次执行；<strong class="font-medium text-slate-800">queued</strong> 表示已入队待
        <code class="rounded bg-slate-100 px-1 font-mono text-xs">job-worker</code> 抢占，
        <strong class="font-medium text-slate-800">running</strong> 起会显示占用该条目的 worker 节点 ID。若长期停在 queued，请确认进程已启动且任务已启用。
      </p>
    </div>

    <div class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-4">
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">任务 ID</span>
        <input v-model.number="jobId" type="number" min="0" class="rounded-lg border border-slate-200 px-3 py-2" placeholder="可选" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">状态</span>
        <select v-model="status" class="rounded-lg border border-slate-200 px-3 py-2">
          <option value="">全部</option>
          <option value="queued">queued</option>
          <option value="running">running</option>
          <option value="success">success</option>
          <option value="failed">failed</option>
          <option value="skipped">skipped</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">触发来源</span>
        <select v-model="triggerType" class="rounded-lg border border-slate-200 px-3 py-2">
          <option value="">全部</option>
          <option value="scheduler">scheduler</option>
          <option value="manual">manual</option>
          <option value="retry">retry</option>
        </select>
      </label>
      <div class="flex items-end gap-2">
        <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm" @click="reload">加载</button>
        <label class="flex cursor-pointer items-center gap-2 text-xs text-slate-600">
          <input v-model="autoRefresh" type="checkbox" class="rounded border-slate-300" />
          每 8 秒自动刷新
        </label>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1480px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">ID</th>
              <th class="px-4 py-3">任务</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">执行节点</th>
              <th class="px-4 py-3">触发</th>
              <th class="px-4 py-3">耗时</th>
              <th class="px-4 py-3">摘要</th>
              <th class="px-4 py-3">错误</th>
              <th class="px-4 py-3">时间</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td class="px-4 py-8 text-center text-slate-500" colspan="10">加载中...</td></tr>
            <tr v-else-if="!runs.length"><td class="px-4 py-8 text-center text-slate-500" colspan="10">暂无日志</td></tr>
            <tr v-for="r in runs" :key="r.id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs">{{ r.id }}</td>
              <td class="px-4 py-3">
                <div class="font-medium">{{ r.job_name }}</div>
                <div class="font-mono text-xs text-slate-500">{{ r.job_key }}</div>
              </td>
              <td class="px-4 py-3"><span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="statusClass(r.status)">{{ r.status }}</span></td>
              <td class="max-w-[220px] px-4 py-3 font-mono text-xs text-slate-700" :title="workerTitle(r)">
                {{ workerLabel(r) }}
              </td>
              <td class="px-4 py-3">{{ r.trigger_type }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatDuration(r) }}</td>
              <td class="px-4 py-3 text-slate-700">{{ r.summary || '—' }}</td>
              <td class="max-w-xs px-4 py-3">
                <div v-if="r.error_code" class="font-mono text-[11px] text-slate-500">{{ r.error_code }}</div>
                <div class="text-rose-600">{{ r.error_message || '—' }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatTs(r.started_at || r.scheduled_at) }}</td>
              <td class="px-4 py-3">
                <button v-if="r.status === 'failed'" type="button" class="rounded border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs text-indigo-700" @click="retryRun(r.id)">重试</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="flex items-center justify-between border-t border-slate-100 px-4 py-3 text-xs text-slate-600">
        <div>总计 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button type="button" class="rounded border border-slate-200 px-3 py-1 disabled:opacity-40" :disabled="offset <= 0" @click="prev">上一页</button>
          <button type="button" class="rounded border border-slate-200 px-3 py-1 disabled:opacity-40" :disabled="offset + limit >= total" @click="next">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref, watch } from 'vue'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPost } from '../../../lib/adminApi'

type Run = {
  id: number
  job_id: number
  job_key: string
  job_name: string
  trigger_type: string
  status: string
  duration_ms: number
  summary: string
  error_code: string
  error_message: string
  worker_id: string
  scheduled_at: number
  started_at: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const toast = useUiToast()
const loading = ref(false)
const runs = ref<Run[]>([])
const total = ref(0)
const jobId = ref<number | null>(null)
const status = ref('')
const triggerType = ref('')
const limit = ref(20)
const offset = ref(0)
const autoRefresh = ref(true)

let pollTimer: ReturnType<typeof setInterval> | null = null

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN', { hour12: false })
}

function formatDuration(r: Run): string {
  if (r.status === 'queued') return '—'
  if (r.status === 'running') return '—'
  if (r.duration_ms > 0) return `${r.duration_ms}ms`
  return '—'
}

function workerLabel(r: Run): string {
  const w = (r.worker_id || '').trim()
  if (r.status === 'queued') return w || '待领取（无 worker）'
  if (r.status === 'running') return w || '执行中…'
  return w || '—'
}

function workerTitle(r: Run): string {
  if (r.status === 'queued') {
    return 'queued 时尚未分配 worker；job-worker 认领后写入 worker_id。若长时间不变，请检查 job-worker 是否运行、任务是否启用、是否存在卡死的 running 记录。'
  }
  return r.worker_id || ''
}

function statusClass(s: string): string {
  if (s === 'success') return 'bg-emerald-100 text-emerald-700'
  if (s === 'failed') return 'bg-rose-100 text-rose-700'
  if (s === 'running') return 'bg-indigo-100 text-indigo-700'
  if (s === 'skipped') return 'bg-amber-100 text-amber-700'
  return 'bg-slate-100 text-slate-700'
}

async function reload() {
  loading.value = true
  try {
    const q = new URLSearchParams()
    if (jobId.value && jobId.value > 0) q.set('job_id', String(jobId.value))
    if (status.value) q.set('status', status.value)
    if (triggerType.value) q.set('trigger_type', triggerType.value)
    q.set('limit', String(limit.value))
    q.set('offset', String(offset.value))
    const r = await adminGet<{ runs: Run[]; total: number }>(`/v1/admin/job_runs?${q.toString()}`)
    runs.value = r.runs || []
    total.value = r.total || 0
  } catch (e) {
    toast.error(`日志加载失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    loading.value = false
  }
}

async function retryRun(id: number) {
  try {
    await adminPost(`/v1/admin/job_runs/${id}/retry`, {})
    toast.success('已加入重试队列')
    await reload()
  } catch (e) {
    toast.error(`重试失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

function prev() {
  offset.value = Math.max(0, offset.value - limit.value)
  void reload()
}

function next() {
  offset.value += limit.value
  void reload()
}

function startPoll() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (autoRefresh.value && document.visibilityState === 'visible') void reload()
  }, 8000)
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

let unregister: (() => void) | null = null
onMounted(() => {
  void reload()
  startPoll()
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  stopPoll()
  if (unregister) unregister()
})
</script>
