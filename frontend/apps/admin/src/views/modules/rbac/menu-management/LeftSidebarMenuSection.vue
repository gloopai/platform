<template>
  <div class="space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">说明</div>
      <p class="mt-2 text-sm leading-relaxed text-slate-600">
        维护<strong>左侧导航</strong>的分组与页面。选中左侧某一项后，可在下方配置该菜单范围内的<strong>功能权限</strong>与<strong>HTTP 接口规则</strong>；未挂菜单的能力仍在「其它功能」中维护。
      </p>
    </div>

    <div class="grid gap-6 lg:grid-cols-12">
      <div class="lg:col-span-4">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="flex items-center justify-between gap-2 border-b border-slate-100 bg-slate-50/90 px-4 py-3">
            <div>
              <div class="text-sm font-semibold text-slate-800">侧栏结构</div>
              <p class="mt-0.5 text-xs text-slate-500">仅含「左侧」位置；点击一行可编辑</p>
            </div>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving"
              @click="openCreateDialog"
            >
              新建
            </button>
          </div>
          <div v-if="loading" class="px-4 py-10 text-center text-sm text-slate-500">加载中…</div>
          <div v-else class="max-h-[min(70vh,520px)] divide-y divide-slate-100 overflow-y-auto">
            <div
              v-for="row in treeRows"
              :key="row.id"
              type="button"
              class="flex w-full items-start gap-2 px-4 py-2.5 text-left text-sm transition hover:bg-slate-50 cursor-pointer"
              :class="selectedId === row.id ? 'bg-indigo-50' : ''"
              :style="{ paddingLeft: `${12 + row.depth * 14}px` }"
              @click="selectRow(row.raw)"
            >
              <span
                class="mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold"
                :class="row.raw.kind === 2 ? 'bg-amber-100 text-amber-900' : 'bg-slate-100 text-slate-700'"
              >
                {{ row.raw.kind === 2 ? '分组' : '页面' }}
              </span>
              <div class="min-w-0 flex-1">
                <div class="font-medium text-slate-900">{{ row.raw.label }}</div>
                <div class="truncate font-mono text-[11px] text-slate-500">{{ row.raw.menu_key }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-8 space-y-4">
        <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
          <div class="text-sm font-semibold text-slate-800">编辑菜单项</div>
          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              上级（分组 ID，0 表示顶栏一级）
              <input v-model.number="form.parent_id" type="number" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              唯一键 menu_key
              <input v-model.trim="form.menu_key" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" :disabled="!!form.id" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              显示名称
              <input v-model.trim="form.label" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              图标（预留）
              <input v-model.trim="form.icon" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" placeholder="如 chart" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              类型：1=页面 2=分组
              <select v-model.number="form.kind" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
                <option :value="1">1 · 页面（需填前端路径）</option>
                <option :value="2">2 · 分组（路径留空）</option>
              </select>
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              前端路由 path（仅页面需要，如 /stats）
              <input v-model.trim="form.path" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              排序（数字越小越靠前）
              <input v-model.number="form.sort_order" type="number" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving"
              @click="saveMenu"
            >
              {{ saving ? '保存中…' : form.id ? '保存修改' : '创建' }}
            </button>
            <button
              v-if="form.id"
              type="button"
              class="rounded-lg border border-rose-200 bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-800 disabled:opacity-40"
              :disabled="saving"
              @click="removeMenu"
            >
              删除此项
            </button>
          </div>
        </div>

        <MenuPermRulesPanel
          v-if="selectedId > 0 && scopeMenuKeys.length"
          :scope-menu-keys="scopeMenuKeys"
          :permissions="permissions"
          :api-rules="apiRules"
          @refresh="refreshPermRules"
        />

        <div class="rounded-xl border border-slate-100 bg-slate-50/80 p-4 text-xs text-slate-600">
          <div class="font-semibold text-slate-700">提示</div>
          <ul class="mt-2 list-disc space-y-1 pl-4 leading-relaxed">
            <li>本标签保存的项位置固定为「左侧」；头像下菜单请在「头像下方」标签维护。</li>
            <li>删除前请先删除子菜单；<span class="font-mono">menu_key</span> 创建后勿随意改。</li>
          </ul>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="showCreateDialog" class="modal modal-open">
        <div class="modal-box w-11/12 max-w-2xl rounded-2xl border border-slate-200 bg-white p-5 shadow-2xl">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-slate-900">新建左侧菜单</div>
              <div class="mt-1 text-xs text-slate-500">创建后将自动选中，可继续在右侧配置权限与接口规则。</div>
            </div>
            <button type="button" class="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50" :disabled="saving" @click="closeCreateDialog">关闭</button>
          </div>

          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              上级（分组 ID，0 表示顶栏一级）
              <input v-model.number="createForm.parent_id" type="number" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              唯一键 menu_key
              <input v-model.trim="createForm.menu_key" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              显示名称
              <input v-model.trim="createForm.label" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              图标（预留）
              <input v-model.trim="createForm.icon" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" placeholder="如 chart" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600">
              类型：1=页面 2=分组
              <select v-model.number="createForm.kind" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
                <option :value="1">1 · 页面（需填前端路径）</option>
                <option :value="2">2 · 分组（路径留空）</option>
              </select>
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              前端路由 path（仅页面需要，如 /stats）
              <input v-model.trim="createForm.path" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
            </label>
            <label class="grid gap-1 text-xs font-medium text-slate-600 sm:col-span-2">
              排序（数字越小越靠前）
              <input v-model.number="createForm.sort_order" type="number" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
            </label>
          </div>

          <div class="mt-4 flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
            <button type="button" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700" :disabled="saving" @click="closeCreateDialog">取消</button>
            <button type="button" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving || !createForm.menu_key.trim() || !createForm.label.trim()" @click="createMenu">{{ saving ? '创建中…' : '创建菜单' }}</button>
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
import MenuPermRulesPanel from './MenuPermRulesPanel.vue'
import { collectScopeMenuKeys } from './menuScopeKeys'
import type { AdminPermission, ApiRule, RbacAdminMenu } from './types'

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const dialog = useUiDialog()
const toast = useUiToast()
const menus = ref<RbacAdminMenu[]>([])
const permissions = ref<AdminPermission[]>([])
const apiRules = ref<ApiRule[]>([])
const selectedId = ref(0)
const showCreateDialog = ref(false)

const leftMenus = computed(() =>
  menus.value.filter((m) => {
    const p = (m.placement || 'left').toLowerCase()
    return p !== 'avatar'
  }),
)

const selectedRow = computed(() => menus.value.find((m) => m.id === selectedId.value) ?? null)

const scopeMenuKeys = computed(() => {
  const m = selectedRow.value
  if (!m) return []
  return collectScopeMenuKeys(m, leftMenus.value)
})

const form = reactive({
  id: 0,
  parent_id: 0,
  menu_key: '',
  label: '',
  icon: '',
  kind: 1,
  path: '',
  sort_order: 0,
})

const createForm = reactive({
  parent_id: 0,
  menu_key: '',
  label: '',
  icon: '',
  kind: 1,
  path: '',
  sort_order: 10,
})

const treeRows = computed(() => {
  const list = [...leftMenus.value].sort((a, b) =>
    a.parent_id === b.parent_id ? a.sort_order - b.sort_order || a.id - b.id : a.parent_id - b.parent_id,
  )
  const byParent = new Map<number, RbacAdminMenu[]>()
  for (const m of list) {
    const p = m.parent_id ?? 0
    if (!byParent.has(p)) byParent.set(p, [])
    byParent.get(p)!.push(m)
  }
  for (const arr of byParent.values()) {
    arr.sort((a, b) => (a.sort_order !== b.sort_order ? a.sort_order - b.sort_order : a.id - b.id))
  }
  const out: { id: number; depth: number; raw: RbacAdminMenu }[] = []
  const walk = (parentId: number, depth: number) => {
    const ch = byParent.get(parentId) || []
    for (const m of ch) {
      out.push({ id: m.id, depth, raw: m })
      walk(m.id, depth + 1)
    }
  }
  walk(0, 0)
  return out
})

function selectRow(m: RbacAdminMenu) {
  selectedId.value = m.id
  form.id = m.id
  form.parent_id = m.parent_id ?? 0
  form.menu_key = m.menu_key || ''
  form.label = m.label || ''
  form.icon = m.icon || ''
  form.kind = m.kind ?? 1
  form.path = m.path || ''
  form.sort_order = m.sort_order ?? 0
}

function resetNewForm() {
  selectedId.value = 0
  form.id = 0
  form.parent_id = 0
  form.menu_key = ''
  form.label = ''
  form.icon = ''
  form.kind = 1
  form.path = ''
  form.sort_order = (leftMenus.value.length ? Math.max(...leftMenus.value.map((x) => x.sort_order)) : 0) + 10
}

async function refreshPermRules() {
  try {
    const [pr, rr] = await Promise.all([
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
      adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules'),
    ])
    permissions.value = pr.permissions || []
    apiRules.value = rr.rules || []
  } catch {
    /* ignore */
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [mr, pr, rr] = await Promise.all([
      adminGet<{ menus: RbacAdminMenu[] }>('/v1/admin/rbac/menus'),
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
      adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules'),
    ])
    menus.value = mr.menus || []
    permissions.value = pr.permissions || []
    apiRules.value = rr.rules || []
    if (!form.id && !selectedId.value) resetNewForm()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载左侧菜单失败：${msg}`)
    menus.value = []
    permissions.value = []
    apiRules.value = []
  } finally {
    loading.value = false
  }
}

async function saveMenu() {
  saving.value = true
  error.value = ''
  const keepId = form.id
  const createKey = form.menu_key.trim()
  try {
    const body = {
      parent_id: form.parent_id,
      menu_key: form.menu_key,
      label: form.label,
      icon: form.icon,
      kind: form.kind,
      path: form.path,
      sort_order: form.sort_order,
      placement: 'left',
    }
    let saved: RbacAdminMenu | null = null
    if (form.id) {
      const resp = await adminPut<{ menu: RbacAdminMenu }>(`/v1/admin/rbac/menus/${form.id}`, body)
      saved = resp.menu || null
    } else {
      const resp = await adminPost<{ menu: RbacAdminMenu }>('/v1/admin/rbac/menus', body)
      saved = resp.menu || null
    }
    if (saved) {
      const idx = menus.value.findIndex((x) => x.id === saved!.id)
      if (idx >= 0) menus.value[idx] = saved
      else menus.value.unshift(saved)
    }
    if (keepId) {
      const m = menus.value.find((x) => x.id === keepId)
      if (m) selectRow(m)
    } else if (createKey) {
      const m = menus.value.find((x) => x.menu_key === createKey)
      if (m) selectRow(m)
      else resetNewForm()
    } else {
      resetNewForm()
    }
    toast.success(keepId ? '菜单已保存' : '菜单已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存菜单失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function removeMenu() {
  if (!form.id) return
  const ok = await dialog.confirm('确定删除？须先删掉子菜单。')
  if (!ok) return
  saving.value = true
  error.value = ''
  const removeId = form.id
  try {
    await adminDelete(`/v1/admin/rbac/menus/${removeId}`)
    menus.value = menus.value.filter((x) => x.id !== removeId)
    resetNewForm()
    toast.success('菜单已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除菜单失败：${msg}`)
  } finally {
    saving.value = false
  }
}



function resetCreateForm() {
  createForm.parent_id = 0
  createForm.menu_key = ''
  createForm.label = ''
  createForm.icon = ''
  createForm.kind = 1
  createForm.path = ''
  createForm.sort_order = (leftMenus.value.length ? Math.max(...leftMenus.value.map((x) => x.sort_order)) : 0) + 10
}

function openCreateDialog() {
  resetCreateForm()
  showCreateDialog.value = true
}

function closeCreateDialog() {
  showCreateDialog.value = false
}

async function createMenu() {
  saving.value = true
  error.value = ''
  const createKey = createForm.menu_key.trim()
  try {
    const body = {
      parent_id: createForm.parent_id,
      menu_key: createForm.menu_key,
      label: createForm.label,
      icon: createForm.icon,
      kind: createForm.kind,
      path: createForm.path,
      sort_order: createForm.sort_order,
      placement: 'left',
    }
    const resp = await adminPost<{ menu: RbacAdminMenu }>('/v1/admin/rbac/menus', body)
    const saved = resp.menu || null
    if (saved) {
      const idx = menus.value.findIndex((x) => x.id === saved.id)
      if (idx >= 0) menus.value[idx] = saved
      else menus.value.unshift(saved)
      selectRow(saved)
    } else if (createKey) {
      const m = menus.value.find((x) => x.menu_key === createKey)
      if (m) selectRow(m)
    }
    closeCreateDialog()
    toast.success('菜单已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`创建菜单失败：${msg}`)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
