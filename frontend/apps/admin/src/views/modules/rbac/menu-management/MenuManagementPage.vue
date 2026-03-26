<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">菜单管理</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">统一维护左侧菜单、头像菜单和未绑定菜单的其它功能。</p>
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
