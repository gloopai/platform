<template>
  <div class="min-h-full bg-slate-100">
    <div class="border-b border-slate-200 bg-white">
      <div class="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-3 px-4 py-3">
        <div class="flex items-center gap-3">
          <div class="text-sm font-semibold text-slate-900">总管理台</div>
          <div class="text-xs text-slate-500">Admin System</div>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
            @click="broadcastRefresh"
          >
            刷新
          </button>
          <button
            class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
            @click="logout"
          >
            退出登录
          </button>
        </div>
      </div>
    </div>

    <div class="mx-auto grid max-w-7xl grid-cols-12 gap-4 px-4 py-6">
      <aside class="col-span-12 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-3">
        <div class="text-xs font-semibold text-slate-500">菜单</div>
        <nav class="mt-3 grid gap-1">
          <RouterLink class="nav-item" to="/stats">系统概览</RouterLink>
          <RouterLink class="nav-item" to="/merchants">商户管理</RouterLink>
          <RouterLink class="nav-item" to="/channels">通道/路由</RouterLink>
          <RouterLink class="nav-item" to="/audit">运营与审计</RouterLink>
        </nav>
      </aside>

      <main class="col-span-12 md:col-span-9">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { provide, ref } from 'vue'
import { useRouter, RouterLink, RouterView } from 'vue-router'
import { adminPost, clearAdminSession, loadAdminToken } from '../lib/adminApi'

type RefreshFn = () => void

const router = useRouter()
const adminToken = ref(loadAdminToken())

const refreshFns = ref<RefreshFn[]>([])
provide('adminToken', adminToken)
provide('registerRefresh', (fn: RefreshFn) => {
  refreshFns.value.push(fn)
  return () => {
    refreshFns.value = refreshFns.value.filter((x) => x !== fn)
  }
})

function broadcastRefresh() {
  for (const fn of refreshFns.value) fn()
}

async function logout() {
  try {
    await adminPost<{ ok: boolean }>('/v1/admin/logout')
  } catch {
  } finally {
    clearAdminSession()
    await router.replace('/login')
  }
}
</script>

<style scoped>
.nav-item {
  display: block;
  border-radius: 0.75rem;
  padding: 0.625rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(51 65 85);
}
.nav-item:hover {
  background: rgb(248 250 252);
}
.router-link-active {
  background: rgb(15 23 42);
  color: white;
}
</style>
