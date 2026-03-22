<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">{{ title }}</h1>
      <p class="mt-1 text-sm text-slate-600">{{ meta?.summary ?? '功能建设中。' }}</p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">规划能力</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li v-for="(b, i) in meta?.bullets ?? []" :key="i">{{ b }}</li>
      </ul>
      <div
        v-if="meta?.apiNote"
        class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 font-mono text-xs text-amber-900"
      >
        {{ meta.apiNote }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { adminPathTitle, adminPlaceholderMeta } from '../../adminMenu'

const route = useRoute()

const title = computed(() => adminPathTitle[route.path] ?? '功能模块')

const meta = computed(() => adminPlaceholderMeta[route.path] ?? null)
</script>
