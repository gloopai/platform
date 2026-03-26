<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">功能点</h1>
          <p class="mt-1 max-w-3xl text-sm text-slate-600">集中维护全部权限点，与菜单管理数据实时联动。</p>
        </div>
        <button
          type="button"
          class="rounded border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
          :disabled="saving"
          @click="startCreate"
        >
          ＋ 新建功能点
        </button>
      </div>
      <div class="mt-3 grid gap-2 sm:grid-cols-3">
        <input v-model.trim="q" placeholder="搜索名称/perm_key/menu_key" class="rounded border border-slate-200 px-3 py-2 text-sm sm:col-span-2" />
        <select v-model="menuFilter" class="rounded border border-slate-200 px-3 py-2 text-sm">
          <option value="">全部菜单</option>
          <option value="__empty__">其它功能（menu_key 为空）</option>
          <option v-for="m in menuKeys" :key="m" :value="m">{{ m }}</option>
        </select>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="max-h-[min(70vh,620px)] overflow-auto">
        <table class="w-full min-w-[920px] text-left text-sm">
          <thead class="sticky top-0 border-b border-slate-100 bg-slate-50 text-xs font-semibold text-slate-500">
            <tr>
              <th class="px-3 py-2">名称</th>
              <th class="px-3 py-2">perm_key</th>
              <th class="px-3 py-2">分类</th>
              <th class="px-3 py-2">menu_key</th>
              <th class="px-3 py-2">状态</th>
              <th class="w-40 px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="createOpen" class="bg-amber-50/40">
              <td class="px-3 py-2"><input v-model.trim="draft.label" class="w-full rounded border border-slate-200 px-2 py-1" /></td>
              <td class="px-3 py-2"><input v-model.trim="draft.perm_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs" /></td>
              <td class="px-3 py-2"><input v-model.trim="draft.category" class="w-full rounded border border-slate-200 px-2 py-1" /></td>
              <td class="px-3 py-2">
                <select v-model="draft.menu_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                  <option value="">(空)</option>
                  <option v-for="m in menuKeys" :key="m" :value="m">{{ m }}</option>
                </select>
              </td>
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

            <tr v-for="p in filtered" :key="p.id">
              <td class="px-3 py-2">
                <template v-if="editId === p.id"><input v-model.trim="draft.label" class="w-full rounded border border-slate-200 px-2 py-1" /></template>
                <template v-else>{{ p.label }}</template>
              </td>
              <td class="px-3 py-2 font-mono text-xs">
                <template v-if="editId === p.id"><input disabled :value="p.perm_key" class="w-full rounded border border-slate-100 bg-slate-50 px-2 py-1 font-mono text-xs" /></template>
                <template v-else>{{ p.perm_key }}</template>
              </td>
              <td class="px-3 py-2">
                <template v-if="editId === p.id"><input v-model.trim="draft.category" class="w-full rounded border border-slate-200 px-2 py-1" /></template>
                <template v-else>{{ p.category || '-' }}</template>
              </td>
              <td class="px-3 py-2 font-mono text-xs">
                <template v-if="editId === p.id">
                  <select v-model="draft.menu_key" class="w-full rounded border border-slate-200 px-2 py-1 font-mono text-xs">
                    <option value="">(空)</option>
                    <option v-for="m in menuKeys" :key="m" :value="m">{{ m }}</option>
                  </select>
                </template>
                <template v-else>{{ p.menu_key || '-' }}</template>
              </td>
              <td class="px-3 py-2">
                <template v-if="editId === p.id">
                  <select v-model.number="draft.status" class="w-full rounded border border-slate-200 px-2 py-1 text-xs">
                    <option :value="1">启用</option>
                    <option :value="0">停用</option>
                  </select>
                </template>
                <template v-else>{{ p.status === 1 ? '启用' : '停用' }}</template>
              </td>
              <td class="px-3 py-2">
                <div class="flex justify-end gap-2">
                  <template v-if="editId === p.id">
                    <button class="rounded bg-slate-900 px-2 py-1 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving" @click="submitEdit(p.id)">保存</button>
                    <button class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-700 disabled:opacity-40" :disabled="saving" @click="editId = 0">取消</button>
                  </template>
                  <template v-else>
                    <button class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-700" @click="startEdit(p)">编辑</button>
                    <button class="rounded border border-rose-200 bg-rose-50 px-2 py-1 text-xs font-semibold text-rose-700" @click="removePerm(p.id)">删除</button>
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
import type { AdminPermission, RbacAdminMenu } from './menu-management/types'

const saving = ref(false)
const error = ref('')
const q = ref('')
const menuFilter = ref('')
const permissions = ref<AdminPermission[]>([])
const menus = ref<RbacAdminMenu[]>([])
const createOpen = ref(false)
const editId = ref(0)

const draft = reactive({
  perm_key: '',
  label: '',
  category: '',
  menu_key: '',
  status: 1,
})

const menuKeys = computed(() =>
  [...new Set(menus.value.map((m) => (m.menu_key || '').trim()).filter(Boolean))].sort((a, b) => a.localeCompare(b)),
)

const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  return permissions.value
    .filter((p) => {
      if (menuFilter.value === '__empty__') return !(p.menu_key || '').trim()
      if (menuFilter.value && (p.menu_key || '').trim() !== menuFilter.value) return false
      if (!s) return true
      return [p.label, p.perm_key, p.menu_key, p.category].join(' ').toLowerCase().includes(s)
    })
    .sort((a, b) => a.perm_key.localeCompare(b.perm_key))
})

function resetDraft() {
  draft.perm_key = ''
  draft.label = ''
  draft.category = ''
  draft.menu_key = ''
  draft.status = 1
}

function startCreate() {
  editId.value = 0
  createOpen.value = true
  resetDraft()
}

function startEdit(p: AdminPermission) {
  createOpen.value = false
  editId.value = p.id
  draft.perm_key = p.perm_key
  draft.label = p.label
  draft.category = p.category || ''
  draft.menu_key = p.menu_key || ''
  draft.status = p.status
}

async function load() {
  error.value = ''
  try {
    const [pr, mr] = await Promise.all([
      adminGet<{ permissions: AdminPermission[] }>('/v1/admin/rbac/permissions'),
      adminGet<{ menus: RbacAdminMenu[] }>('/v1/admin/rbac/menus'),
    ])
    permissions.value = pr.permissions || []
    menus.value = mr.menus || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    permissions.value = []
    menus.value = []
  }
}

async function submitCreate() {
  const permKey = draft.perm_key.trim()
  const label = draft.label.trim()
  if (!permKey || !label) {
    error.value = 'perm_key 和 名称为必填项'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/rbac/permissions', {
      perm_key: permKey,
      label,
      category: draft.category.trim(),
      menu_key: draft.menu_key.trim(),
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
  const label = draft.label.trim()
  if (!label) {
    error.value = '名称为必填项'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/permissions/${id}`, {
      label,
      category: draft.category.trim(),
      menu_key: draft.menu_key.trim(),
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

async function removePerm(id: number) {
  if (!confirm('删除该功能点？')) return
  saving.value = true
  error.value = ''
  try {
    await adminDelete(`/v1/admin/rbac/permissions/${id}`)
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
