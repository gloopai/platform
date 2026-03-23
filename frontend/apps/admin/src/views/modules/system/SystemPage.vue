<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统管理</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：提供平台<strong>全局展示配置</strong>（国家/货币）与管理员账号只读列表。角色权限、新建账号、改密等为后续迭代。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
      <div class="mb-3 text-sm font-semibold text-slate-800">全局展示配置</div>
      <div class="grid gap-3 sm:grid-cols-3">
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">国家代码</span>
          <input v-model.trim="countryCode" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">货币代码</span>
          <input v-model.trim="currencyCode" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">货币符号</span>
          <input v-model.trim="currencySymbol" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
      </div>
      <div class="mt-3 flex items-center gap-2">
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="saving"
          @click="saveDisplaySettings"
        >
          {{ saving ? '保存中…' : '保存配置' }}
        </button>
        <span v-if="saved" class="text-xs text-emerald-700">已保存</span>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/90 px-4 py-3 text-sm font-semibold text-slate-800">管理员账号</div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[480px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-white text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">ID</th>
              <th class="whitespace-nowrap px-4 py-3">用户名</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td class="px-4 py-8 text-center text-slate-500" colspan="3">加载中…</td>
            </tr>
            <tr v-else-if="!users.length">
              <td class="px-4 py-10 text-center text-slate-500" colspan="3">暂无数据</td>
            </tr>
            <tr v-for="u in users" v-else :key="u.id" class="hover:bg-slate-50/80">
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
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950">
      后续可接入：RBAC、操作审计、参数中心；接口预留如 <span class="font-mono text-xs">POST /v1/admin/admin_users</span> 等。
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet } from '../../../lib/adminApi'
import { adminPut } from '../../../lib/adminApi'
import { applyAdminDisplaySettings } from '../../../lib/displaySettings'

type AdminUser = {
  id: number
  username: string
  status: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref('')
const users = ref<AdminUser[]>([])
const countryCode = ref('CN')
const currencyCode = ref('CNY')
const currencySymbol = ref('¥')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const ds = await adminGet<{ country_code: string; currency_code: string; currency_symbol: string }>('/v1/admin/display_settings')
    countryCode.value = ds.country_code || 'CN'
    currencyCode.value = ds.currency_code || 'CNY'
    currencySymbol.value = ds.currency_symbol || '¥'
    const r = await adminGet<{ users: AdminUser[] }>('/v1/admin/admin_users')
    users.value = r.users ?? []
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    users.value = []
  } finally {
    loading.value = false
  }
}

async function saveDisplaySettings() {
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const r = await adminPut<{ country_code: string; currency_code: string; currency_symbol: string }>('/v1/admin/display_settings', {
      country_code: countryCode.value.trim().toUpperCase(),
      currency_code: currencyCode.value.trim().toUpperCase(),
      currency_symbol: currencySymbol.value.trim(),
    })
    applyAdminDisplaySettings(r)
    countryCode.value = r.country_code
    currencyCode.value = r.currency_code
    currencySymbol.value = r.currency_symbol
    saved.value = true
  } catch {
    error.value = '保存失败，请检查字段与登录态'
  } finally {
    saving.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void load()
  if (registerRefresh) unregister = registerRefresh(() => void load())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
