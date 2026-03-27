<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">角色与授权</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">选择角色后，按菜单和其它功能分组配置可见菜单与权限点。</p>
    </div>

    <div class="grid gap-6 lg:grid-cols-12">
      <div class="lg:col-span-3">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="flex items-center justify-between gap-2 border-b border-slate-100 bg-slate-50/90 px-4 py-3">
            <div class="text-sm font-semibold text-slate-800">角色</div>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving"
              @click="createRole"
            >
              新建
            </button>
          </div>
          <div class="p-4">
            <div class="grid gap-2">
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                角色编码（唯一）
                <input v-model.trim="newRoleCode" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
              </label>
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                角色名称
                <input v-model.trim="newRoleName" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
              </label>
            </div>
          </div>
          <div class="border-t border-slate-100">
            <div v-if="loading" class="px-4 py-10 text-center text-sm text-slate-500">加载中…</div>
            <div v-else class="divide-y divide-slate-100">
              <div
                v-for="r in roles"
                :key="r.id"
                type="button"
                class="flex w-full cursor-pointer items-center justify-between gap-3 px-4 py-3 text-left text-sm transition hover:bg-slate-50"
                :class="selectedRoleId === r.id ? 'bg-indigo-50' : ''"
                @click="selectRole(r.id)"
              >
                <div class="min-w-0">
                  <div class="truncate font-semibold text-slate-900">{{ r.name }}</div>
                  <div class="truncate font-mono text-[11px] text-slate-500">{{ r.code }}</div>
                </div>
                <span
                  class="inline-flex shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold"
                  :class="r.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ r.status === 1 ? '启用' : '停用' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-9 space-y-6">
        <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold text-slate-800">授权操作</div>
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="reloadRoleGrants"
              >
                重新加载
              </button>
              <button
                type="button"
                class="rounded-lg border border-indigo-200 bg-indigo-50 px-3 py-2 text-xs font-semibold text-indigo-900 disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                title="按已勾选的操作权限，自动勾选对应菜单入口（左侧分组/页面与头像菜单）"
                @click="syncMenusFromPerms"
              >
                按权限同步菜单
              </button>
              <button
                type="button"
                class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-semibold text-rose-700 disabled:opacity-40"
                :disabled="!selectedRoleId || saving || isSelectedSuperAdmin"
                :title="isSelectedSuperAdmin ? '超管角色不可删除' : ''"
                @click="deleteRole"
              >
                删除角色
              </button>
              <button
                type="button"
                class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="saveRoleGrants"
              >
                {{ saving ? '保存中…' : '保存' }}
              </button>
            </div>
          </div>

          <div v-if="selectedRole" class="mt-3 space-y-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-3 text-xs text-slate-600">
            <div>
              当前角色：
              <span class="font-semibold text-slate-900">{{ selectedRole.name }}</span>
              <span class="mx-1 text-slate-400">·</span>
              <span class="font-mono text-slate-700">{{ selectedRole.code }}</span>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <label class="grid gap-1 text-xs font-medium text-slate-700">
                角色名称
                <input
                  v-model.trim="editRoleName"
                  type="text"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900"
                  :disabled="saving"
                />
              </label>
              <label class="grid gap-1 text-xs font-medium text-slate-700">
                状态
                <select
                  v-model.number="editRoleStatus"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900"
                  :disabled="saving || isSelectedSuperAdmin"
                  :title="isSelectedSuperAdmin ? '超管角色须保持启用' : ''"
                >
                  <option :value="1">启用</option>
                  <option :value="0">停用</option>
                </select>
              </label>
            </div>
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-800 shadow-sm disabled:opacity-40"
              :disabled="saving || !editRoleName.trim()"
              @click="saveRoleProfile"
            >
              保存角色信息
            </button>
          </div>

          <div class="mt-4">
            <div v-if="!selectedRoleId" class="rounded-xl border border-slate-200 bg-white px-4 py-10 text-center text-sm text-slate-500">
              请选择左侧角色
            </div>
            <div v-else class="max-h-[min(560px,76vh)] space-y-5 overflow-y-auto pr-1">
              <section v-if="leftMenuRows.length" class="rounded-2xl border border-slate-200 bg-white shadow-sm">
                <div class="border-b border-slate-100 bg-slate-50/70 px-4 py-3">
                  <div class="flex items-center justify-between gap-2">
                    <div class="text-sm font-semibold text-slate-900">左侧菜单</div>
                    <span class="rounded-full bg-slate-200 px-2 py-0.5 text-[10px] font-semibold text-slate-700">{{ leftMenuRows.length }} 项</span>
                  </div>
                  <p class="mt-1 text-xs text-slate-500">每行从左到右：先勾菜单可见，再勾该菜单下权限点。</p>
                </div>
                <div class="divide-y divide-slate-100">
                  <div v-for="r in leftMenuRows" :key="r.menu.id" class="px-4 py-3">
                    <div class="flex items-center gap-2">
                      <label class="flex min-w-0 flex-1 cursor-pointer items-center gap-2" :style="{ paddingLeft: `${r.depth * 16}px` }">
                        <input v-model="selectedMenuIds" :value="r.menu.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                        <span class="truncate text-sm font-semibold text-slate-900">{{ r.menu.label }}</span>
                        <span class="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-600">
                          {{ r.menu.kind === 2 ? `分组 L${r.depth + 1}` : `页面 L${r.depth + 1}` }}
                        </span>
                        <span
                          v-if="r.perms.length"
                          class="ml-1 shrink-0 rounded-full bg-indigo-100 px-2 py-0.5 text-[10px] font-semibold text-indigo-900"
                        >
                          {{ r.perms.length }} 权限
                        </span>
                      </label>
                      <span class="shrink-0 font-mono text-[10px] text-slate-400">{{ r.menu.menu_key }}</span>
                    </div>
                    <div v-if="r.pathText" class="mt-1 text-[11px] text-slate-500" :style="{ paddingLeft: `${r.depth * 16 + 24}px` }">
                      {{ r.pathText }}
                    </div>

                    <div
                      v-if="r.perms.length"
                      class="mt-2 rounded-xl border border-slate-200 bg-slate-50 p-3"
                      :class="r.depth > 0 ? 'border-l-2 border-l-slate-300' : ''"
                      :style="{ marginLeft: `${r.depth * 16}px` }"
                    >
                      <div class="grid gap-2 sm:grid-cols-2">
                        <label
                          v-for="p in r.perms"
                          :key="p.perm_key"
                          class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs hover:border-slate-300"
                        >
                          <input v-model="selectedPermKeys" :value="p.perm_key" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                          <span class="min-w-0 truncate text-slate-800">{{ p.label }}</span>
                          <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ p.perm_key }}</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <section v-if="avatarMenuRows.length" class="rounded-2xl border border-slate-200 bg-white shadow-sm">
                <div class="border-b border-slate-100 bg-slate-50/70 px-4 py-3">
                  <div class="flex items-center justify-between gap-2">
                    <div class="text-sm font-semibold text-slate-900">头像菜单</div>
                    <span class="rounded-full bg-slate-200 px-2 py-0.5 text-[10px] font-semibold text-slate-700">{{ avatarMenuRows.length }} 项</span>
                  </div>
                  <p class="mt-1 text-xs text-slate-500">头像下拉入口及其权限点。</p>
                </div>
                <div class="divide-y divide-slate-100">
                  <div v-for="r in avatarMenuRows" :key="r.menu.id" class="px-4 py-3">
                    <div class="flex items-center gap-2">
                      <label class="flex min-w-0 flex-1 cursor-pointer items-center gap-2">
                        <input v-model="selectedMenuIds" :value="r.menu.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                        <span class="truncate text-sm font-semibold text-slate-900">{{ r.menu.label }}</span>
                        <span
                          v-if="r.perms.length"
                          class="ml-1 shrink-0 rounded-full bg-indigo-100 px-2 py-0.5 text-[10px] font-semibold text-indigo-900"
                        >
                          {{ r.perms.length }} 权限
                        </span>
                      </label>
                      <span class="shrink-0 font-mono text-[10px] text-slate-400">{{ r.menu.menu_key }}</span>
                    </div>

                    <div v-if="r.perms.length" class="mt-2 rounded-xl border border-slate-200 bg-slate-50 p-3">
                      <div class="grid gap-2 sm:grid-cols-2">
                        <label
                          v-for="p in r.perms"
                          :key="p.perm_key"
                          class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs hover:border-slate-300"
                        >
                          <input v-model="selectedPermKeys" :value="p.perm_key" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                          <span class="min-w-0 truncate text-slate-800">{{ p.label }}</span>
                          <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ p.perm_key }}</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <section v-if="otherPerms.length" class="rounded-2xl border border-slate-200 bg-white shadow-sm">
                <div class="border-b border-slate-100 bg-slate-50/70 px-4 py-3">
                  <div class="flex items-center justify-between gap-2">
                    <div class="text-sm font-semibold text-slate-900">其它功能</div>
                    <span class="rounded-full bg-slate-200 px-2 py-0.5 text-[10px] font-semibold text-slate-700">{{ otherPerms.length }} 项</span>
                  </div>
                  <p class="mt-1 text-xs text-slate-500">不挂在任何菜单下的能力，直接按权限点授权。</p>
                </div>
                <div class="p-4">
                  <div class="grid gap-2 sm:grid-cols-2">
                    <label
                      v-for="p in otherPerms"
                      :key="p.perm_key"
                      class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs hover:border-slate-300"
                    >
                      <input v-model="selectedPermKeys" :value="p.perm_key" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                      <span class="min-w-0 truncate text-slate-800">{{ p.label }}</span>
                      <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ p.perm_key }}</span>
                    </label>
                  </div>
                </div>
              </section>

              <p v-if="!leftMenuRows.length && !avatarMenuRows.length && !otherPerms.length" class="py-6 text-center text-sm text-slate-500">
                暂无权限点数据
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import { adminDelete, adminGet, adminPost, adminPut } from '../../../lib/adminApi'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'

type AdminRole = { id: number; code: string; name: string; status: number }
type AdminMenu = {
  id: number
  parent_id: number
  menu_key: string
  label: string
  icon: string
  kind: number
  path: string
  sort_order: number
  placement?: string
}
type AdminPermission = {
  id: number
  perm_key: string
  label: string
  category: string
  menu_key: string
  status: number
}

function permBucket(p: AdminPermission, menus: AdminMenu[]): 'left' | 'avatar' | 'other' {
  const mk = (p.menu_key || '').trim()
  if (!mk) return 'other'
  const menu = menus.find((m) => m.menu_key === mk)
  if (!menu) return 'other'
  if ((menu.placement || 'left').toLowerCase() === 'avatar') return 'avatar'
  return 'left'
}

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const dialog = useUiDialog()
const toast = useUiToast()

const roles = ref<AdminRole[]>([])
const menus = ref<AdminMenu[]>([])
const permissions = ref<AdminPermission[]>([])
const selectedRoleId = ref(0)
const selectedMenuIds = ref<number[]>([])
const selectedPermKeys = ref<string[]>([])

const newRoleCode = ref('')
const newRoleName = ref('')

const editRoleName = ref('')
const editRoleStatus = ref(1)

const selectedRole = computed(() => roles.value.find((r) => r.id === selectedRoleId.value) || null)
const isSelectedSuperAdmin = computed(() => (selectedRole.value?.code || '').trim().toLowerCase() === 'super_admin')

watch(
  selectedRole,
  (r) => {
    if (!r) {
      editRoleName.value = ''
      editRoleStatus.value = 1
      return
    }
    editRoleName.value = r.name
    editRoleStatus.value = r.status
  },
  { immediate: true },
)

type MenuGrantRow = { menu: AdminMenu; depth: number; perms: AdminPermission[]; pathText: string }

const permsByMenuKey = computed(() => {
  const map = new Map<string, AdminPermission[]>()
  for (const p of permissions.value) {
    const mk = (p.menu_key || '').trim()
    if (!mk) continue
    if (!map.has(mk)) map.set(mk, [])
    map.get(mk)!.push(p)
  }
  for (const arr of map.values()) arr.sort((a, b) => a.perm_key.localeCompare(b.perm_key))
  return map
})

const leftMenuRows = computed<MenuGrantRow[]>(() => {
  const navMenus = menus.value
    .filter((m) => (m.placement || 'left').toLowerCase() !== 'avatar')
    .sort((a, b) =>
      a.parent_id === b.parent_id ? a.sort_order - b.sort_order || a.id - b.id : a.parent_id - b.parent_id,
    )
  const byParent = new Map<number, AdminMenu[]>()
  for (const m of navMenus) {
    const p = m.parent_id ?? 0
    if (!byParent.has(p)) byParent.set(p, [])
    byParent.get(p)!.push(m)
  }
  for (const arr of byParent.values()) {
    arr.sort((a, b) => (a.sort_order !== b.sort_order ? a.sort_order - b.sort_order : a.id - b.id))
  }
  const out: MenuGrantRow[] = []
  const byID = new Map<number, AdminMenu>()
  for (const m of navMenus) byID.set(m.id, m)
  const buildPathText = (m: AdminMenu): string => {
    if (m.parent_id === 0) return ''
    const segs: string[] = [m.label]
    let cur = m
    while (cur.parent_id > 0) {
      const p = byID.get(cur.parent_id)
      if (!p) break
      segs.unshift(p.label)
      cur = p
      if (segs.length > 8) break
    }
    return segs.join(' / ')
  }
  const walk = (parentId: number, depth: number) => {
    const ch = byParent.get(parentId) || []
    for (const m of ch) {
      out.push({ menu: m, depth, perms: permsByMenuKey.value.get(m.menu_key) || [], pathText: buildPathText(m) })
      walk(m.id, depth + 1)
    }
  }
  walk(0, 0)
  return out
})

const avatarMenuRows = computed<MenuGrantRow[]>(() =>
  menus.value
    .filter((m) => (m.placement || 'left').toLowerCase() === 'avatar')
    .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
    .map((m) => ({ menu: m, depth: 0, perms: permsByMenuKey.value.get(m.menu_key) || [], pathText: '' })),
)

const otherPerms = computed(() =>
  permissions.value
    .filter((p) => permBucket(p, menus.value) === 'other')
    .sort((a, b) => a.perm_key.localeCompare(b.perm_key)),
)

async function load(preferredRoleId?: number) {
  loading.value = true
  error.value = ''
  try {
    const rr = await adminGet<{ roles: AdminRole[] }>('/v1/admin/rbac/roles')
    roles.value = rr.roles || []
    const mr = await adminGet<{ menus: AdminMenu[] }>('/v1/admin/rbac/menus')
    menus.value = mr.menus || []
    const pr = await adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions')
    permissions.value = pr.permissions || []
    const prefer = preferredRoleId != null && preferredRoleId > 0 ? preferredRoleId : 0
    if (prefer && roles.value.some((r) => r.id === prefer)) {
      selectedRoleId.value = prefer
    } else if (!roles.value.some((r) => r.id === selectedRoleId.value)) {
      selectedRoleId.value = roles.value[0]?.id || 0
    }
    if (selectedRoleId.value) {
      await reloadRoleMenus()
      await reloadRolePerms()
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载角色与权限数据失败：${msg}`)
    roles.value = []
    menus.value = []
    permissions.value = []
    selectedRoleId.value = 0
    selectedMenuIds.value = []
    selectedPermKeys.value = []
  } finally {
    loading.value = false
  }
}

async function selectRole(id: number) {
  selectedRoleId.value = id
  await reloadRoleGrants()
}

async function reloadRoleGrants() {
  await reloadRoleMenus()
  await reloadRolePerms()
}

async function reloadRoleMenus() {
  if (!selectedRoleId.value) return
  try {
    const r = await adminGet<{ menu_ids: number[] }>(`/v1/admin/rbac/roles/${selectedRoleId.value}/menus`)
    selectedMenuIds.value = (r.menu_ids || []).slice()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载角色菜单授权失败：${msg}`)
    selectedMenuIds.value = []
  }
}

async function reloadRolePerms() {
  if (!selectedRoleId.value) return
  try {
    const r = await adminGet<{ perm_keys: string[] }>(`/v1/admin/rbac/roles/${selectedRoleId.value}/perm_keys`)
    selectedPermKeys.value = (r.perm_keys || []).slice()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载角色权限点失败：${msg}`)
    selectedPermKeys.value = []
  }
}

async function saveRoleGrants() {
  if (!selectedRoleId.value) return
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/roles/${selectedRoleId.value}/menus`, { menu_ids: selectedMenuIds.value })
    await adminPut(`/v1/admin/rbac/roles/${selectedRoleId.value}/perm_keys`, { perm_keys: selectedPermKeys.value })
    toast.success('角色授权已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存角色授权失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function saveRoleProfile() {
  if (!selectedRoleId.value || !selectedRole.value) return
  const name = editRoleName.value.trim()
  if (!name) {
    toast.error('请填写角色名称')
    return
  }
  let st = editRoleStatus.value
  if (isSelectedSuperAdmin.value) {
    st = 1
  }
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/roles/${selectedRoleId.value}`, { name, status: st })
    await load(selectedRoleId.value)
    toast.success('角色信息已更新')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`更新角色信息失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function syncMenusFromPerms() {
  if (!selectedRoleId.value) return
  const ids = new Set(selectedMenuIds.value)
  for (const pk of selectedPermKeys.value) {
    const p = permissions.value.find((x) => x.perm_key === pk)
    const mk = (p?.menu_key || '').trim()
    if (!mk) continue
    const leaf = menus.value.find((m) => m.menu_key === mk && m.kind === 1)
    if (!leaf) continue
    ids.add(leaf.id)
    if (leaf.parent_id) {
      const parent = menus.value.find((m) => m.id === leaf.parent_id && m.kind === 2)
      if (parent) ids.add(parent.id)
    }
  }
  selectedMenuIds.value = Array.from(ids)
}

async function createRole() {
  const code = newRoleCode.value.trim()
  const name = newRoleName.value.trim()
  if (!code || !name) {
    error.value = '请填写角色编码与名称'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const resp = await adminPost<{ role: AdminRole }>('/v1/admin/rbac/roles', { code, name, status: 1 })
    const newId = resp.role?.id || 0
    newRoleCode.value = ''
    newRoleName.value = ''
    await load(newId > 0 ? newId : undefined)
    toast.success('角色已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`创建角色失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function deleteRole() {
  if (!selectedRole.value) return
  if ((selectedRole.value.code || '').trim().toLowerCase() === 'super_admin') {
    error.value = '超管角色不可删除'
    return
  }
  const ok = await dialog.confirm(`确认删除角色「${selectedRole.value.name}」？此操作不可恢复。`, '删除角色')
  if (!ok) return
  saving.value = true
  error.value = ''
  try {
    await adminDelete(`/v1/admin/rbac/roles/${selectedRole.value.id}`)
    selectedRoleId.value = 0
    selectedMenuIds.value = []
    selectedPermKeys.value = []
    await load()
    toast.success('角色已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`删除角色失败：${msg}`)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
