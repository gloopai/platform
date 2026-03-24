<template>
  <div ref="rootEl" class="relative">
    <button
      type="button"
      class="lang-trigger"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="open = !open"
    >
      <svg class="h-4 w-4 opacity-80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <circle cx="12" cy="12" r="10" />
        <path d="M2 12h20M12 2a15 15 0 0 1 0 20M12 2a15 15 0 0 0 0 20" />
      </svg>
      <span>{{ currentLabel }}</span>
      <svg class="h-3.5 w-3.5 opacity-60" viewBox="0 0 24 24" fill="currentColor">
        <path d="M7 10l5 5 5-5z" />
      </svg>
    </button>
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <ul
        v-show="open"
        class="lang-menu"
        role="listbox"
        @keydown.escape.prevent="open = false"
      >
        <li v-for="opt in options" :key="opt.value">
          <button
            type="button"
            role="option"
            class="lang-option"
            :class="{ active: locale === opt.value }"
            @click="select(opt.value)"
          >
            {{ opt.label }}
          </button>
        </li>
      </ul>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import type { PortalLocale } from '../i18n'

const { locale } = useI18n()
const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)

const options: { value: PortalLocale; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'zh-CN', label: '简体中文' },
  { value: 'zh-TW', label: '繁體中文' },
  { value: 'ja', label: '日本語' },
  { value: 'pt-BR', label: 'Português (Brasil)' },
  { value: 'hi-IN', label: 'हिन्दी' },
]

const currentLabel = computed(() => options.find((o) => o.value === locale.value)?.label ?? 'English')

function select(value: PortalLocale) {
  locale.value = value
  open.value = false
}

function onDocClick(e: MouseEvent) {
  if (rootEl.value && !rootEl.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
})
onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
})
</script>

<style scoped>
.lang-trigger {
  @apply inline-flex items-center gap-2 rounded-full border border-slate-200/80 bg-white/80 px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-sm backdrop-blur-sm transition hover:border-slate-300 hover:bg-white;
}
.lang-menu {
  @apply absolute right-0 z-50 mt-2 max-h-[min(22rem,70vh)] min-w-[11rem] overflow-y-auto overflow-x-hidden rounded-xl border border-slate-200/90 bg-white py-1 shadow-xl shadow-slate-900/10;
}
.lang-option {
  @apply block w-full px-4 py-2.5 text-left text-sm text-slate-700 transition hover:bg-slate-50;
}
.lang-option.active {
  @apply bg-brand-50 font-semibold text-brand-700;
}
</style>
