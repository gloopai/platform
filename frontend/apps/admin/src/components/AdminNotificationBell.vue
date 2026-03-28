<template>
  <div ref="rootRef" class="relative">
    <button
      type="button"
      class="relative inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
      :title="connected ? '通知已连接' : '通知连接中…'"
      aria-label="通知"
      @click.stop="toggleOpen"
    >
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
      <span
        v-if="pendingCount > 0"
        class="absolute -right-0.5 -top-0.5 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white ring-2 ring-white"
      >
        {{ pendingCount > 99 ? '99+' : pendingCount }}
      </span>
      <span
        v-if="!connected"
        class="absolute bottom-0.5 right-0.5 h-2 w-2 rounded-full bg-amber-400 ring-2 ring-white"
        title="重连中"
      />
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
        v-show="open"
        class="absolute right-0 z-[60] mt-2 w-[min(100vw-2rem,22rem)] origin-top-right overflow-hidden rounded-xl border border-slate-200 bg-white shadow-lg ring-1 ring-black/5"
      >
        <div class="border-b border-slate-100 px-3 py-2">
          <div class="text-xs font-semibold text-slate-900">通知</div>
          <div class="text-[10px] text-slate-500">实时推送；离线时仅保留历史记录于服务端</div>
        </div>
        <div class="max-h-80 overflow-y-auto">
          <div v-if="!recent.length" class="px-3 py-8 text-center text-xs text-slate-500">暂无通知</div>
          <button
            v-for="item in recent"
            :key="item.id"
            type="button"
            class="flex w-full flex-col gap-0.5 border-b border-slate-50 px-3 py-2.5 text-left transition hover:bg-slate-50"
            @click="onPick(item)"
          >
            <div class="text-xs font-semibold text-slate-900">{{ item.title }}</div>
            <div v-if="item.body" class="line-clamp-2 text-[11px] text-slate-600">{{ item.body }}</div>
            <div class="text-[10px] text-slate-400">{{ formatTime(item.at) }}</div>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import type { PortalNotifyListItem } from '../lib/portalNotifyTypes'

const props = defineProps<{
  recent: PortalNotifyListItem[]
  pendingCount: number
  connected: boolean
}>()

const emit = defineEmits<{
  openChange: [open: boolean]
  navigate: [item: PortalNotifyListItem]
}>()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

function toggleOpen() {
  open.value = !open.value
  if (open.value) emit('openChange', true)
}

function close() {
  open.value = false
}

function onPick(item: PortalNotifyListItem) {
  emit('navigate', item)
  close()
}

function formatTime(at: number) {
  const d = new Date(at)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function onDocClick(e: MouseEvent) {
  const el = rootRef.value
  if (!el || !open.value) return
  if (!el.contains(e.target as Node)) close()
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
})
</script>
