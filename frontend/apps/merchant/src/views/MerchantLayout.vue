<template>
  <div
    class="merchant-app flex h-screen w-full flex-col overflow-hidden bg-gradient-to-br from-slate-100/90 via-white to-slate-50/80 text-slate-900"
  >
    <!-- 小屏：极薄顶栏（品牌 + 账户菜单）；桌面端品牌与退出在侧栏，主区无全局 Header -->
    <header
      class="relative z-20 flex h-12 shrink-0 items-center justify-between border-b border-slate-200/80 bg-white/85 px-3 backdrop-blur-md md:hidden"
    >
      <div class="flex min-w-0 items-center gap-2.5">
        <div
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-slate-600 to-slate-800 text-xs font-bold text-white shadow-md shadow-slate-900/15"
          aria-hidden="true"
        >
          P
        </div>
        <div class="min-w-0">
          <div class="truncate text-sm font-semibold tracking-tight text-slate-900">商户中心</div>
          <div class="truncate text-[10px] font-mono text-slate-500"> {{ serverTimeText }}</div>
        </div>
      </div>
      <details ref="mobileAccountDetailsRef" class="relative">
        <summary
          class="flex cursor-pointer list-none items-center gap-0 rounded-full outline-none ring-slate-400/40 focus-visible:ring-2 [&::-webkit-details-marker]:hidden"
        >
          <span
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-slate-600 to-slate-800 text-xs font-bold text-white shadow-md shadow-slate-900/15"
            :title="merchantDisplayName"
          >
            {{ monogram }}
          </span>
        </summary>
        <div
          class="absolute right-0 top-[calc(100%+0.35rem)] w-[min(17.5rem,calc(100vw-1.5rem))] overflow-hidden rounded-2xl border border-slate-200/90 bg-white py-1 shadow-lg shadow-slate-900/10"
          role="menu"
        >
          <div class="border-b border-slate-100 px-3 py-2.5">
            <div class="truncate text-sm font-semibold text-slate-900">{{ merchantDisplayName }}</div>
            <div class="mt-0.5 truncate font-mono text-[11px] text-slate-500" :title="merchantIdDisplay">
              {{ merchantIdDisplay }}
            </div>
          </div>
          <button
            type="button"
            class="flex w-full items-center gap-2 px-3 py-2.5 text-left text-sm font-medium text-slate-700 transition hover:bg-slate-50"
            role="menuitem"
            @click="onMobileLogout"
          >
            退出登录
          </button>
        </div>
      </details>
    </header>

    <div class="flex min-h-0 min-w-0 flex-1 gap-4 p-4 sm:p-5 lg:p-6">
      <aside
        class="hidden w-[248px] shrink-0 flex-col overflow-hidden rounded-2xl border border-slate-200/90 bg-white/90 p-3 shadow-sm backdrop-blur-sm md:flex"
      >
        <!-- 单一身份区：产品名作眉题，避免与下方「头像+两行字」重复一套版式 -->
        <div class="mb-3 rounded-2xl border border-slate-200/90 bg-gradient-to-br from-slate-50/95 via-white to-slate-100/50 p-3 shadow-sm">
          <p class="mb-2.5 truncate text-[10px] font-medium text-slate-400">
            商户中心 <span class="text-slate-300">·</span> Partner Console
          </p>
          <div class="flex items-center gap-3">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-slate-600 to-slate-800 text-sm font-bold text-white shadow-md shadow-slate-900/15"
              aria-hidden="true"
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
          <div class="mt-2 border-t border-slate-200/80 pt-2 text-[11px] leading-tight">
            <span class="text-slate-400">服务器时间</span>
            <div class="mt-0.5 font-mono tabular-nums text-slate-700">{{ serverTimeText }}</div>
          </div>
        </div>
        <div class="min-h-0 flex flex-1 flex-col overflow-hidden">
          <div class="px-2 pb-2 pt-0.5 text-[10px] font-semibold uppercase tracking-wider text-slate-400">菜单</div>
          <nav class="merchant-menu-scroll min-h-0 flex-1 overflow-y-auto pr-1 pb-2">
            <RouterLink
              v-for="item in merchantNavItems"
              :key="item.to"
              :to="item.to"
              class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-slate-600 transition hover:bg-slate-100 hover:text-slate-900"
              active-class="nav-item-active"
            >
              <span class="text-slate-400 group-hover:text-slate-700" v-html="item.icon" />
              <span class="truncate">{{ item.label }}</span>
            </RouterLink>
          </nav>
        </div>
        <div class="shrink-0 flex flex-col gap-2 border-t border-slate-100 pt-3">
          <button
            type="button"
            class="w-full rounded-xl border border-slate-200/90 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-400/35"
            @click="logout"
          >
            退出登录
          </button>
          <p class="px-0.5 text-[10px] leading-relaxed text-slate-400">
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
      class="flex shrink-0 overflow-x-auto border-t border-slate-200/90 bg-white/95 px-1 py-2 backdrop-blur-md md:hidden"
    >
      <RouterLink
        v-for="item in merchantNavItems"
        :key="item.to"
        :to="item.to"
        class="flex min-w-[3.25rem] flex-1 flex-col items-center gap-0.5 rounded-lg py-1.5 text-[9px] font-medium text-slate-500 transition sm:min-w-0 sm:text-[10px]"
        active-class="nav-mb-active"
      >
        <span class="opacity-80" v-html="item.iconSm" />
        <span class="truncate px-0.5">{{ item.short }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, RouterLink, RouterView } from 'vue-router'
import { postMerchantLogout } from '@/api/session'
import { merchantNavItems } from '@/config/merchantMenu'
import { clearMerchantSession, loadMerchantAuth, merchantMonogram, resolveMerchantDisplayName } from '@/lib/merchantApi'
import { loadMerchantDisplaySettings } from '@/lib/displaySettings'
import { useServerClock } from '@/composables/useServerClock'

const router = useRouter()
const mobileAccountDetailsRef = ref<HTMLDetailsElement | null>(null)
const { serverTimeText } = useServerClock()

const merchantIdDisplay = computed(() => {
  try {
    return loadMerchantAuth().merchantId || ''
  } catch {
    return ''
  }
})

const merchantDisplayName = computed(() => resolveMerchantDisplayName(merchantIdDisplay.value))

const monogram = computed(() => merchantMonogram(merchantDisplayName.value, merchantIdDisplay.value))

onMounted(() => {
  void loadMerchantDisplaySettings()
})

async function logout() {
  try {
    await postMerchantLogout()
  } catch {
  } finally {
    clearMerchantSession()
    await router.replace('/login')
  }
}

async function onMobileLogout() {
  if (mobileAccountDetailsRef.value) {
    mobileAccountDetailsRef.value.open = false
  }
  await logout()
}
</script>

<style scoped>
.nav-item-active {
  @apply border border-slate-300/90 bg-gradient-to-r from-slate-100 to-slate-50/90 text-slate-900 shadow-sm;
}
.nav-item-active span:first-child {
  @apply text-slate-700;
}
.nav-mb-active {
  @apply bg-slate-100 text-slate-900;
}
.merchant-menu-scroll {
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.45) transparent;
}
.merchant-menu-scroll::-webkit-scrollbar {
  width: 6px;
}
.merchant-menu-scroll::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgba(148, 163, 184, 0.45);
}
</style>
