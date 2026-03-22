<template>
  <div
    class="merchant-app flex h-screen w-full flex-col overflow-hidden bg-gradient-to-br from-slate-50 via-white to-emerald-50/40 text-slate-900"
  >
    <header
      class="relative z-10 shrink-0 border-b border-emerald-100/80 bg-white/90 px-4 py-3 shadow-sm backdrop-blur-md sm:px-6"
    >
      <div class="flex items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 via-teal-500 to-cyan-600 text-sm font-bold text-white shadow-lg shadow-emerald-500/25"
            aria-hidden="true"
          >
            P
          </div>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold tracking-tight text-slate-900">商户中心</div>
            <div class="truncate text-xs text-slate-500">Partner Console · 支付接入与运营</div>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <!-- sm～md 仅顶栏展示商户（无侧栏）；md+ 由左侧身份卡展示，避免重复 -->
          <div class="hidden text-right sm:block md:hidden">
            <div class="text-xs font-semibold leading-tight text-slate-900">{{ merchantDisplayName }}</div>
            <div class="mt-0.5 font-mono text-[10px] leading-none text-slate-400">{{ merchantIdDisplay }}</div>
          </div>
          <button
            type="button"
            class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-500/30"
            @click="logout"
          >
            退出
          </button>
        </div>
      </div>
    </header>

    <!-- 小屏：当前商户身份（与侧栏桌面区信息一致） -->
    <div
      class="flex items-center gap-3 border-b border-emerald-100/60 bg-gradient-to-r from-emerald-50/80 to-white px-4 py-3 sm:hidden"
    >
      <div
        class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 text-sm font-bold text-white shadow-md shadow-emerald-600/20"
      >
        {{ monogram }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="truncate text-sm font-semibold text-slate-900">{{ merchantDisplayName }}</div>
        <div class="mt-0.5 truncate font-mono text-[11px] text-slate-500">{{ merchantIdDisplay }}</div>
      </div>
    </div>

    <div class="flex min-h-0 min-w-0 flex-1 gap-4 p-4 sm:p-5 lg:p-6">
      <aside
        class="hidden w-[248px] shrink-0 flex-col rounded-2xl border border-slate-200/90 bg-white/90 p-3 shadow-sm backdrop-blur-sm md:flex"
      >
        <div class="mb-3 rounded-2xl border border-emerald-100/90 bg-gradient-to-br from-emerald-50/95 via-white to-teal-50/40 p-3 shadow-sm">
          <div class="flex items-center gap-3">
            <div
              class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 text-base font-bold text-white shadow-md shadow-emerald-600/25"
            >
              {{ monogram }}
            </div>
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-semibold leading-snug text-slate-900">{{ merchantDisplayName }}</div>
              <div class="mt-0.5 truncate font-mono text-[10px] leading-tight text-slate-500" :title="merchantIdDisplay">
                {{ merchantIdDisplay }}
              </div>
            </div>
          </div>
        </div>
        <div class="px-2 pb-2 pt-0.5 text-[10px] font-semibold uppercase tracking-wider text-slate-400">菜单</div>
        <nav class="flex flex-col gap-0.5">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-slate-600 transition hover:bg-emerald-50/90 hover:text-emerald-900"
            active-class="nav-item-active"
          >
            <span class="text-slate-400 group-hover:text-emerald-600" v-html="item.icon" />
            <span class="truncate">{{ item.label }}</span>
          </RouterLink>
        </nav>
        <div class="mt-auto border-t border-slate-100 pt-3">
          <p class="px-2 text-[10px] leading-relaxed text-slate-400">
            需要帮助？请查看「开发配置」中的联调说明与签名示例。
          </p>
        </div>
      </aside>

      <main class="min-h-0 min-w-0 flex-1 overflow-y-auto rounded-2xl border border-slate-200/90 bg-white/80 p-4 shadow-sm backdrop-blur-sm sm:p-6 lg:p-8">
        <RouterView />
      </main>
    </div>

    <!-- 小屏底部导航 -->
    <nav
      class="flex shrink-0 border-t border-slate-200/90 bg-white/95 px-2 py-2 backdrop-blur-md md:hidden"
    >
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="flex flex-1 flex-col items-center gap-0.5 rounded-lg py-1.5 text-[10px] font-medium text-slate-500 transition"
        active-class="nav-mb-active"
      >
        <span class="opacity-80" v-html="item.iconSm" />
        <span class="truncate px-0.5">{{ item.short }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, RouterLink, RouterView } from 'vue-router'
import {
  clearMerchantSession,
  loadMerchantAuth,
  merchantConsolePost,
  merchantMonogram,
  resolveMerchantDisplayName,
} from '../lib/merchantApi'

const router = useRouter()

const merchantIdDisplay = computed(() => {
  try {
    return loadMerchantAuth().merchantId || ''
  } catch {
    return ''
  }
})

const merchantDisplayName = computed(() => resolveMerchantDisplayName(merchantIdDisplay.value))

const monogram = computed(() => merchantMonogram(merchantDisplayName.value, merchantIdDisplay.value))

const navItems = [
  {
    to: '/console',
    label: '控制台',
    short: '首页',
    icon: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>`,
    iconSm: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>`,
  },
  {
    to: '/transactions',
    label: '交易管理',
    short: '交易',
    icon: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" /></svg>`,
    iconSm: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /></svg>`,
  },
  {
    to: '/finance',
    label: '财务中心',
    short: '财务',
    icon: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
    iconSm: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
  },
  {
    to: '/developers',
    label: '开发配置',
    short: '开发',
    icon: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>`,
    iconSm: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>`,
  },
]

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
.nav-item-active {
  @apply border border-emerald-200/90 bg-gradient-to-r from-emerald-50 to-teal-50/80 text-emerald-900 shadow-sm;
}
.nav-item-active span:first-child {
  @apply text-emerald-600;
}
.nav-mb-active {
  @apply bg-emerald-50 text-emerald-800;
}
</style>
