<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统管理</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：只读展示平台<strong>管理员账号</strong>（来自 <code class="rounded bg-slate-100 px-1 py-0.5 font-mono text-xs">admin_users</code>）。角色权限、新建账号、改密等为后续迭代。
      </p>
      <p v-if="error" class="mt-2 text-sm text-rose-600">{{ error }}</p>
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

type AdminUser = {
  id: number
  username: string
  status: number
}

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const loading = ref(true)
const error = ref('')
const users = ref<AdminUser[]>([])

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await adminGet<{ users: AdminUser[] }>('/v1/admin/admin_users')
    users.value = r.users ?? []
  } catch {
    error.value = '加载失败，请检查登录态与网关'
    users.value = []
  } finally {
    loading.value = false
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
