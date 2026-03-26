<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">接口规则</h1>
          <p class="mt-1 max-w-3xl text-sm text-slate-600">集中维护接口到权限点的映射，与菜单管理数据实时联动。</p>
        </div>
        <button
          type="button"
          class="rounded border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
          :disabled="saving"
          @click="startCreate"
        >
          ＋ 新建规则
        </button>
      </div>
      <div class="mt-3 grid gap-2 sm:grid-cols-3">
        <input v-model.trim="q" placeholder="搜索 method/path_pattern/perm_key" class="rounded border border-slate-200 px-3 py-2 text-sm sm:col-span-2" />
        <select v-model="permFilter" class="rounded border border-slate-200 px-3 py-2 text-sm">
          <option value="">全部功能点</option>
          <option v-for="p in permKeys" :key="p" :value="p">{{ p }}</option>
        </select>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="max-h-[min(70vh,620px)] overflow-auto">
        <table class="w-full min-w-[920px] text-left text-sm">
          <thead class="sticky top-0 border-b border-slate-100 bg-slate-50 text-xs font-semibold text-slate-500">
            <tr>
              <th class="px-3 py-2">Method</th>
              <th class="px-3 py-2">Path Pattern</th>
              <th class="px-3 py-2">perm_key</th>
              <th class="px-3 py-2">备注</th>
              <th class="px-3 py-2">状态</th>
              <th class="w-40 px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="createOpen" class="bg-amber-50/40">
              <td class="px-3 py-2">
                <select v-model="draft.method" class="w-full rounded border border-slate-200 px-2 py-1 text-xs font-semibold">
                  <option v-for="m in methods" :key="m" :value="m">{{ m }}</option>
                </select>
              </td>
              <td class="px-3 py-2"><input v-model.trim="draft.path_pattern" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs" /></td>
              <td class="px-3 py-2">
                <select v-model.trim="draft.perm_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                  <option value="">请选择</option>
                  <option v-for="p in permKeys" :key="p" :value="p">{{ p }}</option>
                </select>
              </td>
              <td class="px-3 py-2"><input v-model.trim="draft.remark" class="w-full rounded border border-slate-200 px-2 py-1" /></td>
              <td class="px-3 py-2">
                <select v-model.number="draft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                  <option :value="1">启用</option>
                  <option :value="0">停用</option>
                </select>
              </td>
              <td class="px-3 py-2">
                <div class="flex justify-end gap-2">
                  <button class="rounded bg-slate-900 px-2 py-1 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitCreate">保存</button>
                  <button class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="createOpen = false">取消</button>
                </div>
              </td>
            </tr>

            <tr v-for="r in filtered" :key="r.id">
              <td class="px-3 py-2 font-semibold">
                <template v-if="editId === r.id">
                  <select v-model="draft.method" class="w-full rounded border border-slate-200 px-2 py-1 text-xs font-semibold">
                    <option v-for="m in methods" :key="m" :value="m">{{ m }}</option>
                  </select>
                </template>
                <template v-else>{{ r.method }}</template>
              </td>
              <td class="px-3 py-2 font-mono text-xs">
                <template v-if="editId === r.id"><input v-model.trim="draft.path_pattern" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs" /></template>
                <template v-else>{{ r.path_pattern }}</template>
              </td>
              <td class="px-3 py-2 font-mono text-xs">
                <template v-if="editId === r.id">
                  <select v-model.trim="draft.perm_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                    <option value="">请选择</option>
                    <option v-for="p in permKeys" :key="p" :value="p">{{ p }}</option>
                  </select>
                </template>
                <template v-else>{{ r.perm_key }}</template>
              </td>
              <td class="px-3 py-2">
                <template v-if="editId === r.id"><input v-model.trim="draft.remark" class="w-full rounded border border-slate-200 px-2 py-1" /></template>
                <template v-else>{{ r.remark || '-' }}</template>
              </td>
              <td class="px-3 py-2">
                <template v-if="editId === r.id">
                  <select v-model.number="draft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                    <option :value="1">启用</option>
                    <option :value="0">停用</option>
                  </select>
                </template>
                <template v-else>{{ r.status === 1 ? '启用' : '停用' }}</template>
              </td>
              <td class="px-3 py-2">
                <div class="flex justify-end gap-2">
                  <template v-if="editId === r.id">
                    <button class="rounded bg-slate-900 px-2 py-1 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitEdit(r.id)">保存</button>
                    <button class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="editId = 0">取消</button>
                  </template>
                  <template v-else>
                    <button class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-700" @click="startEdit(r)">编辑</button>
                    <button class="rounded border border-rose-200 bg-rose-50 px-2 py-1 text-xs font-semibold text-rose-700" @click="removeRule(r.id)">删除</button>
                  </template>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td class="px-3 py-8 text-center text-slate-500" colspan="6">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { adminDelete, adminGet, adminPost, adminPut } from '../../../lib/adminApi'
import type { AdminPermission, ApiRule } from './menu-management/types'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS']

const saving = ref(false)
const error = ref('')
const q = ref('')
const permFilter = ref('')
const rules = ref<ApiRule[]>([])
const permissions = ref<AdminPermission[]>([])
const createOpen = ref(false)
const editId = ref(0)

const draft = reactive({
  method: 'GET',
  path_pattern: '',
  perm_key: '',
  remark: '',
  status: 1,
})

const permKeys = computed(() =>
  [...new Set(permissions.value.map((p) => p.perm_key.trim()).filter(Boolean))].sort((a, b) => a.localeCompare(b)),
)

const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  return rules.value
    .filter((r) => {
      if (permFilter.value && r.perm_key !== permFilter.value) return false
      if (!s) return true
      return [r.method, r.path_pattern, r.perm_key, r.remark].join(' ').toLowerCase().includes(s)
    })
    .sort((a, b) => a.path_pattern.localeCompare(b.path_pattern))
})

function resetDraft() {
  draft.method = 'GET'
  draft.path_pattern = ''
  draft.perm_key = ''
  draft.remark = ''
  draft.status = 1
}

function startCreate() {
  editId.value = 0
  createOpen.value = true
  resetDraft()
}

function startEdit(r: ApiRule) {
  createOpen.value = false
  editId.value = r.id
  draft.method = r.method
  draft.path_pattern = r.path_pattern
  draft.perm_key = r.perm_key
  draft.remark = r.remark || ''
  draft.status = r.status
}

async function load() {
  error.value = ''
  try {
    const [rr, pr] = await Promise.all([
      adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules'),
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
    ])
    rules.value = rr.rules || []
    permissions.value = pr.permissions || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    rules.value = []
    permissions.value = []
  }
}

async function submitCreate() {
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/rbac/api_rules', {
      method: draft.method,
      path_pattern: draft.path_pattern.trim(),
      perm_key: draft.perm_key.trim(),
      remark: draft.remark.trim(),
      status: draft.status,
    })
    createOpen.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function submitEdit(id: number) {
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/api_rules/${id}`, {
      method: draft.method,
      path_pattern: draft.path_pattern.trim(),
      perm_key: draft.perm_key.trim(),
      remark: draft.remark.trim(),
      status: draft.status,
    })
    editId.value = 0
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function removeRule(id: number) {
  if (!confirm('删除该接口规则？')) return
  saving.value = true
  error.value = ''
  try {
    await adminDelete(`/v1/admin/rbac/api_rules/${id}`)
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
