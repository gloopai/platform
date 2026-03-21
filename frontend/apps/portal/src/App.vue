<template>
  <div class="min-h-full bg-slate-50">
    <header class="border-b border-slate-200 bg-white">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
        <RouterLink class="text-sm font-semibold text-slate-900" to="/">gloop pay</RouterLink>
        <nav class="flex items-center gap-4 text-sm">
          <RouterLink class="font-medium text-slate-700 hover:text-slate-900" to="/">首页</RouterLink>
          <RouterLink class="font-medium text-slate-700 hover:text-slate-900" to="/products">产品与服务</RouterLink>
          <RouterLink class="font-medium text-slate-700 hover:text-slate-900" to="/docs">开发者中心</RouterLink>
          <RouterLink class="font-medium text-slate-700 hover:text-slate-900" to="/about">关于与合规</RouterLink>
        </nav>
      </div>
    </header>
    <main>
      <RouterView />
    </main>
    <footer class="border-t border-slate-200 bg-white">
      <div class="mx-auto max-w-6xl px-4 py-6 text-xs text-slate-500">© gloop pay</div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>
