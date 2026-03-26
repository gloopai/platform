<template>
  <div class="space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-gradient-to-br from-slate-50 to-white p-5 shadow-sm">
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">权限与安全</h1>
      <p class="mt-1 max-w-3xl text-sm leading-relaxed text-slate-600">
        分步配置：<strong>角色</strong>绑定<strong>菜单</strong>（能进哪些页）与<strong>权限点</strong>（能调哪些接口）；<strong>接口规则</strong>把网关路径映射到权限点，上线新接口后只需在此维护规则，无需改代码。
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
  { to: '/rbac/overview', label: '概览' },
  { to: '/rbac/roles', label: '角色与授权' },
  { to: '/rbac/permissions', label: '权限点' },
  { to: '/rbac/api-rules', label: '接口规则' },
]

function isActive(to: string) {
  const p = route.path
  if (to === '/rbac/overview') return p === '/rbac' || p === '/rbac/overview'
  return p === to || p.startsWith(`${to}/`)
}
</script>
