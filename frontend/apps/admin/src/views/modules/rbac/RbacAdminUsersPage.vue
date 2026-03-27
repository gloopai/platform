<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">后台用户</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        选择用户后在右侧配置状态与角色；菜单与权限仍在「角色与授权」中维护。
      </p>
    </div>

    <div class="grid gap-6 lg:grid-cols-12">
      <div class="lg:col-span-3">
        <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
          <div class="border-b border-slate-100 bg-slate-50/90 px-4 py-3">
            <div class="text-sm font-semibold text-slate-800">用户</div>
          </div>
          <div class="p-4">
            <div class="mb-2 text-xs font-semibold text-slate-700">新建用户</div>
            <div class="grid gap-2">
              <label class="grid gap-1 text-xs font-medium text-slate-600">
                用户名（创建后不可改）
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
              <div>
                <div class="mb-1 text-xs font-medium text-slate-600">初始角色</div>
                <div class="flex max-h-28 flex-col gap-1.5 overflow-y-auto pr-0.5">
                  <label v-for="r in roles" :key="'new_' + r.id" class="inline-flex cursor-pointer items-center gap-1.5 text-xs">
                    <input v-model="newRoleIds" :value="r.id" type="checkbox" class="h-4 w-4 shrink-0 rounded border-slate-300" />
                    <span class="min-w-0 truncate">{{ r.name }}</span>
                  </label>
                </div>
              </div>
            </div>
            <button
              type="button"
              class="mt-3 w-full rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving || !newUsername || !newPassword"
              @click="createUser"
            >
              创建用户
            </button>
          </div>
          <div class="border-t border-slate-100">
            <div v-if="loading" class="px-4 py-10 text-center text-sm text-slate-500">加载中…</div>
            <div v-else class="max-h-[min(420px,62vh)] divide-y divide-slate-100 overflow-y-auto">
              <div
                v-for="u in users"
                :key="u.id"
                type="button"
                class="flex w-full cursor-pointer items-center justify-between gap-3 px-4 py-3 text-left text-sm transition hover:bg-slate-50"
                :class="selectedUserId === u.id ? 'bg-indigo-50' : ''"
                @click="selectUser(u.id)"
              >
                <div class="min-w-0">
                  <div class="truncate font-semibold text-slate-900">{{ u.username }}</div>
                  <div class="truncate font-mono text-[11px] text-slate-500">#{{ u.id }}</div>
                </div>
                <div class="flex shrink-0 flex-col items-end gap-1">
                  <span
                    class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                    :class="u.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                  >
                    {{ u.status === 1 ? '正常' : '停用' }}
                  </span>
                  <span
                    class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                    :class="u.mfa_enabled === 1 ? 'bg-indigo-100 text-indigo-800' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ u.mfa_enabled === 1 ? 'MFA' : '无 MFA' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-9 space-y-6">
        <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold text-slate-800">账号配置</div>
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm disabled:opacity-40"
              :disabled="!selectedUserId || saving"
              @click="reloadSelectedGrants"
            >
              重新加载角色
            </button>
          </div>

          <div v-if="!selectedUserId" class="mt-8 rounded-xl border border-slate-200 bg-slate-50/50 px-4 py-12 text-center text-sm text-slate-500">
            请选择左侧用户
          </div>

          <div v-else-if="!selectedUser" class="mt-8 rounded-xl border border-slate-200 bg-amber-50/50 px-4 py-10 text-center text-sm text-amber-900">
            所选用户已不存在，请重新选择左侧列表。
          </div>

          <template v-else>
            <div class="mt-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-3 text-xs text-slate-600">
              <div class="flex flex-wrap items-center gap-2">
                <span class="font-semibold text-slate-900">{{ selectedUser.username }}</span>
                <span class="font-mono text-slate-500">#{{ selectedUser.id }}</span>
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                  :class="selectedUser.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ selectedUser.status === 1 ? '正常' : '停用' }}
                </span>
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                  :class="selectedUser.mfa_enabled === 1 ? 'bg-indigo-100 text-indigo-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ selectedUser.mfa_enabled === 1 ? 'MFA 已启用' : 'MFA 未启用' }}
                </span>
              </div>
            </div>

            <div class="mt-4 space-y-4">
              <label class="grid w-full max-w-xs gap-1 text-xs font-medium text-slate-700 sm:w-auto">
                状态
                <select v-model.number="editStatus" class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm" :disabled="saving">
                  <option :value="1">正常</option>
                  <option :value="0">停用</option>
                </select>
              </label>

              <div>
                <div class="mb-2 text-xs font-semibold text-slate-800">角色</div>
                <div class="grid gap-2 sm:grid-cols-2">
                  <label
                    v-for="r in roles"
                    :key="'e_' + r.id"
                    class="flex cursor-pointer items-center gap-2 rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2 text-xs hover:border-slate-300"
                  >
                    <input v-model="editRoleIds" :value="r.id" type="checkbox" class="h-4 w-4 rounded border-slate-300" :disabled="saving" />
                    <span class="min-w-0 font-medium text-slate-800">{{ r.name }}</span>
                    <span class="ml-auto shrink-0 font-mono text-[10px] text-slate-400">{{ r.code }}</span>
                  </label>
                </div>
              </div>

              <button
                type="button"
                class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="saving"
                @click="saveUserEdit"
              >
                {{ saving ? '保存中…' : '保存配置' }}
              </button>
            </div>

            <div class="mt-6 border-t border-slate-100 pt-4">
              <div class="text-xs font-semibold text-slate-700">更多操作</div>
              <div class="mt-3 flex flex-wrap gap-2">
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-800 shadow-sm"
                  :disabled="saving"
                  @click="openResetPwd(selectedUserId)"
                >
                  重置密码
                </button>
                <button
                  v-if="selectedUser.mfa_enabled !== 1"
                  type="button"
                  class="rounded-lg border border-indigo-200 bg-indigo-50 px-3 py-2 text-xs font-semibold text-indigo-800"
                  :disabled="saving"
                  @click="setupMfa(selectedUserId)"
                >
                  绑定 MFA
                </button>
                <button
                  v-else
                  type="button"
                  class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800"
                  :disabled="saving"
                  @click="disableMfa(selectedUserId)"
                >
                  禁用 MFA
                </button>
                <button
                  type="button"
                  class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-semibold text-rose-800"
                  :disabled="saving"
                  @click="deleteUser(selectedUserId)"
                >
                  删除用户
                </button>
              </div>
            </div>
          </template>
        </div>

        <div
          v-if="mfaUserId > 0 && mfaQrDataUrl"
          class="rounded-2xl border border-indigo-200 bg-indigo-50/40 p-4 shadow-sm"
        >
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
                <button
                  type="button"
                  class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
                  :disabled="saving || !mfaCode"
                  @click="confirmMfa"
                >
                  确认绑定
                </button>
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
                  @click="clearMfaSetup"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

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
const selectedUserId = ref(0)
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

const selectedUser = computed(() => users.value.find((u) => u.id === selectedUserId.value) || null)

function syncEditFromSelected() {
  const u = selectedUser.value
  if (!u) {
    editRoleIds.value = []
    editStatus.value = 1
    return
  }
  editRoleIds.value = (userRoles.value[u.id] || []).slice()
  editStatus.value = u.status
}

function selectUser(id: number) {
  selectedUserId.value = id
  syncEditFromSelected()
}

async function reloadSelectedGrants() {
  if (!selectedUserId.value) return
  try {
    const r = await adminGet<{ role_ids: number[] }>(`/v1/admin/rbac/admin_users/${selectedUserId.value}/roles`)
    userRoles.value = { ...userRoles.value, [selectedUserId.value]: r.role_ids || [] }
    syncEditFromSelected()
    toast.success('角色已重新加载')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`加载角色失败：${msg}`)
  }
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

    if (selectedUserId.value && !users.value.some((u) => u.id === selectedUserId.value)) {
      selectedUserId.value = 0
    }
    if (!selectedUserId.value && users.value.length) {
      selectedUserId.value = users.value[0]!.id
    }
    syncEditFromSelected()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载后台用户失败：${msg}`)
  } finally {
    loading.value = false
  }
}

async function saveUserEdit() {
  const uid = selectedUserId.value
  if (!uid) return
  saving.value = true
  error.value = ''
  try {
    await adminPut(`/v1/admin/admin_users/${uid}`, { status: editStatus.value, role_ids: editRoleIds.value })
    await adminPut(`/v1/admin/rbac/admin_users/${uid}/roles`, { role_ids: editRoleIds.value })
    const r = await adminGet<{ role_ids: number[] }>(`/v1/admin/rbac/admin_users/${uid}/roles`)
    userRoles.value = { ...userRoles.value, [uid]: r.role_ids || [] }
    await load()
    toast.success('用户配置已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存用户配置失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function createUser() {
  saving.value = true
  error.value = ''
  const createdUsername = newUsername.value.trim()
  try {
    await adminPost('/v1/admin/admin_users', {
      username: createdUsername,
      password: newPassword.value,
      status: newStatus.value,
      role_ids: newRoleIds.value,
    })
    newUsername.value = ''
    newPassword.value = ''
    newStatus.value = 1
    newRoleIds.value = []
    await load()
    const created = users.value.find((u) => u.username === createdUsername)
    if (created) {
      selectedUserId.value = created.id
      syncEditFromSelected()
    }
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
    if (selectedUserId.value === uid) {
      selectedUserId.value = 0
    }
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
    selectedUserId.value = uid
    syncEditFromSelected()
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
