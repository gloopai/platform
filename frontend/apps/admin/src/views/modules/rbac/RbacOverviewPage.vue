<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">配置总览</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">按菜单查看对应权限点和接口规则，便于快速核对配置关系。</p>
    </div>

    <div class="grid gap-4 lg:grid-cols-12">
      <div class="lg:col-span-4">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="border-b border-slate-100 px-4 py-3">
            <div class="text-sm font-semibold text-slate-800">从菜单视角</div>
            <p class="mt-0.5 text-xs text-slate-500">与左侧导航一致</p>
          </div>
          <div v-if="loading" class="px-4 py-10 text-center text-sm text-slate-500">加载中…</div>
          <div v-else class="max-h-[min(72vh,560px)] overflow-y-auto">
            <button
              type="button"
              class="flex w-full items-center gap-2 border-b border-slate-50 px-4 py-3 text-left text-sm transition hover:bg-slate-50"
              :class="selectedKey === UNBOUND ? 'bg-indigo-50' : ''"
              @click="selectedKey = UNBOUND"
            >
              <span class="rounded bg-slate-200 px-2 py-0.5 text-[10px] font-semibold text-slate-800">其它</span>
              <span class="font-medium text-slate-900">未绑定具体菜单的权限</span>
            </button>
            <button
              v-for="row in menuPickerRows"
              :key="row.raw.menu_key"
              type="button"
              class="flex w-full items-start gap-2 border-b border-slate-50 py-2.5 pr-3 text-left text-sm transition hover:bg-slate-50"
              :style="{ paddingLeft: `${12 + row.depth * 12}px` }"
              :class="selectedKey === row.raw.menu_key ? 'bg-indigo-50' : ''"
              @click="selectedKey = row.raw.menu_key"
            >
              <span
                class="mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold"
                :class="row.raw.kind === 2 ? 'bg-amber-100 text-amber-900' : 'bg-slate-100 text-slate-700'"
              >
                {{ row.raw.kind === 2 ? '组' : '页' }}
              </span>
              <div class="min-w-0">
                <div class="font-medium text-slate-900">{{ row.raw.label }}</div>
                <div class="truncate font-mono text-[11px] text-slate-500">{{ row.raw.menu_key }}</div>
              </div>
            </button>
          </div>
        </div>
      </div>

      <div class="lg:col-span-8 space-y-4">
        <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
          <div class="text-sm font-semibold text-slate-900">{{ panelTitle }}</div>
          <p class="mt-1 text-xs text-slate-500">{{ panelSubtitle }}</p>

          <div class="mt-4">
            <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">功能权限（权限点）</div>
            <p v-if="!visiblePerms.length" class="mt-2 text-sm text-slate-500">暂无挂在此菜单下的权限点。可在「菜单管理 → 其它功能」新建未绑定菜单的能力。</p>
            <ul v-else class="mt-2 divide-y divide-slate-100 rounded-lg border border-slate-100">
              <li v-for="p in visiblePerms" :key="p.perm_key" class="flex flex-wrap items-baseline justify-between gap-2 px-3 py-2.5 text-sm">
                <div>
                  <div class="font-medium text-slate-900">{{ p.label }}</div>
                  <div class="font-mono text-[11px] text-slate-500">{{ p.perm_key }}</div>
                </div>
                <span
                  class="shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold"
                  :class="p.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-600'"
                >
                  {{ p.status === 1 ? '启用' : '停用' }}
                </span>
              </li>
            </ul>
          </div>

          <div class="mt-6">
            <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">这些权限对应的 HTTP 接口</div>
            <p class="mt-1 text-xs text-slate-500">由「接口规则」维护；下列条目按权限点聚合在此菜单下。</p>
            <p v-if="!visibleRules.length" class="mt-2 text-sm text-slate-500">没有接口映射到当前列表中的权限点。</p>
            <ul v-else class="mt-2 divide-y divide-slate-100 rounded-lg border border-slate-100">
              <li v-for="r in visibleRules" :key="r.id" class="px-3 py-2.5 font-mono text-xs text-slate-800">
                <span class="font-semibold text-indigo-700">{{ r.method }}</span>
                <span class="text-slate-400"> · </span>
                <span>{{ r.path_pattern }}</span>
                <div class="mt-0.5 text-[11px] text-slate-500">需要权限 {{ r.perm_key }}</div>
              </li>
            </ul>
          </div>
        </div>

        <div class="flex flex-wrap gap-2 text-xs">
          <RouterLink to="/rbac/menus" class="rounded-lg bg-slate-900 px-3 py-2 font-semibold text-white">菜单管理</RouterLink>
          <RouterLink
            to="/rbac/menus?tab=other"
            class="rounded-lg border border-slate-200 bg-white px-3 py-2 font-semibold text-slate-700"
          >
            其它功能
          </RouterLink>
          <RouterLink
            to="/rbac/roles"
            class="rounded-lg border border-slate-200 bg-white px-3 py-2 font-semibold text-slate-700"
          >
            去给角色授权
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { adminGet } from '../../../lib/adminApi'

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

type ApiRule = {
  id: number
  method: string
  path_pattern: string
  perm_key: string
  status: number
  remark: string
}

const UNBOUND = '__unbound__'
const selectedKey = ref<string>(UNBOUND)
const loading = ref(true)
const error = ref('')
const menus = ref<AdminMenu[]>([])
const permissions = ref<AdminPermission[]>([])
const rules = ref<ApiRule[]>([])

const menuPickerRows = computed(() => {
  const list = menus.value.filter((m) => (m.placement || 'left').toLowerCase() !== 'avatar')
  const byParent = new Map<number, AdminMenu[]>()
  for (const m of list) {
    const p = m.parent_id ?? 0
    if (!byParent.has(p)) byParent.set(p, [])
    byParent.get(p)!.push(m)
  }
  for (const arr of byParent.values()) {
    arr.sort((a, b) => (a.sort_order !== b.sort_order ? a.sort_order - b.sort_order : a.id - b.id))
  }
  const out: { depth: number; raw: AdminMenu }[] = []
  const walk = (parentId: number, depth: number) => {
    const ch = byParent.get(parentId) || []
    for (const m of ch) {
      out.push({ depth, raw: m })
      walk(m.id, depth + 1)
    }
  }
  walk(0, 0)
  return out
})

const panelTitle = computed(() => {
  if (selectedKey.value === UNBOUND) return '未绑定到左侧某一菜单的能力'
  const m = menus.value.find((x) => x.menu_key === selectedKey.value)
  return m ? `${m.label}` : selectedKey.value
})

const panelSubtitle = computed(() => {
  if (selectedKey.value === UNBOUND) {
    return '常见于跨页面共用的权限点；可在「菜单管理 → 其它功能」维护未绑定侧栏菜单的能力。'
  }
  const m = menus.value.find((x) => x.menu_key === selectedKey.value)
  if (!m) return ''
  return m.kind === 2 ? '这是一个分组，下列为挂在此 menu_key 下的权限（一般填在子页面上）。' : `页面前端路径：${m.path || '（空）'}`
})

const visiblePermKeys = computed(() => {
  const keys = new Set<string>()
  if (selectedKey.value === UNBOUND) {
    for (const p of permissions.value) {
      if (!(p.menu_key || '').trim()) keys.add(p.perm_key)
    }
    return keys
  }
  const collectMenuKeys = (k: string): Set<string> => {
    const s = new Set<string>([k])
    const m = menus.value.find((x) => x.menu_key === k)
    if (m && m.kind === 2) {
      const children = menus.value.filter((c) => c.parent_id === m.id)
      for (const ch of children) {
        for (const sub of collectMenuKeys(ch.menu_key)) s.add(sub)
      }
    }
    return s
  }
  const allowedMenuKeys = collectMenuKeys(selectedKey.value)
  for (const p of permissions.value) {
    const mk = (p.menu_key || '').trim()
    if (mk && allowedMenuKeys.has(mk)) keys.add(p.perm_key)
  }
  return keys
})

const visiblePerms = computed(() => {
  const k = visiblePermKeys.value
  return permissions.value.filter((p) => k.has(p.perm_key)).sort((a, b) => a.perm_key.localeCompare(b.perm_key))
})

const visibleRules = computed(() => {
  const k = visiblePermKeys.value
  return rules.value.filter((r) => r.status === 1 && k.has(r.perm_key)).sort((a, b) => a.path_pattern.localeCompare(b.path_pattern))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [mr, pr, rr] = await Promise.all([
      adminGet<{ menus: AdminMenu[] }>('/v1/admin/rbac/menus'),
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
      adminGet<{ rules: ApiRule[] }>('/v1/admin/rbac/api_rules'),
    ])
    menus.value = mr.menus || []
    permissions.value = pr.permissions || []
    rules.value = rr.rules || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
