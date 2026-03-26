<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">后台用户</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        查看管理员账号并分配角色；角色的菜单和权限配置在「角色与授权」中维护。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
      <div class="mb-3 text-sm font-semibold text-slate-800">新增后台用户</div>
      <div class="grid gap-3 md:grid-cols-5">
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          用户名（创建后不可修改）
          <input v-model.trim="newUsername" type="text" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          初始密码
          <input v-model.trim="newPassword" type="password" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </label>
        <label class="grid gap-1 text-xs font-medium text-slate-600">
          状态
          <select v-model.number="newStatus" class="rounded-lg border border-slate-200 px-3 py-2 text-sm">
            <option :value="1">正常</option>
            <option :value="0">停用</option>
          </select>
        </label>
        <div class="md:col-span-2">
          <div class="mb-1 text-xs font-medium text-slate-600">初始角色</div>
          <div class="flex flex-wrap gap-3">
            <label v-for="r in roles" :key="'new_' + r.id" class="inline-flex items-center gap-1.5 text-xs">
              <input v-model="newRoleIds" :value="r.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" />
              <span>{{ r.name }}</span>
            </label>
          </div>
        </div>
      </div>
      <div class="mt-3">
        <button type="button" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving || !newUsername || !newPassword" @click="createUser">
          新建用户
        </button>
      </div>
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
              <th class="px-4 py-3">MFA</th>
              <th class="px-4 py-3">角色</th>
              <th class="px-4 py-3 w-[360px]"></th>
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
              <td class="px-4 py-3">
                <span class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold" :class="u.mfa_enabled === 1 ? 'bg-indigo-100 text-indigo-800' : 'bg-slate-200 text-slate-700'">
                  {{ u.mfa_enabled === 1 ? '已启用' : '未启用' }}
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
                <div v-if="editId !== u.id" class="flex flex-wrap gap-2">
                  <button type="button" class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800" @click="startEdit(u.id, u.status)">
                    编辑
                  </button>
                  <button type="button" class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800" @click="openResetPwd(u.id)">
                    改密
                  </button>
                  <button v-if="u.mfa_enabled !== 1" type="button" class="rounded-lg border border-indigo-200 bg-indigo-50 px-3 py-1.5 text-xs font-semibold text-indigo-800" @click="setupMfa(u.id)">
                    绑定 MFA
                  </button>
                  <button v-else type="button" class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-1.5 text-xs font-semibold text-amber-800" @click="disableMfa(u.id)">
                    禁用 MFA
                  </button>
                  <button type="button" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-1.5 text-xs font-semibold text-rose-800" @click="deleteUser(u.id)">
                    删除
                  </button>
                </div>
                <div v-else class="flex flex-col gap-2">
                  <div class="flex items-center gap-2">
                    <span class="text-xs text-slate-500">状态</span>
                    <select v-model.number="editStatus" class="rounded border border-slate-200 px-2 py-1 text-xs">
                      <option :value="1">正常</option>
                      <option :value="0">停用</option>
                    </select>
                  </div>
                  <div class="flex gap-2">
                    <button type="button" class="rounded-lg bg-slate-900 px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving" @click="saveUserEdit(u.id)">
                      保存
                    </button>
                    <button type="button" class="text-xs text-slate-500 hover:underline" @click="cancelEdit">取消</button>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="mfaUserId > 0 && mfaQrDataUrl" class="rounded-2xl border border-indigo-200 bg-indigo-50/40 p-4 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">绑定 MFA</div>
      <p class="mt-1 text-xs text-slate-600">请使用 Authenticator 扫码后输入 6 位验证码完成绑定。</p>
      <div class="mt-3 flex flex-wrap items-start gap-4">
        <img :src="mfaQrDataUrl" alt="mfa qrcode" class="h-40 w-40 rounded-lg border border-slate-200 bg-white p-2" />
        <div class="min-w-[240px] space-y-2">
          <div class="text-xs text-slate-600">密钥：<span class="font-mono text-slate-800">{{ mfaSecret }}</span></div>
          <label class="grid gap-1 text-xs font-medium text-slate-600">
            验证码
            <input v-model.trim="mfaCode" maxlength="8" class="rounded border border-slate-200 px-3 py-2 font-mono text-sm" />
          </label>
          <div class="flex gap-2">
            <button type="button" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40" :disabled="saving || !mfaCode" @click="confirmMfa">
              确认绑定
            </button>
            <button type="button" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700" @click="clearMfaSetup">
              取消
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { adminDelete, adminGet, adminPost, adminPut } from '../../../lib/adminApi'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'

type AdminUser = { id: number; username: string; status: number; mfa_enabled: number }
type AdminRole = { id: number; code: string; name: string; status: number }

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const toast = useUiToast()
const dialog = useUiDialog()
const users = ref<AdminUser[]>([])
const roles = ref<AdminRole[]>([])
const userRoles = ref<Record<number, number[]>>({})
const editId = ref(0)
const editRoleIds = ref<number[]>([])
const editStatus = ref(1)
const newUsername = ref('')
const newPassword = ref('')
const newStatus = ref(1)
const newRoleIds = ref<number[]>([])
const mfaUserId = ref(0)
const mfaSecret = ref('')
const mfaQrDataUrl = ref('')
const mfaCode = ref('')
const resetPwdUserId = ref(0)
const resetPwd = ref('')

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

function startEdit(uid: number, status: number) {
  editId.value = uid
  editRoleIds.value = (userRoles.value[uid] || []).slice()
  editStatus.value = status
}

function cancelEdit() {
  editId.value = 0
  editRoleIds.value = []
}

async function saveUserEdit(uid: number) {
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/admin_users/${uid}`, { status: editStatus.value, role_ids: editRoleIds.value })
    await adminPut(`/v1/admin/rbac/admin_users/${uid}/roles`, { role_ids: editRoleIds.value })
    const r = await adminGet<{ role_ids: number[] }>(`/v1/admin/rbac/admin_users/${uid}/roles`)
    userRoles.value = { ...userRoles.value, [uid]: r.role_ids || [] }
    cancelEdit()
    await load()
    toast.success('用户配置已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存角色分配失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function createUser() {
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/admin_users', {
      username: newUsername.value.trim(),
      password: newPassword.value,
      status: newStatus.value,
      role_ids: newRoleIds.value,
    })
    newUsername.value = ''
    newPassword.value = ''
    newStatus.value = 1
    newRoleIds.value = []
    await load()
    toast.success('后台用户已创建')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`创建后台用户失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function openResetPwd(uid: number) {
  resetPwdUserId.value = uid
  resetPwd.value = ''
  void doResetPwd()
}

async function doResetPwd() {
  if (!resetPwdUserId.value) return
  const ok = await dialog.open({
    title: '重置密码',
    message: '请输入新密码后点击“确定”进行重置。',
    confirmText: '继续',
    cancelText: '取消',
  })
  if (!ok) return
  const pwd = window.prompt('请输入新密码（至少 6 位）')
  if (!pwd || pwd.trim().length < 6) {
    toast.error('密码长度至少 6 位')
    return
  }
  saving.value = true
  try {
    await adminPost(`/v1/admin/admin_users/${resetPwdUserId.value}/reset_password`, { password: pwd.trim() })
    toast.success('密码已重置')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`重置密码失败：${msg}`)
  } finally {
    saving.value = false
    resetPwdUserId.value = 0
  }
}

async function deleteUser(uid: number) {
  const ok = await dialog.danger('删除该后台用户？此操作不可恢复。', '删除后台用户')
  if (!ok) return
  saving.value = true
  try {
    await adminDelete(`/v1/admin/admin_users/${uid}`)
    await load()
    toast.success('后台用户已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`删除后台用户失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function setupMfa(uid: number) {
  saving.value = true
  try {
    const r = await adminPost<{ secret: string; qr_data_url: string }>(`/v1/admin/admin_users/${uid}/mfa/setup`, {})
    mfaUserId.value = uid
    mfaSecret.value = r.secret
    mfaQrDataUrl.value = r.qr_data_url
    mfaCode.value = ''
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`生成 MFA 二维码失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function confirmMfa() {
  if (!mfaUserId.value) return
  saving.value = true
  try {
    await adminPost(`/v1/admin/admin_users/${mfaUserId.value}/mfa/confirm`, { code: mfaCode.value.trim() })
    clearMfaSetup()
    await load()
    toast.success('MFA 已绑定')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`绑定 MFA 失败：${msg}`)
  } finally {
    saving.value = false
  }
}

function clearMfaSetup() {
  mfaUserId.value = 0
  mfaSecret.value = ''
  mfaQrDataUrl.value = ''
  mfaCode.value = ''
}

async function disableMfa(uid: number) {
  const ok = await dialog.confirm('禁用该用户 MFA？禁用后登录无需验证码。', '禁用 MFA')
  if (!ok) return
  saving.value = true
  try {
    await adminPost(`/v1/admin/admin_users/${uid}/mfa/disable`, {})
    await load()
    toast.success('MFA 已禁用')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`禁用 MFA 失败：${msg}`)
  } finally {
    saving.value = false
  }
}

onMounted(() => void load())
</script>
