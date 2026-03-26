<template>
  <div class="space-y-4 rounded-2xl border border-indigo-100 bg-indigo-50/30 p-5 shadow-sm">
    <div>
      <div class="text-sm font-semibold text-slate-900">功能权限与接口</div>
      <p class="mt-1 text-xs text-slate-600">选中左侧菜单后，这里只展示列表；编辑在表格行内进行。</p>
    </div>

    <div class="space-y-4">
      <div class="rounded-xl border border-slate-200 bg-white">
        <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div class="text-xs font-semibold text-slate-800">权限点列表</div>
          <button
            type="button"
            class="rounded border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 disabled:opacity-40"
            :disabled="saving || !scopeMenuKeys.length"
            @click="startPermCreate"
          >
            ＋ 新建
          </button>
        </div>
        <div class="max-h-[260px] overflow-y-auto">
          <table class="w-full min-w-[620px] text-left text-xs">
            <caption v-if="permCreateOpen" class="border-b border-slate-100 bg-amber-50 px-3 py-2 text-left text-[11px] text-amber-900">
              新建权限点：在下方第一行填写并保存
            </caption>
            <thead class="sticky top-0 border-b border-slate-100 bg-slate-50 text-slate-500">
              <tr>
                <th class="px-3 py-2">名称</th>
                <th class="px-3 py-2">perm_key</th>
                <th class="px-3 py-2">所属菜单</th>
                <th class="px-3 py-2">状态</th>
                <th class="w-36 px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="permCreateOpen" class="bg-amber-50/40">
                <td class="px-3 py-2">
                  <input v-model.trim="permDraft.label" class="w-full rounded border border-slate-200 px-2 py-1 text-xs" />
                </td>
                <td class="px-3 py-2">
                  <input v-model.trim="permDraft.perm_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-[11px]" />
                </td>
                <td class="px-3 py-2">
                  <select v-model="permDraft.menu_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-[11px]">
                    <option v-for="k in scopeMenuKeys" :key="k" :value="k">{{ k }}</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <select v-model.number="permDraft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                    <option :value="1">启用</option>
                    <option :value="0">停用</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <div class="flex justify-end gap-2">
                    <button type="button" class="rounded bg-slate-900 px-2 py-1 font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitPermCreate">
                      保存
                    </button>
                    <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700" :disabled="saving" @click="cancelPermCreate">
                      取消
                    </button>
                  </div>
                </td>
              </tr>

              <tr
                v-for="p in scopedPerms"
                :key="p.id"
                :class="selectedPerm?.id === p.id ? 'bg-indigo-50/40' : ''"
              >
                <td class="px-3 py-2">
                  <template v-if="permEditId === p.id">
                    <input v-model.trim="permDraft.label" class="w-full rounded border border-slate-200 px-2 py-1 text-xs" />
                  </template>
                  <button
                    v-else
                    type="button"
                    class="max-w-[220px] truncate font-medium text-slate-800 hover:underline"
                    :title="p.label"
                    @click="selectPerm(p)"
                  >
                    {{ p.label }}
                  </button>
                </td>
                <td class="px-3 py-2 font-mono text-[11px] text-slate-600">
                  <template v-if="permEditId === p.id">
                    <input disabled :value="p.perm_key" class="w-full rounded border border-slate-100 bg-slate-50 px-2 py-1 font-mono text-[11px]" />
                  </template>
                  <template v-else>
                    <div class="max-w-[180px] truncate" :title="p.perm_key">{{ p.perm_key }}</div>
                  </template>
                </td>
                <td class="px-3 py-2 font-mono text-[11px] text-slate-500">
                  <template v-if="permEditId === p.id">
                    <select v-model="permDraft.menu_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-[11px]">
                      <option v-for="k in scopeMenuKeys" :key="k" :value="k">{{ k }}</option>
                    </select>
                  </template>
                  <template v-else>
                    <div class="max-w-[180px] truncate" :title="p.menu_key || '-'">{{ p.menu_key || '-' }}</div>
                  </template>
                </td>
                <td class="px-3 py-2">
                  <template v-if="permEditId === p.id">
                    <select v-model.number="permDraft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                      <option :value="1">启用</option>
                      <option :value="0">停用</option>
                    </select>
                  </template>
                  <template v-else>{{ p.status === 1 ? '启用' : '停用' }}</template>
                </td>
                <td class="px-3 py-2">
                  <div class="flex justify-end gap-2">
                    <template v-if="permEditId === p.id">
                      <button type="button" class="rounded bg-slate-900 px-2 py-1 font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitPermEdit(p)">
                        保存
                      </button>
                      <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="cancelPermEdit">
                        取消
                      </button>
                    </template>
                    <template v-else>
                      <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700" @click="startPermEdit(p)">编辑</button>
                      <button type="button" class="rounded border border-rose-200 bg-rose-50 px-2 py-1 font-semibold text-rose-700" @click="removePerm(p)">删除</button>
                    </template>
                  </div>
                </td>
              </tr>
              <tr v-if="!scopedPerms.length">
                <td class="px-3 py-6 text-center text-slate-500" colspan="5">该菜单范围暂无权限点</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="rounded-xl border border-slate-200 bg-white">
        <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div>
            <div class="text-xs font-semibold text-slate-800">接口规则列表</div>
            <p class="mt-0.5 text-[11px] text-slate-500">当前权限点：{{ selectedPerm?.perm_key || '未选择' }}</p>
          </div>
          <button
            type="button"
            class="rounded border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 disabled:opacity-40"
            :disabled="saving || !selectedPerm"
            @click="startRuleCreate"
          >
            ＋ 新建规则
          </button>
        </div>
        <div class="max-h-[260px] overflow-y-auto">
          <table class="w-full min-w-[620px] text-left text-xs">
            <caption v-if="ruleCreateOpen" class="border-b border-slate-100 bg-amber-50 px-3 py-2 text-left text-[11px] text-amber-900">
              新建规则：在下方第一行填写并保存（绑定到当前 perm_key）
            </caption>
            <thead class="sticky top-0 border-b border-slate-100 bg-slate-50 text-slate-500">
              <tr>
                <th class="px-3 py-2">Method</th>
                <th class="px-3 py-2">Path</th>
                <th class="px-3 py-2">状态</th>
                <th class="w-36 px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="ruleCreateOpen" class="bg-amber-50/40">
                <td class="px-3 py-2">
                  <select v-model="ruleDraft.method" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                    <option v-for="m in httpMethods" :key="m" :value="m">{{ m }}</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <input v-model.trim="ruleDraft.path_pattern" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs" />
                </td>
                <td class="px-3 py-2">
                  <select v-model.number="ruleDraft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                    <option :value="1">启用</option>
                    <option :value="0">停用</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <div class="flex justify-end gap-2">
                    <button type="button" class="rounded bg-slate-900 px-2 py-1 font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitRuleCreate">
                      保存
                    </button>
                    <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="cancelRuleCreate">
                      取消
                    </button>
                  </div>
                </td>
              </tr>

              <tr v-for="r in rulesForPerm" :key="r.id">
                <td class="px-3 py-2 font-mono">
                  <template v-if="ruleEditId === r.id">
                    <select v-model="ruleDraft.method" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                      <option v-for="m in httpMethods" :key="m" :value="m">{{ m }}</option>
                    </select>
                  </template>
                  <template v-else>
                    <div class="max-w-[90px] truncate" :title="r.method">{{ r.method }}</div>
                  </template>
                </td>
                <td class="px-3 py-2 font-mono">
                  <template v-if="ruleEditId === r.id">
                    <input v-model.trim="ruleDraft.path_pattern" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs" />
                  </template>
                  <template v-else>
                    <div class="max-w-[320px] truncate" :title="r.path_pattern">{{ r.path_pattern }}</div>
                  </template>
                </td>
                <td class="px-3 py-2">
                  <template v-if="ruleEditId === r.id">
                    <select v-model.number="ruleDraft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                      <option :value="1">启用</option>
                      <option :value="0">停用</option>
                    </select>
                  </template>
                  <template v-else>{{ r.status === 1 ? '启用' : '停用' }}</template>
                </td>
                <td class="px-3 py-2">
                  <div class="flex justify-end gap-2">
                    <template v-if="ruleEditId === r.id">
                      <button type="button" class="rounded bg-slate-900 px-2 py-1 font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitRuleEdit(r)">
                        保存
                      </button>
                      <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="cancelRuleEdit">
                        取消
                      </button>
                    </template>
                    <template v-else>
                      <button type="button" class="rounded border border-slate-200 px-2 py-1 font-semibold text-slate-700" :disabled="!selectedPerm" @click="startRuleEdit(r)">
                        编辑
                      </button>
                      <button type="button" class="rounded border border-rose-200 bg-rose-50 px-2 py-1 font-semibold text-rose-700" @click="removeRule(r.id)">删除</button>
                    </template>
                  </div>
                </td>
              </tr>
              <tr v-if="!rulesForPerm.length">
                <td class="px-3 py-6 text-center text-slate-500" colspan="4">暂无规则</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'

import { useUiDialog, useUiToast } from '../../../../composables/ui'
import { adminDelete, adminPost, adminPut } from '../../../../lib/adminApi'
import type { AdminPermission, ApiRule } from './types'

const props = defineProps<{
  scopeMenuKeys: string[]
  permissions: AdminPermission[]
  apiRules: ApiRule[]
}>()

const emit = defineEmits<{
  refresh: []
}>()

const saving = ref(false)
const error = ref('')
const dialog = useUiDialog()
const toast = useUiToast()
const selectedPerm = ref<AdminPermission | null>(null)
const permCreateOpen = ref(false)
const permEditId = ref(0)
const ruleCreateOpen = ref(false)
const ruleEditId = ref(0)

const permDraft = reactive({
  perm_key: '',
  label: '',
  category: '',
  menu_key: '',
  status: 1,
})

const ruleDraft = reactive({
  method: 'GET',
  path_pattern: '',
  status: 1,
})

const httpMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

const scopeSet = computed(() => new Set(props.scopeMenuKeys.map((k) => k.trim()).filter(Boolean)))
const scopedPerms = computed(() =>
  props.permissions
    .filter((p) => scopeSet.value.has((p.menu_key || '').trim()))
    .sort((a, b) => a.perm_key.localeCompare(b.perm_key)),
)
const rulesForPerm = computed(() => {
  const pk = selectedPerm.value?.perm_key
  if (!pk) return []
  return props.apiRules.filter((r) => r.perm_key === pk)
})

const scopeMenuKeys = computed(() => props.scopeMenuKeys)

watch(
  () => props.scopeMenuKeys.join('|'),
  () => {
    selectedPerm.value = null
    permCreateOpen.value = false
    permEditId.value = 0
    ruleCreateOpen.value = false
    ruleEditId.value = 0
  },
)

watch(
  () => props.permissions,
  (list) => {
    const cur = selectedPerm.value
    if (!cur) return
    const next = list.find((x) => x.id === cur.id)
    selectedPerm.value = next ?? null
  },
  { deep: true },
)

function resetPermDraft() {
  permDraft.perm_key = ''
  permDraft.label = ''
  permDraft.category = ''
  permDraft.menu_key = props.scopeMenuKeys[0] || ''
  permDraft.status = 1
}

function resetRuleDraft() {
  ruleDraft.method = 'GET'
  ruleDraft.path_pattern = ''
  ruleDraft.status = 1
}

function selectPerm(p: AdminPermission) {
  selectedPerm.value = p
}

function startPermCreate() {
  error.value = ''
  permEditId.value = 0
  ruleCreateOpen.value = false
  ruleEditId.value = 0
  resetPermDraft()
  permCreateOpen.value = true
}

function cancelPermCreate() {
  permCreateOpen.value = false
}

async function submitPermCreate() {
  const permKey = permDraft.perm_key.trim()
  const label = permDraft.label.trim()
  if (!permKey || !label) {
    error.value = 'perm_key 和 名称为必填项'
    return
  }
  error.value = ''
  saving.value = true
  try {
    await adminPost('/v1/admin/rbac/permissions', {
      perm_key: permKey,
      label,
      category: permDraft.category.trim(),
      menu_key: permDraft.menu_key.trim(),
      status: permDraft.status,
    })
    permCreateOpen.value = false
    emit('refresh')
    toast.success('权限点已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`创建权限点失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function startPermEdit(p: AdminPermission) {
  error.value = ''
  selectedPerm.value = p
  permCreateOpen.value = false
  permEditId.value = p.id
  ruleCreateOpen.value = false
  ruleEditId.value = 0
  permDraft.perm_key = p.perm_key
  permDraft.label = p.label
  permDraft.category = p.category || ''
  permDraft.menu_key = (p.menu_key || '').trim() || props.scopeMenuKeys[0] || ''
  permDraft.status = p.status
}

function cancelPermEdit() {
  permEditId.value = 0
  selectedPerm.value = null
  ruleCreateOpen.value = false
  ruleEditId.value = 0
}

async function submitPermEdit(p: AdminPermission) {
  const label = permDraft.label.trim()
  if (!label) {
    error.value = '名称为必填项'
    return
  }
  error.value = ''
  saving.value = true
  try {
    await adminPut(`/v1/admin/rbac/permissions/${p.id}`, {
      label,
      category: permDraft.category.trim(),
      menu_key: permDraft.menu_key.trim(),
      status: permDraft.status,
    })
    permEditId.value = 0
    emit('refresh')
    toast.success('权限点已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存权限点失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function startRuleCreate() {
  if (!selectedPerm.value) return
  error.value = ''
  permCreateOpen.value = false
  resetRuleDraft()
  ruleCreateOpen.value = true
  ruleEditId.value = 0
}

function cancelRuleCreate() {
  ruleCreateOpen.value = false
}

async function submitRuleCreate() {
  if (!selectedPerm.value) return
  error.value = ''
  saving.value = true
  try {
    await adminPost('/v1/admin/rbac/api_rules', {
      method: ruleDraft.method.trim().toUpperCase(),
      path_pattern: ruleDraft.path_pattern.trim(),
      perm_key: selectedPerm.value.perm_key,
      status: ruleDraft.status,
      remark: '',
    })
    ruleCreateOpen.value = false
    emit('refresh')
    toast.success('接口规则已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`创建接口规则失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function startRuleEdit(r: ApiRule) {
  if (!selectedPerm.value) return
  error.value = ''
  permCreateOpen.value = false
  ruleCreateOpen.value = false
  ruleEditId.value = r.id
  ruleDraft.method = r.method
  ruleDraft.path_pattern = r.path_pattern
  ruleDraft.status = r.status
}

function cancelRuleEdit() {
  ruleEditId.value = 0
}

async function submitRuleEdit(r: ApiRule) {
  if (!selectedPerm.value) return
  error.value = ''
  saving.value = true
  try {
    const methodChanged = r.method !== ruleDraft.method.trim().toUpperCase()
    const pathChanged = r.path_pattern !== ruleDraft.path_pattern.trim()
    if (methodChanged || pathChanged) {
      await adminDelete(`/v1/admin/rbac/api_rules/${r.id}`)
    }
    await adminPost('/v1/admin/rbac/api_rules', {
      method: ruleDraft.method.trim().toUpperCase(),
      path_pattern: ruleDraft.path_pattern.trim(),
      perm_key: selectedPerm.value.perm_key,
      status: ruleDraft.status,
      remark: '',
    })
    ruleEditId.value = 0
    emit('refresh')
    toast.success('接口规则已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存接口规则失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function removePerm(p: AdminPermission) {
  const ok = await dialog.confirm('删除该权限点？')
  if (!ok) return
  error.value = ''
  saving.value = true
  try {
    await adminDelete(`/v1/admin/rbac/permissions/${p.id}`)
    if (selectedPerm.value?.id === p.id) selectedPerm.value = null
    emit('refresh')
    toast.success('权限点已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除权限点失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function removeRule(id: number) {
  const ok = await dialog.confirm('删除该规则？')
  if (!ok) return
  error.value = ''
  saving.value = true
  try {
    await adminDelete(`/v1/admin/rbac/api_rules/${id}`)
    emit('refresh')
    toast.success('接口规则已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除接口规则失败：${msg}`)
  } finally {
    saving.value = false
  }
}
</script>
