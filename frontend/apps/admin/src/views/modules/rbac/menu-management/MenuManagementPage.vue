<template>
  <div class="space-y-6">
    <div class="rounded-2xl border border-slate-200/90 bg-gradient-to-br from-slate-50 to-white p-5 shadow-sm">
      <h2 class="text-sm font-semibold text-slate-900">菜单管理</h2>
      <p class="mt-2 max-w-3xl text-sm leading-relaxed text-slate-600">
        统一维护侧栏、头像菜单与未挂菜单的能力：侧栏支持分组；头像下为一级页面；其它功能用于仅有接口/权限、无导航入口的能力。
      </p>
    </div>

    <PlacementTabs v-model="tab" />

    <LeftSidebarMenuSection v-if="tab === 'left'" />
    <AvatarMenuSection v-else-if="tab === 'avatar'" />
    <OrphanCapabilitiesSection v-else />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AvatarMenuSection from './AvatarMenuSection.vue'
import LeftSidebarMenuSection from './LeftSidebarMenuSection.vue'
import OrphanCapabilitiesSection from './OrphanCapabilitiesSection.vue'
import PlacementTabs from './PlacementTabs.vue'
import type { MenuMgmtTab } from './types'

const route = useRoute()
const router = useRouter()

const tab = computed<MenuMgmtTab>({
  get() {
    const q = route.query.tab
    const s = typeof q === 'string' ? q : Array.isArray(q) ? q[0] : ''
    if (s === 'avatar' || s === 'other') return s
    return 'left'
  },
  set(v: MenuMgmtTab) {
    if (v === 'left') {
      const { tab: _t, ...rest } = route.query
      router.replace({ path: route.path, query: rest })
    } else {
      router.replace({ path: route.path, query: { ...route.query, tab: v } })
    }
  },
})
</script>
