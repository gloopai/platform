<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">操作日志</h1>
      <p class="mt-1 text-sm text-slate-600">审计后台管理员操作，记录操作者、IP、接口、权限点与执行结果。</p>
    </div>

    <div class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-4">
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">开始时间</span>
        <input v-model="startLocal" type="datetime-local" class="rounded-lg border border-slate-200 px-3 py-2" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">结束时间</span>
        <input v-model="endLocal" type="datetime-local" class="rounded-lg border border-slate-200 px-3 py-2" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">操作人 ID</span>
        <input v-model.number="adminUserId" type="number" min="0" class="rounded-lg border border-slate-200 px-3 py-2" placeholder="0=不限" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">结果</span>
        <select v-model="success" class="rounded-lg border border-slate-200 px-3 py-2">
          <option value="">全部</option>
          <option value="1">成功</option>
          <option value="0">失败</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">方法</span>
        <select v-model="method" class="rounded-lg border border-slate-200 px-3 py-2">
          <option value="">全部</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="PATCH">PATCH</option>
          <option value="DELETE">DELETE</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-sm md:col-span-2">
        <span class="text-slate-700">路径关键字</span>
        <input v-model.trim="pathKeyword" type="text" class="rounded-lg border border-slate-200 px-3 py-2" placeholder="/v1/admin/channels" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-slate-700">权限点</span>
        <input v-model.trim="permKey" type="text" class="rounded-lg border border-slate-200 px-3 py-2" placeholder="admin.channels.write" />
      </label>
      <div class="flex items-end gap-2">
        <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm" @click="load">查询</button>
        <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm" @click="resetRange">最近24小时</button>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1500px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">时间</th>
              <th class="px-4 py-3">操作人</th>
              <th class="px-4 py-3">IP</th>
              <th class="px-4 py-3">方法</th>
              <th class="px-4 py-3">路径</th>
              <th class="px-4 py-3">权限点</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">耗时</th>
              <th class="px-4 py-3">请求ID</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td class="px-4 py-8 text-center text-slate-500" colspan="9">加载中...</td></tr>
            <tr v-else-if="!rows.length"><td class="px-4 py-8 text-center text-slate-500" colspan="9">暂无日志</td></tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-xs">{{ formatTs(r.created_at) }}</td>
              <td class="px-4 py-3">
                <div>{{ r.admin_username || '—' }}</div>
                <div class="font-mono text-[11px] text-slate-500">#{{ r.admin_user_id }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ r.operator_ip || '—' }}</td>
              <td class="px-4 py-3"><span class="rounded bg-slate-100 px-2 py-0.5 font-mono text-xs">{{ r.method }}</span></td>
              <td class="px-4 py-3">
                <div class="font-mono text-xs">{{ r.path }}</div>
                <div v-if="r.query_string" class="font-mono text-[11px] text-slate-500">?{{ r.query_string }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ r.perm_key || '—' }}</td>
              <td class="px-4 py-3">
                <div :class="r.success ? 'text-emerald-700' : 'text-rose-700'" class="text-xs font-semibold">
                  {{ r.success ? '成功' : '失败' }}
                </div>
                <div class="font-mono text-[11px] text-slate-500">HTTP {{ r.http_status }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ r.duration_ms }}ms</td>
              <td class="px-4 py-3 font-mono text-xs">{{ r.request_id || '—' }}</td>
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
import { onMounted, ref } from 'vue'
import { adminGet } from '../../../lib/adminApi'
import { useUiToast } from '../../../composables/useUiToast'

type Row = {
  id: number
  created_at: number
  request_id: string
  admin_user_id: number
  admin_username: string
  operator_ip: string
  user_agent: string
  method: string
  path: string
  query_string: string
  perm_key: string
  http_status: number
  success: boolean
  duration_ms: number
  error_message: string
}

const toast = useUiToast()
const loading = ref(false)
const rows = ref<Row[]>([])
const total = ref(0)
const limit = ref(20)
const offset = ref(0)

const startLocal = ref('')
const endLocal = ref('')
const adminUserId = ref(0)
const method = ref('')
const pathKeyword = ref('')
const permKey = ref('')
const success = ref('')

function toLocalInput(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function localInputToSec(s: string): number {
  if (!s) return 0
  const t = new Date(s)
  return Number.isNaN(t.getTime()) ? 0 : Math.floor(t.getTime() / 1000)
}

function formatTs(ts: number) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('zh-CN', { hour12: false })
}

function resetRange() {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 3600 * 1000)
  endLocal.value = toLocalInput(end)
  startLocal.value = toLocalInput(start)
}

function buildQuery() {
  const p = new URLSearchParams()
  const ss = localInputToSec(startLocal.value)
  const es = localInputToSec(endLocal.value)
  if (ss > 0) p.set('start_sec', String(ss))
  if (es > 0) p.set('end_sec', String(es))
  if (adminUserId.value > 0) p.set('admin_user_id', String(adminUserId.value))
  if (method.value) p.set('method', method.value)
  if (pathKeyword.value) p.set('path_keyword', pathKeyword.value)
  if (permKey.value) p.set('perm_key', permKey.value)
  if (success.value !== '') p.set('success', success.value)
  p.set('limit', String(limit.value))
  p.set('offset', String(offset.value))
  return p.toString()
}

async function load() {
  loading.value = true
  try {
    const q = buildQuery()
    const r = await adminGet<{ rows: Row[]; total: number }>(`/v1/admin/op_logs?${q}`)
    rows.value = r.rows || []
    total.value = r.total || 0
  } catch (e) {
    toast.error(`加载操作日志失败：${e instanceof Error ? e.message : String(e)}`)
  } finally {
    loading.value = false
  }
}

function prev() {
  offset.value = Math.max(0, offset.value - limit.value)
  void load()
}

function next() {
  offset.value += limit.value
  void load()
}

onMounted(() => {
  resetRange()
  void load()
})
</script>
