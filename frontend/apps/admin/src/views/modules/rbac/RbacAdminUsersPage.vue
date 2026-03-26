<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">后台用户</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        查看管理员账号并分配角色；角色的菜单和权限配置在「角色与授权」中维护。
      </p>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div v-if="loading" class="px-5 py-12 text-center text-sm text-slate-500">加载中…</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[640px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3">ID</th>
              <th class="px-4 py-3">用户名</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">角色</th>
              <th class="px-4 py-3 w-40"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="u in users" :key="u.id" class="align-top hover:bg-slate-50/80">
              <td class="px-4 py-3 font-mono text-slate-800">#{{ u.id }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ u.username }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="u.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ u.status === 1 ? '正常' : '停用' }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-slate-700">
                <div v-if="editId === u.id" class="grid gap-2">
                  <label v-for="r in roles" :key="r.id" class="flex cursor-pointer items-center gap-2">
                    <input v-model="editRoleIds" :value="r.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
                    <span>{{ r.name }}</span>
                    <span class="font-mono text-[10px] text-slate-400">{{ r.code }}</span>
                  </label>
                </div>
                <div v-else class="max-w-xs">
                  {{ roleNamesForUser(u.id) }}
                </div>
              </td>
              <td class="px-4 py-3">
                <button
                  v-if="editId !== u.id"
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800"
                  @click="startEdit(u.id)"
                >
                  分配角色
                </button>
                <div v-else class="flex flex-col gap-2">
                  <button
                    type="button"
                    class="rounded-lg bg-slate-900 px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-40"
                    :disabled="saving"
                    @click="saveUserRoles(u.id)"
                  >
                    保存
                  </button>
                  <button type="button" class="text-xs text-slate-500 hover:underline" @click="cancelEdit">取消</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { adminGet, adminPut } from '../../../lib/adminApi'
import { useUiToast } from '../../../composables/useUiToast'

type AdminUser = { id: number; username: string; status: number }
type AdminRole = { id: number; code: string; name: string; status: number }

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const toast = useUiToast()
const users = ref<AdminUser[]>([])
const roles = ref<AdminRole[]>([])
const userRoles = ref<Record<number, number[]>>({})
const editId = ref(0)
const editRoleIds = ref<number[]>([])

function roleNamesForUser(uid: number): string {
  const ids = userRoles.value[uid] || []
  if (!ids.length) return '—'
  const names = ids
    .map((id) => roles.value.find((r) => r.id === id)?.name)
    .filter(Boolean) as string[]
  return names.length ? names.join('、') : '—'
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [ur, rr] = await Promise.all([
      adminGet<{ users: AdminUser[] }>('/v1/admin/admin_users'),
      adminGet<{ roles: AdminRole[] }>('/v1/admin/rbac/roles'),
    ])
    users.value = ur.users || []
    roles.value = rr.roles || []
    const map: Record<number, number[]> = {}
    await Promise.all(
      users.value.map(async (u) => {
        try {
          const r = await adminGet<{ role_ids: number[] }>(`/v1/admin/rbac/admin_users/${u.id}/roles`)
          map[u.id] = r.role_ids || []
        } catch {
          map[u.id] = []
        }
      }),
    )
    userRoles.value = map
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载后台用户失败：${msg}`)
  } finally {
    loading.value = false
  }
}

function startEdit(uid: number) {
  editId.value = uid
  editRoleIds.value = (userRoles.value[uid] || []).slice()
}

function cancelEdit() {
  editId.value = 0
  editRoleIds.value = []
}

async function saveUserRoles(uid: number) {
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/rbac/admin_users/${uid}/roles`, { role_ids: editRoleIds.value })
    const r = await adminGet<{ role_ids: number[] }>(`/v1/admin/rbac/admin_users/${uid}/roles`)
    userRoles.value = { ...userRoles.value, [uid]: r.role_ids || [] }
    cancelEdit()
    toast.success('角色分配已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存角色分配失败：${msg}`)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
