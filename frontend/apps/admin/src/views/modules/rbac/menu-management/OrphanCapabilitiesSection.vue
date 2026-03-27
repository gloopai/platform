<template>
  <div class="space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">说明</div>
      <p class="mt-2 text-sm leading-relaxed text-slate-600">
        「其它功能」列出<strong>未绑定侧栏菜单</strong>的权限点（<span class="font-mono">menu_key</span> 为空），用于跨页面接口、纯后端能力等。可在此维护
        <span class="font-mono">perm_key</span> 与网关接口规则。
      </p>
    </div>

    <div class="grid gap-4 lg:grid-cols-12">
      <div class="lg:col-span-4">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="border-b border-slate-100 px-4 py-3">
            <div class="text-sm font-semibold text-slate-800">能力列表</div>
            <input
              v-model.trim="q"
              type="search"
              placeholder="搜索…"
              class="mt-2 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
            />
            <button type="button" class="mt-2 w-full rounded-lg border border-slate-200 bg-white py-2 text-xs font-semibold text-slate-700" @click="openCreateDialog">
              ＋ 新建
            </button>
          </div>
          <div class="max-h-[min(70vh,560px)] overflow-y-auto">
            <div
              v-for="p in filteredOrphans"
              :key="p.perm_key"
              type="button"
              class="flex w-full cursor-pointer flex-col items-start gap-0.5 border-b border-slate-50 px-4 py-2.5 text-left text-sm transition hover:bg-slate-50"
              :class="selected?.perm_key === p.perm_key ? 'bg-indigo-50' : ''"
              @click="select(p)"
            >
              <span class="font-medium text-slate-900">{{ p.label }}</span>
              <span class="font-mono text-[11px] text-slate-500">{{ p.perm_key }}</span>
          </div>
            <div v-if="!filteredOrphans.length && !loading" class="px-4 py-8 text-center text-sm text-slate-500">无匹配项</div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-8 space-y-4">
        <div v-if="!selected && !creating" class="rounded-2xl border border-dashed border-slate-200 bg-slate-50/80 px-6 py-16 text-center text-sm text-slate-500">
          请选择左侧一项进行编辑
        </div>
        <template v-else>
          <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
            <div class="text-sm font-semibold text-slate-900">编辑能力</div>
            <p class="mt-1 text-xs text-slate-500">不绑定侧栏菜单的 perm_key，供接口规则与角色授权引用。</p>
            <div class="mt-4 grid gap-3 sm:grid-cols-2">
              <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
                perm_key（唯一）
                <input
                  v-model.trim="form.perm_key"
                  type="text"
                  class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
                  :disabled="true"
                />
              </label>
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                名称
                <input v-model.trim="form.label" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
              </label>
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                内部分类（可选）
                <input v-model.trim="form.category" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
              </label>
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                状态
                <select v-model.number="form.status" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
                  <option :value="1">启用</option>
                  <option :value="0">停用</option>
                </select>
              </label>
            </div>
            <div class="mt-4 flex flex-wrap gap-2">
              <button
                type="button"
                class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="saving"
                @click="savePerm"
              >
                {{ saving ? '保存中…' : '保存' }}
              </button>
              <button
                v-if="!creating && selected"
                type="button"
                class="rounded-lg border border-rose-200 bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-800 disabled:opacity-40"
                :disabled="saving"
                @click="removePerm"
              >
                删除
              </button>
            </div>
          </div>
          <div v-if="!creating && selected" class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
            <div class="text-sm font-semibold text-slate-900">HTTP 接口规则</div>
            <p class="mt-0.5 text-xs text-slate-500">网关按方法 + 路径匹配；多条规则可共用同一 perm_key。</p>
            <div class="mt-4 grid gap-2 sm:grid-cols-12">
              <label class="sm:col-span-2 grid gap-1 text-xs font-medium text-slate-600">
                Method
                <input v-model.trim="ruleForm.method" class="rounded-lg border border-slate-200 px-2 py-2 font-mono text-sm uppercase" />
              </label>
              <label class="sm:col-span-6 grid gap-1 text-xs font-medium text-slate-600">
                Path
                <input v-model.trim="ruleForm.path_pattern" class="rounded-lg border border-slate-200 px-2 py-2 font-mono text-sm" />
              </label>
              <label class="sm:col-span-2 grid gap-1 text-xs font-medium text-slate-600">
                状态
                <select v-model.number="ruleForm.status" class="rounded-lg border border-slate-200 px-2 py-2 text-sm">
                  <option :value="1">启用</option>
                  <option :value="0">停用</option>
                </select>
              </label>
              <div class="sm:col-span-2 flex items-end">
                <button
                  type="button"
                  class="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white disabled:opacity-40"
                  :disabled="saving"
                  @click="saveRule"
                >
                  添加/更新
                </button>
              </div>
            </div>
            <div class="mt-4 overflow-x-auto rounded-lg border border-slate-100">
              <table class="w-full min-w-[480px] text-left text-sm">
                <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold text-slate-500">
                  <tr>
                    <th class="px-3 py-2">Method</th>
                    <th class="px-3 py-2">Path</th>
                    <th class="px-3 py-2">状态</th>
                    <th class="w-24 px-3 py-2"></th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="r in rulesForSelected" :key="r.id">
                    <td class="px-3 py-2 font-mono text-xs">{{ r.method }}</td>
                    <td class="px-3 py-2 font-mono text-xs">{{ r.path_pattern }}</td>
                    <td class="px-3 py-2 text-xs">{{ r.status === 1 ? '启用' : '停用' }}</td>
                    <td class="px-3 py-2">
                      <button type="button" class="text-xs font-semibold text-rose-600 hover:underline" @click="deleteRule(r.id)">删除</button>
                    </td>
                  </tr>
                  <tr v-if="!rulesForSelected.length">
                    <td class="px-3 py-6 text-center text-slate-500" colspan="4">暂无规则</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </template>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="showCreateDialog" class="modal modal-open">
        <div class="modal-box w-11/12 max-w-xl rounded-2xl border border-slate-200 bg-white p-5 shadow-2xl">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-slate-900">新建能力</div>
              <div class="mt-1 text-xs text-slate-500">创建未绑定侧栏菜单的权限点（menu_key 为空）。</div>
            </div>
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50"
              :disabled="saving"
              @click="closeCreateDialog"
            >
              关闭
            </button>
          </div>

          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              perm_key（唯一）
              <input
                v-model.trim="form.perm_key"
                type="text"
                class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              名称
              <input v-model.trim="form.label" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              内部分类（可选）
              <input v-model.trim="form.category" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              状态
              <select v-model.number="form.status" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
                <option :value="1">启用</option>
                <option :value="0">停用</option>
              </select>
            </label>
          </div>

          <div class="mt-4 flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              :disabled="saving"
              @click="closeCreateDialog"
            >
              取消
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving || !form.perm_key.trim() || !form.label.trim()"
              @click="savePerm"
            >
              {{ saving ? '创建中…' : '创建能力' }}
            </button>
          </div>
        </div>
        <form method="dialog" class="modal-backdrop">
          <button type="button" @click="closeCreateDialog">close</button>
        </form>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, Teleport } from 'vue'

import { useUiDialog } from '../../../../composables/useUiDialog'
import { useUiToast } from '../../../../composables/useUiToast'
import { adminDelete, adminGet, adminPost, adminPut } from '../../../../lib/adminApi'
import type { AdminPermission, ApiRule } from './types'

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const dialog = useUiDialog()
const toast = useUiToast()
const q = ref('')
const permissions = ref<AdminPermission[]>([])
const apiRules = ref<ApiRule[]>([])
const selected = ref<AdminPermission | null>(null)
const creating = ref(false)
const showCreateDialog = ref(false)

const form = reactive({
  id: 0,
  perm_key: '',
  label: '',
  category: '',
  status: 1,
})

const ruleForm = reactive({
  method: 'GET',
  path_pattern: '',
  status: 1,
})

const orphans = computed(() => permissions.value.filter((p) => !(p.menu_key || '').trim()))

const filteredOrphans = computed(() => {
  const s = q.value.trim().toLowerCase()
  const list = [...orphans.value].sort((a, b) => a.label.localeCompare(b.label))
  if (!s) return list
  return list.filter(
    (p) =>
      p.label.toLowerCase().includes(s) ||
      p.perm_key.toLowerCase().includes(s),
  )
})

const rulesForSelected = computed(() => {
  const pk = selected.value?.perm_key
  if (!pk) return []
  return apiRules.value.filter((r) => r.perm_key === pk)
})

function select(p: AdminPermission) {
  creating.value = false
  selected.value = p
  form.id = p.id
  form.perm_key = p.perm_key
  form.label = p.label
  form.category = p.category || ''
  form.status = p.status
  ruleForm.method = 'GET'
  ruleForm.path_pattern = ''
  ruleForm.status = 1
}

function startCreate() {
  creating.value = true
  selected.value = null
  form.id = 0
  form.perm_key = ''
  form.label = ''
  form.category = ''
  form.status = 1
}

function openCreateDialog() {
  startCreate()
  showCreateDialog.value = true
}

function closeCreateDialog() {
  showCreateDialog.value = false
  creating.value = false
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [pr, rr] = await Promise.all([
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
      adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules'),
    ])
    permissions.value = pr.permissions || []
    apiRules.value = rr.rules || []
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载能力列表失败：${msg}`)
  } finally {
    loading.value = false
  }
}

async function savePerm() {
  const isCreate = creating.value
  saving.value = true
  error.value = ''
  try {
    if (creating.value) {
      await adminPost('/v1/admin/rbac/permissions', {
        perm_key: form.perm_key.trim(),
        label: form.label.trim(),
        category: form.category.trim(),
        menu_key: '',
        status: form.status,
      })
    } else if (selected.value) {
      await adminPut(`/v1/admin/rbac/permissions/${form.id}`, {
        label: form.label.trim(),
        category: form.category.trim(),
        menu_key: '',
        status: form.status,
      })
    }
    await load()
    creating.value = false
    showCreateDialog.value = false
    const p = permissions.value.find((x) => x.perm_key === form.perm_key.trim())
    if (p) select(p)
    else selected.value = null
    toast.success(isCreate ? '能力已创建' : '能力已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存能力失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function removePerm() {
  if (!selected.value) return
  const ok = await dialog.confirm('确定删除？将先解除角色绑定。')
  if (!ok) return
  saving.value = true
  error.value = ''
  try {
    await adminDelete(`/v1/admin/rbac/permissions/${selected.value.id}`)
    selected.value = null
    creating.value = false
    await load()
    toast.success('能力已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除能力失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function saveRule() {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/rbac/api_rules', {
      method: ruleForm.method.trim().toUpperCase(),
      path_pattern: ruleForm.path_pattern.trim(),
      perm_key: selected.value.perm_key,
      status: ruleForm.status,
      remark: '',
    })
    const rr = await adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules')
    apiRules.value = rr.rules || []
    ruleForm.path_pattern = ''
    toast.success('接口规则已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存接口规则失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function deleteRule(id: number) {
  const ok = await dialog.confirm('删除该接口规则？')
  if (!ok) return
  saving.value = true
  error.value = ''
  try {
    await adminDelete(`/v1/admin/rbac/api_rules/${id}`)
    const rr = await adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules')
    apiRules.value = rr.rules || []
    toast.success('接口规则已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除接口规则失败：${msg}`)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
