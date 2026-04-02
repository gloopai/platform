<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">定时任务</h1>
      <p class="mt-1 text-sm text-slate-600">系统默认任务 + 自定义任务；调度写入队列，由 job-worker 消费执行（可多实例横向扩展）。</p>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800"
        @click="openCreate"
      >
        新增任务
      </button>
      <div class="flex flex-wrap items-center gap-2 text-sm text-slate-600">
        <span>每页</span>
        <select v-model.number="limit" class="rounded-lg border border-slate-200 px-2 py-1.5" @change="onPageSizeChange">
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
        </select>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1100px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">任务</th>
              <th class="px-4 py-3">分类</th>
              <th class="px-4 py-3">类型</th>
              <th class="px-4 py-3">调度</th>
              <th class="px-4 py-3">下次执行</th>
              <th class="px-4 py-3">上次状态</th>
              <th class="px-4 py-3">启用</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td class="px-4 py-8 text-center text-slate-500" colspan="8">加载中...</td></tr>
            <tr v-else-if="!jobs.length"><td class="px-4 py-8 text-center text-slate-500" colspan="8">暂无任务</td></tr>
            <tr v-for="j in jobs" :key="j.id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3">
                <div class="font-medium text-slate-900">{{ j.name }}</div>
                <div class="font-mono text-xs text-slate-500">{{ j.job_key }}</div>
              </td>
              <td class="px-4 py-3 text-slate-700">{{ j.category || '—' }}</td>
              <td class="px-4 py-3">{{ j.builtin ? '系统默认' : '自定义' }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ scheduleText(j) }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatTs(j.next_run_at) }}</td>
              <td class="px-4 py-3">
                <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="statusClass(j.last_status)">
                  {{ j.last_status || '—' }}
                </span>
                <div v-if="j.last_error" class="mt-1 max-w-xs truncate text-xs text-rose-600" :title="j.last_error">{{ j.last_error }}</div>
              </td>
              <td class="px-4 py-3">
                <span :class="j.enabled ? 'text-emerald-600' : 'text-slate-500'">{{ j.enabled ? '是' : '否' }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-2">
                  <button type="button" class="rounded border border-slate-200 bg-white px-2 py-1 text-xs" @click="openEdit(j)">编辑</button>
                  <button
                    type="button"
                    class="rounded border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs text-indigo-700"
                    title="立刻入队并由 job-worker 抢占执行，不等待 next_run_at；未启用调度时也可手动跑一次"
                    @click="runNow(j.id)"
                  >
                    立即执行
                  </button>
                  <button type="button" class="rounded border border-amber-200 bg-amber-50 px-2 py-1 text-xs text-amber-700" @click="toggle(j)">
                    {{ j.enabled ? '停用' : '启用' }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-2 border-t border-slate-100 px-4 py-3 text-xs text-slate-600">
        <div>共 {{ total }} 条 · 第 {{ currentPage }} / {{ totalPages }} 页</div>
        <div class="flex items-center gap-2">
          <button type="button" class="rounded border border-slate-200 px-3 py-1 disabled:opacity-40" :disabled="offset <= 0" @click="prev">
            上一页
          </button>
          <button type="button" class="rounded border border-slate-200 px-3 py-1 disabled:opacity-40" :disabled="offset + limit >= total" @click="next">
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 新增 / 编辑 -->
    <div v-if="showDialog" class="modal modal-open" role="dialog" aria-modal="true">
      <div class="modal-box max-h-[90vh] w-11/12 max-w-2xl overflow-y-auto rounded-2xl border border-slate-200 bg-white p-5 shadow-2xl">
        <h2 class="text-lg font-semibold text-slate-900">{{ editingId > 0 ? '编辑任务' : '新增任务' }}</h2>
        <p class="mt-2 text-sm leading-relaxed text-slate-600">
          <strong class="text-slate-800">执行位置：</strong>
          调度与队列在 job-worker（可多副本部署）。任务通过 <code class="rounded bg-slate-100 px-1 font-mono text-xs">job_key</code> 在 worker 内注册到具体处理函数；需要调其它微服务时，在对应 handler 里发 gRPC/HTTP 即可。
          <span class="mt-1 block">
            <strong class="text-slate-800">多节点：</strong>
            多个 job-worker 从同一队列表抢占执行（不同实例 <code class="font-mono text-xs">worker_id</code> 不同）。同一
            <code class="font-mono text-xs">job_key</code> 的多次触发会被各节点分摊处理；默认并发策略 forbid 保证同一任务同一时刻仅一条 run 在执行。扩容可加 worker 副本数，或拆成多个 job_key。
          </span>
        </p>
        <div v-if="editingId > 0" class="mt-4 rounded-lg border border-slate-100 bg-slate-50/90 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <span class="text-xs font-medium text-slate-600">job_key（与 job-worker 里注册的 key 一致，此处只读）</span>
            <button
              type="button"
              class="rounded border border-slate-200 bg-white px-2 py-0.5 text-xs text-slate-700 hover:bg-slate-50"
              @click="copyJobKey"
            >
              复制
            </button>
          </div>
          <div class="mt-1.5 break-all font-mono text-sm text-slate-900">{{ formJobKey || '—' }}</div>
        </div>
        <div v-else class="mt-4 space-y-3">
          <div class="rounded-lg border border-dashed border-slate-200 bg-slate-50/60 px-3 py-2 text-xs leading-relaxed text-slate-600">
            下拉选项来自<strong class="text-slate-800">服务端代码</strong>（<code class="font-mono">jobkeys</code> 与 job-worker 里
            <code class="font-mono">registerJob</code> 一致），<strong class="text-slate-800">不需要</strong>先在数据库里插任务才能看到选项。保存新建时才会写入
            <code class="font-mono">scheduled_jobs</code>。新 key 先在代码里注册并部署后再用「自定义」填写。
            <span v-if="keyPattern" class="mt-1 block text-slate-500">格式：{{ keyPattern }}</span>
          </div>
          <label class="flex flex-col gap-1 text-sm">
            <span class="text-slate-700">job_key</span>
            <select v-model="jobKeyPick" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" @change="onJobKeyPick">
              <option value="">请选择</option>
              <option v-for="k in knownKeys" :key="k" :value="k">{{ k }}</option>
              <option value="__custom__">自定义（须已在 runner 注册）</option>
            </select>
          </label>
          <label v-if="jobKeyPick === '__custom__'" class="flex flex-col gap-1 text-sm">
            <span class="text-slate-700">自定义 job_key</span>
            <input
              v-model.trim="form.job_key"
              type="text"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="小写字母开头，如 my_sync_task"
              autocomplete="off"
            />
          </label>
        </div>
        <div class="mt-4 grid gap-3 md:grid-cols-2">
          <label class="flex flex-col gap-1 text-sm md:col-span-2">
            <span class="text-slate-700">任务名称</span>
            <input v-model.trim="form.name" type="text" class="rounded-lg border border-slate-200 px-3 py-2" placeholder="任务名称" />
          </label>
          <label class="flex flex-col gap-1 text-sm">
            <span class="text-slate-700">执行策略</span>
            <select v-model="form.schedule_type" class="rounded-lg border border-slate-200 px-3 py-2">
              <option value="fixed_interval">每多少秒执行一次</option>
              <option value="hourly">每小时第几分钟</option>
              <option value="daily">每天固定时刻</option>
            </select>
          </label>
          <label class="flex flex-col gap-1 text-sm">
            <span class="text-slate-700">
              {{ form.schedule_type === 'fixed_interval' ? '间隔秒' : form.schedule_type === 'hourly' ? '分钟(0-59)' : '时间(HH:MM)' }}
            </span>
            <input
              v-if="form.schedule_type === 'fixed_interval'"
              v-model.number="form.interval_seconds"
              type="number"
              min="1"
              class="rounded-lg border border-slate-200 px-3 py-2"
            />
            <input
              v-else
              v-model.trim="form.cron_expr"
              type="text"
              class="rounded-lg border border-slate-200 px-3 py-2"
              :placeholder="form.schedule_type === 'hourly' ? '如 15' : '如 03:30'"
            />
          </label>
        </div>
        <label class="mt-3 flex flex-col gap-1 text-sm">
          <span class="text-slate-700">任务参数（可选 JSON）</span>
          <textarea v-model="form.payload_json" rows="4" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-xs" placeholder='{"batch_size":200}' />
        </label>
        <div class="mt-6 flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm" @click="closeDialog">取消</button>
          <button type="button" class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800" @click="save">保存</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" class="h-full min-h-[100dvh] w-full cursor-default bg-transparent" aria-label="关闭" @click="closeDialog" />
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'
import { useUiToast } from '../../../composables/useUiToast'
import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

type Job = {
  id: number
  job_key: string
  name: string
  category: string
  builtin: boolean
  enabled: boolean
  interval_seconds: number
  schedule_type: string
  cron_expr: string
  payload_json: string
  next_run_at: number
  last_status: string
  last_error: string
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const toast = useUiToast()
const loading = ref(false)
const jobs = ref<Job[]>([])
const total = ref(0)
const limit = ref(20)
const offset = ref(0)

const showDialog = ref(false)
const editingId = ref(0)
/** 仅编辑时展示：后端生成的 job_key，不在此表单修改 */
const formJobKey = ref('')
const knownKeys = ref<string[]>([])
const keyPattern = ref('')
const jobKeyPick = ref('')
const form = ref({
  name: '',
  job_key: '',
  interval_seconds: 60,
  schedule_type: 'fixed_interval',
  cron_expr: '',
  payload_json: '{}',
})

const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit.value)))

function formatTs(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN', { hour12: false })
}

function statusClass(s: string): string {
  if (s === 'success') return 'bg-emerald-100 text-emerald-700'
  if (s === 'failed') return 'bg-rose-100 text-rose-700'
  if (s === 'running') return 'bg-indigo-100 text-indigo-700'
  if (s === 'skipped') return 'bg-amber-100 text-amber-700'
  return 'bg-slate-100 text-slate-700'
}

function resetForm() {
  editingId.value = 0
  formJobKey.value = ''
  jobKeyPick.value = ''
  form.value = {
    name: '',
    job_key: '',
    interval_seconds: 60,
    schedule_type: 'fixed_interval',
    cron_expr: '',
    payload_json: '{}',
  }
}

function onJobKeyPick() {
  if (jobKeyPick.value === '__custom__') {
    form.value.job_key = ''
    return
  }
  form.value.job_key = jobKeyPick.value
}

function openCreate() {
  resetForm()
  showDialog.value = true
}

function openEdit(j: Job) {
  editingId.value = j.id
  formJobKey.value = j.job_key || ''
  form.value = {
    job_key: j.job_key || '',
    name: j.name,
    interval_seconds: j.interval_seconds,
    schedule_type: j.schedule_type || 'fixed_interval',
    cron_expr: j.cron_expr || '',
    payload_json: j.payload_json || '{}',
  }
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  resetForm()
}

async function copyJobKey() {
  const t = formJobKey.value.trim()
  if (!t) {
    toast.error('无 job_key')
    return
  }
  try {
    await navigator.clipboard.writeText(t)
    toast.success('已复制 job_key')
  } catch {
    toast.error('复制失败，请手动选择复制')
  }
}

function scheduleText(j: Job): string {
  if (j.schedule_type === 'hourly') return `每小时 ${j.cron_expr || '0'} 分`
  if (j.schedule_type === 'daily') return `每天 ${j.cron_expr || '00:00'}`
  return `每 ${j.interval_seconds || 60}s`
}

async function loadJobKeys() {
  try {
    const r = await adminGet<{ keys: string[]; pattern: string }>('/v1/admin/jobs/keys')
    knownKeys.value = Array.isArray(r.keys) ? r.keys : []
    keyPattern.value = r.pattern || ''
  } catch {
    knownKeys.value = []
  }
}

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams()
    q.set('limit', String(limit.value))
    q.set('offset', String(offset.value))
    let r = await adminGet<{ jobs: Job[]; total: number }>(`/v1/admin/jobs?${q.toString()}`)
    let t = typeof r.total === 'number' ? r.total : 0
    if (t > 0 && offset.value >= t) {
      offset.value = Math.max(0, Math.floor((t - 1) / limit.value) * limit.value)
      q.set('offset', String(offset.value))
      r = await adminGet<{ jobs: Job[]; total: number }>(`/v1/admin/jobs?${q.toString()}`)
      t = typeof r.total === 'number' ? r.total : 0
    }
    jobs.value = r.jobs || []
    total.value = t
  } catch (e) {
    toast.error(`任务加载失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    loading.value = false
  }
}

function onPageSizeChange() {
  offset.value = 0
  void load()
}

function prev() {
  offset.value = Math.max(0, offset.value - limit.value)
  void load()
}

function next() {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    void load()
  }
}

async function save() {
  if (!form.value.name.trim()) {
    toast.error('任务名称不能为空')
    return
  }
  if (editingId.value <= 0) {
    const jk = form.value.job_key.trim()
    if (!jk) {
      toast.error('请选择或填写 job_key')
      return
    }
  }
  try {
    if (editingId.value > 0) {
      await adminPut(`/v1/admin/jobs/${editingId.value}`, {
        name: form.value.name.trim(),
        interval_seconds: form.value.interval_seconds,
        schedule_type: form.value.schedule_type,
        cron_expr: form.value.cron_expr.trim(),
        payload_json: form.value.payload_json,
      })
      toast.success('任务已更新')
    } else {
      await adminPost('/v1/admin/jobs', {
        job_key: form.value.job_key.trim(),
        name: form.value.name.trim(),
        category: 'custom',
        enabled: true,
        schedule_type: form.value.schedule_type,
        cron_expr: form.value.cron_expr.trim(),
        interval_seconds: form.value.interval_seconds,
        timezone: 'Asia/Shanghai',
        payload_json: form.value.payload_json,
        concurrency_policy: 'forbid',
        misfire_policy: 'run_once',
      })
      toast.success('任务已创建')
    }
    closeDialog()
    offset.value = 0
    await load()
  } catch (e) {
    toast.error(`保存失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

async function toggle(j: Job) {
  try {
    await adminPost(`/v1/admin/jobs/${j.id}/toggle`, { enabled: !j.enabled })
    toast.success(j.enabled ? '任务已停用' : '任务已启用')
    await load()
  } catch (e) {
    toast.error(`操作失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

async function runNow(id: number) {
  try {
    await adminPost(`/v1/admin/jobs/${id}/run`, {})
    toast.success('已入队，worker 将尽快执行（不等待定时）')
  } catch (e) {
    toast.error(`执行失败：${e instanceof Error ? e.message : String(e)}`)
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void loadJobKeys()
  void load()
  if (registerRefresh) unregister = registerRefresh(() => void load())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
