<template>
  <div class="space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-gradient-to-br from-slate-50 to-white p-5 shadow-sm">
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">权限与安全</h1>
      <p class="mt-1 max-w-3xl text-sm leading-relaxed text-slate-600">
        用「页面」把能力串起来：<strong>菜单管理</strong>（侧栏 / 头像菜单 / 其它功能）决定入口与能力清单；再给<strong>角色</strong>勾选菜单与权限即可。
      </p>
    </div>

    <div class="flex flex-wrap gap-2 border-b border-slate-200 pb-3">
      <RouterLink
        v-for="t in tabs"
        :key="t.to"
        :to="t.to"
        class="rounded-lg px-3 py-2 text-sm font-medium transition"
        :class="
          isActive(t.to)
            ? 'bg-indigo-600 text-white shadow-sm'
            : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'
        "
      >
        {{ t.label }}
      </RouterLink>
    </div>

    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()

const tabs = [
  { to: '/rbac/overview', label: '配置总览' },
  { to: '/rbac/menus', label: '菜单管理' },
  { to: '/rbac/roles', label: '角色与授权' },
  { to: '/rbac/admin-users', label: '后台用户' },
]

function isActive(to: string) {
  const p = route.path
  if (to === '/rbac/overview') return p === '/rbac' || p === '/rbac/overview'
  return p === to || p.startsWith(`${to}/`)
}
</script>
