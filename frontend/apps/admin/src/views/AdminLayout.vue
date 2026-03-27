<template>
  <div class="admin-shell flex h-screen w-full flex-col overflow-hidden bg-slate-100">
    <div class="flex min-h-0 min-w-0 flex-1">
      <aside
        :class="[
          'relative z-10 flex shrink-0 flex-col border-r border-slate-800/80 bg-gradient-to-b from-slate-950 via-slate-950 to-slate-900 text-slate-400 shadow-[4px_0_24px_-8px_rgba(15,23,42,0.35)] transition-[width] duration-200 ease-out',
          sidebarCollapsed ? 'w-[72px]' : 'w-60',
        ]"
      >
        <!-- 品牌区：与深色侧栏一体，避免顶栏通栏割裂 -->
        <div class="shrink-0 border-b border-white/10">
          <div
            class="flex items-center gap-2 px-3 py-3"
            :class="sidebarCollapsed ? 'flex-col justify-center gap-3' : 'justify-between'"
          >
            <div
              class="flex min-w-0 items-center gap-2.5"
              :class="sidebarCollapsed ? 'flex-col gap-2' : ''"
            >
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 via-violet-600 to-fuchsia-500 text-sm font-bold text-white shadow-lg shadow-indigo-900/40 ring-1 ring-white/10"
                aria-hidden="true"
              >
                P
              </div>
              <div v-show="!sidebarCollapsed" class="min-w-0">
                <div class="truncate text-sm font-semibold tracking-tight text-white">聚合支付</div>
                <div class="truncate text-[10px] font-medium text-slate-500">总管理台</div>
              </div>
            </div>
            <button
              type="button"
              class="shrink-0 rounded-lg p-1.5 text-slate-400 transition hover:bg-white/10 hover:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500/40"
              :title="sidebarCollapsed ? '展开菜单' : '收起菜单'"
              @click="toggleSidebar"
            >
              <svg
                class="h-5 w-5 transition duration-200"
                :class="sidebarCollapsed ? '' : 'rotate-180'"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        </div>

        <nav ref="navScrollRef" class="admin-nav-scroll flex-1 space-y-1 overflow-y-auto px-2 py-3">
          <template v-for="entry in menu" :key="entry.kind === 'leaf' ? entry.to : entry.key">
            <!-- 单页 -->
            <div
              v-if="entry.kind === 'leaf'"
              class="relative"
              @mouseenter="onLeafEnter(entry.to)"
              @mouseleave="onLeafLeave"
            >
              <RouterLink
                :to="entry.to"
                class="group flex items-center rounded-lg py-2.5 text-sm font-medium transition"
                :class="sidebarCollapsed ? 'justify-center px-2' : 'gap-3 px-3'"
                active-class="nav-active"
              >
                <span class="nav-icon shrink-0 text-slate-500 group-hover:text-slate-300" v-html="icons[entry.icon]" />
                <span v-show="!sidebarCollapsed" class="truncate">{{ entry.label }}</span>
              </RouterLink>
              <div
                v-show="sidebarCollapsed && leafTooltip === entry.to"
                class="nav-flyout absolute left-full top-1/2 z-[100] -translate-y-1/2 pl-2"
              >
                <div
                  class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-xs font-semibold text-white shadow-xl ring-1 ring-black/40"
                >
                  {{ entry.label }}
                </div>
              </div>
            </div>

            <!-- 分组 -->
            <div v-else class="pt-1">
              <div v-if="!sidebarCollapsed" class="space-y-1">
                <button
                  type="button"
                  class="flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-left text-sm font-medium text-slate-300 transition hover:bg-white/5"
                  @click="toggleGroup(entry.key)"
                >
                  <span class="flex min-w-0 items-center gap-3">
                    <span class="nav-icon shrink-0 text-slate-500" v-html="icons[entry.icon]" />
                    <span class="truncate">{{ entry.label }}</span>
                  </span>
                  <svg
                    class="h-4 w-4 shrink-0 text-slate-500 transition"
                    :class="openGroups[entry.key] ? 'rotate-180' : ''"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path d="M6 9l6 6 6-6" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
                <div v-show="openGroups[entry.key]" class="ml-6 space-y-0.5 border-l border-white/10 pl-4">
                  <RouterLink
                    v-for="child in entry.children"
                    :key="child.to"
                    :to="child.to"
                    class="group flex items-center gap-2 rounded-md py-2 pl-2 pr-2 text-sm font-medium text-slate-400 transition hover:bg-white/5 hover:text-slate-200"
                    active-class="nav-sub-active"
                  >
                    <span class="h-1.5 w-1.5 shrink-0 rounded-full bg-slate-600 group-hover:bg-slate-400" />
                    <span class="truncate">{{ child.label }}</span>
                  </RouterLink>
                </div>
              </div>

              <div
                v-else
                :ref="(el) => setGroupTriggerRef(entry.key, el)"
                class="relative"
                @mouseenter="() => onGroupFlyoutEnter(entry.key)"
                @mouseleave="onGroupFlyoutLeave"
              >
                <button
                  type="button"
                  class="flex w-full items-center justify-center rounded-lg px-2 py-2.5 text-sm font-medium transition hover:bg-white/5"
                  :class="isGroupRouteActive(entry) ? 'nav-active' : 'text-slate-300'"
                >
                  <span
                    class="nav-icon shrink-0 text-slate-500"
                    :class="isGroupRouteActive(entry) ? '!text-indigo-300' : ''"
                    v-html="icons[entry.icon]"
                  />
                </button>
              </div>
            </div>
          </template>
        </nav>

        <div v-show="!sidebarCollapsed" class="border-t border-white/5 p-3">
          <div class="rounded-lg bg-white/5 px-3 py-2 text-[10px] leading-relaxed text-slate-500">
            聚合多通道生产环境：变更路由与资金类操作前请二次确认。
          </div>
        </div>
      </aside>

      <div class="relative z-20 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
        <header
          class="flex h-[52px] shrink-0 items-center justify-between gap-4 border-b border-slate-200/90 bg-white px-4 shadow-sm sm:px-6"
        >
          <div class="flex min-w-0 flex-1 items-center">
            <div class="min-w-0">
              <div class="text-[10px] font-medium uppercase tracking-[0.18em] text-slate-400">当前页面</div>
              <div class="truncate text-sm font-semibold text-slate-900 sm:text-base">{{ pageTitle }}</div>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-2">
            <div class="hidden rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1 text-xs font-mono text-slate-600 lg:block">
               {{ serverTimeText }}
            </div>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              @click="broadcastRefresh"
            >
              <svg class="h-3.5 w-3.5 text-slate-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
              刷新数据
            </button>

            <div ref="userMenuRoot" class="relative">
              <button
                type="button"
                class="flex items-center gap-2 rounded-xl  py-1.5 pl-1.5 pr-2.5 text-left transition"
                @click.stop="userMenuOpen = !userMenuOpen"
              >
                <span
                  class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-slate-700 to-slate-900 text-xs font-bold text-white"
                >
                  {{ userInitial }}
                </span>
                <div class="hidden text-left leading-tight sm:block">
                  <div class="text-xs font-semibold text-slate-900">{{ displayName }}</div>
                  <div class="text-[10px] text-slate-500">{{ adminRoleLabel }}</div>
                </div>
                <svg
                  class="h-4 w-4 text-slate-400 transition"
                  :class="userMenuOpen ? 'rotate-180' : ''"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path d="M6 9l6 6 6-6" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </button>

              <Transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
              >
                <div
                  v-show="userMenuOpen"
                  class="absolute right-0 z-50 mt-2 w-52 origin-top-right rounded-xl border border-slate-200 bg-white py-1 shadow-lg ring-1 ring-black/5"
                >
                  <div class="border-b border-slate-100 px-3 py-2">
                    <div class="text-xs font-semibold text-slate-900">{{ displayName }}</div>
                    <div class="text-[10px] text-slate-500">已登录 · 总管理台</div>
                  </div>
                  <RouterLink
                    v-for="a in avatarLinks"
                    :key="a.to"
                    :to="a.to"
                    class="flex w-full items-center gap-2 border-b border-slate-100 px-3 py-2 text-left text-xs font-medium text-slate-700 hover:bg-slate-50"
                    @click="userMenuOpen = false"
                  >
                    <span
                      v-if="icons[a.icon]"
                      class="shrink-0 text-slate-400 [&_svg]:h-4 [&_svg]:w-4"
                      v-html="icons[a.icon]"
                    />
                    <span class="truncate">{{ a.label }}</span>
                  </RouterLink>
                  <button
                    type="button"
                    class="flex w-full items-center gap-2 px-3 py-3 text-left text-xs font-medium text-slate-700 hover:bg-slate-50"
                    @click="logout"
                  >
                    <svg class="h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                    退出登录
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </header>

        <main class="min-h-0 flex-1 overflow-y-auto bg-slate-100/80">
          <div class="min-h-full w-full p-4 sm:p-6 lg:p-8">
            <RouterView />
          </div>
        </main>
      </div>
    </div>
  </div>

  <UiToastHost />
  <UiDialogHost />

  <Teleport to="body">
    <div
      v-show="flyoutKey && sidebarCollapsed"
      class="fixed z-[300] min-w-[208px] overflow-hidden rounded-xl border border-slate-700 bg-slate-900 py-2 shadow-2xl ring-1 ring-black/50"
      :style="{ top: `${flyoutPos.top}px`, left: `${flyoutPos.left}px` }"
      @mouseenter="onFlyoutPanelEnter"
      @mouseleave="onFlyoutPanelLeave"
    >
      <template v-if="flyoutGroup">
        <div class="border-b border-white/10 px-3 pb-2 pt-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
          {{ flyoutGroup.label }}
        </div>
        <div class="mt-1 space-y-0.5 px-1">
          <RouterLink
            v-for="child in flyoutGroup.children"
            :key="child.to"
            :to="child.to"
            class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-white/10 hover:text-white"
            active-class="nav-flyout-active"
            @click="closeFlyout"
          >
            <span class="h-1.5 w-1.5 shrink-0 rounded-full bg-slate-500" />
            <span class="truncate">{{ child.label }}</span>
          </RouterLink>
        </div>
      </template>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, provide, reactive, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter, RouterLink, RouterView } from 'vue-router'
import UiDialogHost from '../components/UiDialogHost.vue'
import UiToastHost from '../components/UiToastHost.vue'
import { adminPathTitle, findGroupKeyForPath, pathBelongsToGroup, type AdminMenuEntry, type AdminMenuGroup } from '../adminMenu'
import { adminGet, adminPost, clearAdminSession, loadAdminIdentity, loadAdminToken, saveAdminIdentity } from '../lib/adminApi'
import { loadAdminDisplaySettings } from '../lib/displaySettings'
import { useServerClock } from '../composables/useServerClock'
import { useUiDialog } from '../composables/useUiDialog'

type RefreshFn = () => void

const SIDEBAR_KEY = 'admin_sidebar_collapsed'

const router = useRouter()
const route = useRoute()
const dialog = useUiDialog()
const adminToken = ref(loadAdminToken())
const adminIdentity = ref(loadAdminIdentity())
const adminRoleLabel = ref('管理员')
const { serverTimeText } = useServerClock()

const menu = ref<AdminMenuEntry[]>([])

type AvatarLink = { to: string; label: string; icon: string }
const avatarLinks = ref<AvatarLink[]>([])

const refreshFns = ref<RefreshFn[]>([])
provide('adminToken', adminToken)
provide('registerRefresh', (fn: RefreshFn) => {
  refreshFns.value.push(fn)
  return () => {
    refreshFns.value = refreshFns.value.filter((x) => x !== fn)
  }
})

const userMenuOpen = ref(false)
const userMenuRoot = ref<HTMLElement | null>(null)

const sidebarCollapsed = ref(typeof localStorage !== 'undefined' && localStorage.getItem(SIDEBAR_KEY) === '1')

const openGroups = reactive<Record<string, boolean>>({})

function syncOpenGroups() {
  for (const k of Object.keys(openGroups)) delete openGroups[k]
  for (const e of menu.value) {
    if (e.kind === 'group') openGroups[e.key] = true
  }
}
syncOpenGroups()

const flyoutKey = ref<string | null>(null)
const leafTooltip = ref<string | null>(null)

const navScrollRef = ref<HTMLElement | null>(null)
const groupTriggerRefs = ref<Record<string, HTMLElement | null>>({})
const flyoutPos = ref({ top: 0, left: 0 })

let leafTimer: ReturnType<typeof setTimeout> | null = null
let groupTimer: ReturnType<typeof setTimeout> | null = null
const HOVER_MS = 140

function setGroupTriggerRef(key: string, el: unknown) {
  if (!el) {
    groupTriggerRefs.value[key] = null
    return
  }
  const dom = el instanceof HTMLElement ? el : null
  groupTriggerRefs.value[key] = dom
}

function clearLeafTimer() {
  if (leafTimer) {
    clearTimeout(leafTimer)
    leafTimer = null
  }
}

function clearGroupTimer() {
  if (groupTimer) {
    clearTimeout(groupTimer)
    groupTimer = null
  }
}

function updateFlyoutPosition() {
  const k = flyoutKey.value
  if (!k) return
  const el = groupTriggerRefs.value[k]
  if (!el) return
  const r = el.getBoundingClientRect()
  flyoutPos.value = { top: r.top, left: r.right + 8 }
}

function onGroupFlyoutEnter(key: string) {
  clearGroupTimer()
  clearLeafTimer()
  leafTooltip.value = null
  flyoutKey.value = key
  nextTick(() => updateFlyoutPosition())
}

function onGroupFlyoutLeave() {
  clearGroupTimer()
  groupTimer = setTimeout(() => {
    flyoutKey.value = null
    groupTimer = null
  }, HOVER_MS)
}

function onFlyoutPanelEnter() {
  clearGroupTimer()
}

function onFlyoutPanelLeave() {
  onGroupFlyoutLeave()
}

function closeFlyout() {
  clearGroupTimer()
  flyoutKey.value = null
}

function toggleGroup(key: string) {
  openGroups[key] = !openGroups[key]
}

function isGroupRouteActive(entry: AdminMenuGroup): boolean {
  return pathBelongsToGroup(route.path, entry)
}

function onLeafEnter(path: string) {
  if (!sidebarCollapsed.value) return
  clearLeafTimer()
  clearGroupTimer()
  flyoutKey.value = null
  leafTooltip.value = path
}

function onLeafLeave() {
  clearLeafTimer()
  leafTimer = setTimeout(() => {
    leafTooltip.value = null
    leafTimer = null
  }, HOVER_MS)
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  try {
    localStorage.setItem(SIDEBAR_KEY, sidebarCollapsed.value ? '1' : '0')
  } catch {
  }
  clearLeafTimer()
  clearGroupTimer()
  flyoutKey.value = null
  leafTooltip.value = null
}

const icons: Record<string, string> = {
  chart: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>`,
  briefcase: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M20 7h-4V5a2 2 0 00-2-2h-4a2 2 0 00-2 2v2H4a2 2 0 00-2 2v11a2 2 0 002 2h16a2 2 0 002-2V9a2 2 0 00-2-2zm-6 0h-4V5h4v2z" /></svg>`,
  layers: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 7l8-4 8 4M4 7v10l8 4 8-4V7M4 7l8 4 8-4M12 11v10" /></svg>`,
  credit: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>`,
  shield: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>`,
  cog: `<svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>`,
}

const pageTitle = computed(() => adminPathTitle[route.path] ?? '管理台')

const flyoutGroup = computed((): AdminMenuGroup | null => {
  const k = flyoutKey.value
  if (!k) return null
  const e = menu.value.find((x) => x.kind === 'group' && x.key === k)
  return e && e.kind === 'group' ? e : null
})

const displayName = computed(() => adminIdentity.value || '管理员')
const userInitial = computed(() => displayName.value.slice(0, 1).toUpperCase())

function broadcastRefresh() {
  for (const fn of refreshFns.value) fn()
}

async function logout() {
  const ok = await dialog.confirm('确认退出当前登录账号？', '退出登录')
  if (!ok) return
  userMenuOpen.value = false
  try {
    await adminPost<{ ok: boolean }>('/v1/admin/logout')
  } catch {
  } finally {
    clearAdminSession()
    await router.replace('/login')
  }
}

function onDocClick(e: MouseEvent) {
  const el = userMenuRoot.value
  if (!el) return
  if (!el.contains(e.target as Node)) userMenuOpen.value = false
}

watch(
  () => route.path,
  (p) => {
    const gk = findGroupKeyForPath(p, menu.value)
    if (gk) openGroups[gk] = true
  },
  { immediate: true },
)

watch(sidebarCollapsed, (c) => {
  if (!c) {
    flyoutKey.value = null
    leafTooltip.value = null
  } else if (flyoutKey.value) {
    nextTick(() => updateFlyoutPosition())
  }
})

watch(flyoutKey, (v) => {
  if (v) nextTick(() => updateFlyoutPosition())
})

watchEffect((onCleanup) => {
  const nav = navScrollRef.value
  if (!nav) return
  const onScroll = () => {
    if (flyoutKey.value) updateFlyoutPosition()
  }
  nav.addEventListener('scroll', onScroll, { passive: true })
  onCleanup(() => nav.removeEventListener('scroll', onScroll))
})

function onWindowResizeOrScroll() {
  if (flyoutKey.value) updateFlyoutPosition()
}

onMounted(() => {
  void loadMe()
  void loadAdminDisplaySettings()
  void loadMenu()
  document.addEventListener('click', onDocClick)
  window.addEventListener('resize', onWindowResizeOrScroll)
})
onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  window.removeEventListener('resize', onWindowResizeOrScroll)
})
onBeforeUnmount(() => {
  clearLeafTimer()
  clearGroupTimer()
})

function allowedPathsFromMenu(entries: AdminMenuEntry[], extra: { to: string }[]): string[] {
  const out: string[] = []
  for (const e of entries) {
    if (e.kind === 'leaf') out.push(e.to)
    else out.push(...e.children.map((c) => c.to))
  }
  for (const x of extra) {
    const p = (x.to || '').trim()
    if (p && !out.includes(p)) out.push(p)
  }
  return out
}

async function loadMenu() {
  try {
    const raw = await adminGet<unknown>('/v1/admin/rbac/my_menu')
    let sidebar: AdminMenuEntry[] | null = null
    let av: AvatarLink[] = []
    if (Array.isArray(raw)) {
      sidebar = raw as AdminMenuEntry[]
    } else if (raw && typeof raw === 'object' && raw !== null && 'sidebar' in raw) {
      const o = raw as { sidebar?: AdminMenuEntry[]; avatar_links?: AvatarLink[] }
      sidebar = o.sidebar ?? null
      av = Array.isArray(o.avatar_links) ? o.avatar_links : []
    }
    menu.value = Array.isArray(sidebar) ? sidebar : []
    avatarLinks.value = av
    syncOpenGroups()
    const allowed = allowedPathsFromMenu(menu.value, avatarLinks.value)
    try {
      localStorage.setItem('admin_allowed_paths', JSON.stringify(allowed))
    } catch {
    }
    if (allowed.length && !allowed.includes(route.path)) {
      await router.replace(allowed[0])
    }
    if (!allowed.length && route.path !== '/login') {
      // No authorized menu path for this account: keep UI shell but hide sidebar entries.
      await router.replace('/')
    }
  } catch {
    // Fail-safe: never fallback to full static menu when permission fetch fails.
    menu.value = []
    avatarLinks.value = []
    syncOpenGroups()
    try {
      localStorage.setItem('admin_allowed_paths', JSON.stringify([]))
    } catch {
    }
  }
}

async function loadMe() {
  try {
    const me = await adminGet<{ id: number; username: string; email: string; display_name: string; role: string }>('/v1/admin/me')
    const next = (me.display_name || me.email || me.username || '').trim()
    if (next) {
      adminIdentity.value = next
      saveAdminIdentity(next)
    }
    const role = (me.role || '').trim()
    if (role) adminRoleLabel.value = role
  } catch {
    // keep local fallback identity
  }
}
</script>

<style scoped>
.admin-nav-scroll {
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.35) transparent;
}
.admin-nav-scroll::-webkit-scrollbar {
  width: 6px;
}
.admin-nav-scroll::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgba(148, 163, 184, 0.35);
}
.nav-active {
  @apply bg-indigo-500/15 text-white shadow-inner shadow-indigo-500/10;
}
.nav-active .nav-icon {
  @apply text-indigo-300;
}
.nav-sub-active {
  @apply bg-white/10 text-white;
}
.nav-sub-active .rounded-full {
  @apply bg-indigo-400;
}
.nav-flyout-active {
  @apply bg-indigo-500/20 text-white;
}
.nav-flyout-active .rounded-full {
  @apply bg-indigo-400;
}
</style>
