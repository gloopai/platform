<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div class="grid gap-6 lg:grid-cols-12">
      <div class="lg:col-span-4">
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
              <button
                v-for="r in roles"
                :key="r.id"
                type="button"
                class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left text-sm transition hover:bg-slate-50"
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
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-8 space-y-6">
        <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-slate-800">菜单授权</div>
              <p class="mt-0.5 text-xs text-slate-500">控制左侧导航与页面入口</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="reloadRoleMenus"
              >
                重新加载
              </button>
              <button
                type="button"
                class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="saveRoleMenus"
              >
                {{ saving ? '保存中…' : '保存' }}
              </button>
            </div>
          </div>

          <div class="mt-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-600">
            当前：<span class="font-mono text-slate-900">{{ selectedRole?.code || '—' }}</span>
            <span class="mx-1">·</span>
            <span class="font-semibold text-slate-800">{{ selectedRole?.name || '未选择' }}</span>
          </div>

          <div class="mt-4">
            <div v-if="!selectedRoleId" class="rounded-xl border border-slate-200 bg-white px-4 py-10 text-center text-sm text-slate-500">
              请选择左侧角色
            </div>
            <div v-else class="space-y-3">
              <div v-for="g in menuTree" :key="g.id" class="rounded-xl border border-slate-200 p-3">
                <div class="flex items-center justify-between gap-2">
                  <div class="text-sm font-semibold text-slate-900">{{ g.label }}</div>
                  <div class="text-[11px] font-mono text-slate-500">{{ g.menu_key }}</div>
                </div>
                <div class="mt-2 grid gap-2 sm:grid-cols-2">
                  <label
                    v-for="c in g.children"
                    :key="c.id"
                    class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm hover:border-slate-300"
                  >
                    <input v-model="selectedMenuIds" :value="c.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                    <span class="min-w-0 truncate text-slate-800">{{ c.label }}</span>
                    <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ c.path }}</span>
                  </label>
                </div>
              </div>

              <div v-if="rootLeaves.length" class="rounded-xl border border-slate-200 p-3">
                <div class="text-sm font-semibold text-slate-900">顶层页面</div>
                <div class="mt-2 grid gap-2 sm:grid-cols-2">
                  <label
                    v-for="c in rootLeaves"
                    :key="c.id"
                    class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm hover:border-slate-300"
                  >
                    <input v-model="selectedMenuIds" :value="c.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                    <span class="min-w-0 truncate text-slate-800">{{ c.label }}</span>
                    <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ c.path }}</span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-slate-800">权限点授权</div>
              <p class="mt-0.5 text-xs text-slate-500">控制接口调用（与「接口规则」配合）</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="reloadRolePerms"
              >
                重新加载
              </button>
              <button
                type="button"
                class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="!selectedRoleId || saving"
                @click="saveRolePerms"
              >
                {{ saving ? '保存中…' : '保存' }}
              </button>
            </div>
          </div>

          <div class="mt-4">
            <div v-if="!selectedRoleId" class="rounded-xl border border-slate-200 bg-white px-4 py-10 text-center text-sm text-slate-500">
              请选择左侧角色
            </div>
            <div v-else class="grid max-h-[420px] gap-2 overflow-y-auto sm:grid-cols-2">
              <label
                v-for="p in permissions"
                :key="p.perm_key"
                class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm hover:border-slate-300"
              >
                <input v-model="selectedPermKeys" :value="p.perm_key" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                <span class="min-w-0 truncate text-slate-800">{{ p.label }}</span>
                <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ p.perm_key }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

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
}
type MenuNode = AdminMenu & { children: AdminMenu[] }
type AdminPermission = { id: number; perm_key: string; label: string; category: string; status: number }

const loading = ref(true)
const saving = ref(false)
const error = ref('')

const roles = ref<AdminRole[]>([])
const menus = ref<AdminMenu[]>([])
const permissions = ref<AdminPermission[]>([])
const selectedRoleId = ref(0)
const selectedMenuIds = ref<number[]>([])
const selectedPermKeys = ref<string[]>([])

const newRoleCode = ref('')
const newRoleName = ref('')

const selectedRole = computed(() => roles.value.find((r) => r.id === selectedRoleId.value) || null)

const menuTree = computed<MenuNode[]>(() => {
  const groups = menus.value.filter((m) => m.kind === 2).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
  const leaves = menus.value.filter((m) => m.kind === 1 && m.parent_id !== 0).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
  const byParent = new Map<number, AdminMenu[]>()
  for (const l of leaves) {
    const arr = byParent.get(l.parent_id) || []
    arr.push(l)
    byParent.set(l.parent_id, arr)
  }
  return groups
    .map((g) => ({ ...g, children: (byParent.get(g.id) || []).slice() }))
    .filter((g) => g.children.length)
})

const rootLeaves = computed(() =>
  menus.value.filter((m) => m.kind === 1 && m.parent_id === 0).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id),
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const rr = await adminGet<{ roles: AdminRole[] }>('/v1/admin/rbac/roles')
    roles.value = rr.roles || []
    const mr = await adminGet<{ menus: AdminMenu[] }>('/v1/admin/rbac/menus')
    menus.value = mr.menus || []
    const pr = await adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions')
    permissions.value = pr.permissions || []
    if (!selectedRoleId.value && roles.value.length) selectedRoleId.value = roles.value[0].id
    if (selectedRoleId.value) {
      await reloadRoleMenus()
      await reloadRolePerms()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
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
  await reloadRoleMenus()
  await reloadRolePerms()
}

async function reloadRoleMenus() {
  if (!selectedRoleId.value) return
  try {
    const r = await adminGet<{ menu_ids: number[] }>(`/v1/admin/rbac/roles/${selectedRoleId.value}/menus`)
    selectedMenuIds.value = (r.menu_ids || []).slice()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    selectedMenuIds.value = []
  }
}

async function saveRoleMenus() {
  if (!selectedRoleId.value) return
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/roles/${selectedRoleId.value}/menus`, { menu_ids: selectedMenuIds.value })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function reloadRolePerms() {
  if (!selectedRoleId.value) return
  try {
    const r = await adminGet<{ perm_keys: string[] }>(`/v1/admin/rbac/roles/${selectedRoleId.value}/perm_keys`)
    selectedPermKeys.value = (r.perm_keys || []).slice()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    selectedPermKeys.value = []
  }
}

async function saveRolePerms() {
  if (!selectedRoleId.value) return
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/roles/${selectedRoleId.value}/perm_keys`, { perm_keys: selectedPermKeys.value })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
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
    await adminPost('/v1/admin/rbac/roles', { code, name, status: 1 })
    newRoleCode.value = ''
    newRoleName.value = ''
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
