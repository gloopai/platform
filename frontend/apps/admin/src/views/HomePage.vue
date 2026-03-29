<template>
  <div class="mx-auto max-w-4xl space-y-8">
    <div
      class="overflow-hidden rounded-2xl border border-slate-200/90 bg-gradient-to-br from-slate-900 via-slate-900 to-indigo-950 px-6 py-8 text-white shadow-lg shadow-slate-900/20 sm:px-10 sm:py-10"
    >
      <p class="text-[10px] font-semibold uppercase tracking-[0.2em] text-indigo-200/90">工作台</p>
      <h1 class="mt-2 text-2xl font-semibold tracking-tight sm:text-3xl">
        {{ greeting }}，{{ displayName }}
      </h1>
      <p class="mt-3 max-w-xl text-sm leading-relaxed text-slate-300">
        这是平台管理脚手架的默认首页。侧栏可进入权限配置、系统参数与运维探活；生产环境变更角色与接口规则前请二次确认。
      </p>
    </div>

    <div class="grid gap-4 sm:grid-cols-3">
      <RouterLink
        v-for="card in quickLinks"
        :key="card.to"
        :to="card.to"
        class="group rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-indigo-200 hover:shadow-md"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="text-sm font-semibold text-slate-900 group-hover:text-indigo-700">{{ card.title }}</div>
            <div class="mt-1 text-xs leading-relaxed text-slate-500">{{ card.desc }}</div>
          </div>
          <span class="rounded-lg bg-slate-100 px-2 py-1 text-[10px] font-medium text-slate-600 group-hover:bg-indigo-50 group-hover:text-indigo-700">
            进入
          </span>
        </div>
      </RouterLink>
    </div>

    <div v-if="loadError" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
      {{ loadError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { adminGet } from '../lib/adminApi'

const displayName = ref('管理员')
const loadError = ref('')

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '上午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const quickLinks = [
  { to: '/rbac/menus', title: '权限与安全', desc: '菜单、角色、功能点与接口规则' },
  { to: '/system', title: '系统管理', desc: '展示类全局配置' },
  { to: '/ops', title: '运维监控', desc: '依赖服务与节点状态' },
] as const

onMounted(async () => {
  try {
    const me = await adminGet<{ username: string; email: string; display_name: string }>('/v1/admin/me')
    const next = (me.display_name || me.email || me.username || '').trim()
    if (next) displayName.value = next
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : '无法加载当前用户信息'
  }
})
</script>
