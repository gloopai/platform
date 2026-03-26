<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-800">新建权限点</div>
      <p class="mt-1 text-xs text-slate-500">命名建议：<span class="font-mono">admin.&lt;模块&gt;.&lt;动作&gt;</span></p>
      <div class="mt-4 grid gap-3 sm:grid-cols-3">
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          perm_key
          <input v-model.trim="newPermKey" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          名称
          <input v-model.trim="newPermLabel" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          分类
          <input v-model.trim="newPermCategory" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" placeholder="如 merchants" />
        </label>
      </div>
      <button
        type="button"
        class="mt-4 rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
        :disabled="saving"
        @click="createPermission"
      >
        创建
      </button>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 px-5 py-3 text-sm font-semibold text-slate-800">权限点目录（按分类）</div>
      <div v-if="loading" class="px-5 py-12 text-center text-sm text-slate-500">加载中…</div>
      <div v-else class="divide-y divide-slate-100">
        <div v-for="cat in grouped" :key="cat.name" class="px-5 py-4">
          <div class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">{{ cat.name || '未分类' }}</div>
          <div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="p in cat.items"
              :key="p.perm_key"
              class="rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2.5"
            >
              <div class="truncate text-sm font-medium text-slate-900">{{ p.label }}</div>
              <div class="mt-0.5 truncate font-mono text-[11px] text-slate-500">{{ p.perm_key }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { adminGet, adminPost } from '../../../lib/adminApi'

type AdminPermission = { id: number; perm_key: string; label: string; category: string; status: number }

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const permissions = ref<AdminPermission[]>([])

const newPermKey = ref('')
const newPermLabel = ref('')
const newPermCategory = ref('')

const grouped = computed(() => {
  const map = new Map<string, AdminPermission[]>()
  for (const p of permissions.value) {
    const k = (p.category || '').trim() || ''
    if (!map.has(k)) map.set(k, [])
    map.get(k)!.push(p)
  }
  const names = Array.from(map.keys()).sort((a, b) => a.localeCompare(b))
  return names.map((name) => ({
    name,
    items: (map.get(name) || []).sort((a, b) => a.perm_key.localeCompare(b.perm_key)),
  }))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const pr = await adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions')
    permissions.value = pr.permissions || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    permissions.value = []
  } finally {
    loading.value = false
  }
}

async function createPermission() {
  const perm_key = newPermKey.value.trim()
  const label = newPermLabel.value.trim()
  const category = newPermCategory.value.trim()
  if (!perm_key || !label) {
    error.value = '请填写 perm_key 与名称'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/rbac/permissions', { perm_key, label, category, status: 1 })
    newPermKey.value = ''
    newPermLabel.value = ''
    newPermCategory.value = ''
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
