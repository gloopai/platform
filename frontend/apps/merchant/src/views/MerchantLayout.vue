<template>
  <div class="min-h-full bg-slate-100">
    <div class="border-b border-slate-200 bg-white">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3">
        <div class="text-sm font-semibold text-slate-900">商户平台</div>
        <div class="flex items-center gap-3">
          <div class="text-xs text-slate-500">Merchant Dashboard</div>
          <button class="rounded-md border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50" @click="logout">
            退出登录
          </button>
        </div>
      </div>
    </div>

    <div class="mx-auto grid max-w-7xl grid-cols-12 gap-4 px-4 py-6">
      <aside class="col-span-12 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-3">
        <div class="text-xs font-semibold text-slate-500">菜单</div>
        <nav class="mt-3 grid gap-1">
          <RouterLink class="nav-item" to="/console">控制台</RouterLink>
          <RouterLink class="nav-item" to="/transactions">交易管理</RouterLink>
          <RouterLink class="nav-item" to="/finance">财务中心</RouterLink>
          <RouterLink class="nav-item" to="/developers">开发配置</RouterLink>
        </nav>
      </aside>

      <main class="col-span-12 md:col-span-9">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter, RouterLink, RouterView } from 'vue-router'
import { clearMerchantSession, merchantConsolePost } from '../lib/merchantApi'

const router = useRouter()

async function logout() {
  try {
    await merchantConsolePost<{ ok: boolean }>('/v1/merchant/logout')
  } catch {
  } finally {
    clearMerchantSession()
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
