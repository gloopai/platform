<template>
  <div class="flex min-h-full flex-col bg-slate-50">
    <header
      class="sticky top-0 z-40 border-b border-slate-200/80 bg-white/75 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60"
    >
      <div class="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
        <RouterLink to="/" class="group flex items-center gap-3">
          <span
            class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-indigo-600 text-sm font-bold text-white shadow-lg shadow-brand-500/25"
          >
            G
          </span>
          <span class="flex flex-col leading-tight">
            <span class="text-sm font-bold tracking-tight text-slate-900">{{ t('brand.name') }}</span>
            <span class="text-[10px] font-medium uppercase tracking-[0.2em] text-slate-500">{{
              t('brand.tagline')
            }}</span>
          </span>
        </RouterLink>

        <div class="flex items-center gap-2 sm:gap-6">
          <nav class="hidden items-center gap-1 md:flex">
            <RouterLink
              v-for="link in navLinks"
              :key="link.to"
              :to="link.to"
              class="nav-link"
              active-class="nav-link-active"
            >
              {{ link.label }}
            </RouterLink>
          </nav>
          <LanguageSwitcher />
        </div>
      </div>
      <!-- Mobile nav -->
      <nav class="flex flex-wrap gap-1 border-t border-slate-100 px-4 py-2 md:hidden">
        <RouterLink
          v-for="link in navLinks"
          :key="link.to + 'm'"
          :to="link.to"
          class="rounded-lg px-3 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-100"
          active-class="bg-brand-50 text-brand-800"
        >
          {{ link.label }}
        </RouterLink>
      </nav>
    </header>

    <main class="flex-1">
      <RouterView />
    </main>

    <footer class="mt-auto border-t border-slate-200 bg-slate-900 text-slate-300">
      <div class="mx-auto max-w-6xl px-4 py-12 sm:px-6 lg:px-8">
        <div class="flex flex-col gap-8 md:flex-row md:items-start md:justify-between">
          <div>
            <div class="flex items-center gap-2">
              <span
                class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-400 to-indigo-500 text-xs font-bold text-white"
                >G</span
              >
              <span class="font-semibold text-white">{{ t('brand.name') }}</span>
            </div>
            <p class="mt-3 max-w-sm text-sm text-slate-400">{{ t('meta.description') }}</p>
          </div>
          <div class="flex flex-wrap gap-8 text-sm">
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-slate-500">{{ t('nav.products') }}</div>
              <RouterLink to="/products" class="mt-2 block text-slate-300 hover:text-white">{{ t('nav.products') }}</RouterLink>
            </div>
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-slate-500">{{ t('nav.docs') }}</div>
              <RouterLink to="/docs" class="mt-2 block text-slate-300 hover:text-white">{{ t('nav.docs') }}</RouterLink>
            </div>
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-slate-500">{{ t('nav.about') }}</div>
              <RouterLink to="/about" class="mt-2 block text-slate-300 hover:text-white">{{ t('nav.about') }}</RouterLink>
            </div>
          </div>
        </div>
        <div
          class="mt-10 flex flex-col gap-3 border-t border-slate-800 pt-8 text-xs text-slate-500 sm:flex-row sm:items-center sm:justify-between"
        >
          <span>© {{ year }} {{ t('brand.name') }}. {{ t('footer.rights') }}</span>
          <div class="flex flex-wrap gap-4">
            <span class="cursor-default hover:text-slate-400">{{ t('footer.privacy') }}</span>
            <span class="cursor-default hover:text-slate-400">{{ t('footer.terms') }}</span>
            <span class="cursor-default hover:text-slate-400">{{ t('footer.contact') }}</span>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'

import LanguageSwitcher from './components/LanguageSwitcher.vue'
import { htmlLangFromLocale, persistLocale } from './i18n'

const { t, locale } = useI18n()

const year = new Date().getFullYear()

const navLinks = computed(() => [
  { to: '/', label: t('nav.home') },
  { to: '/products', label: t('nav.products') },
  { to: '/docs', label: t('nav.docs') },
  { to: '/about', label: t('nav.about') },
])

watch(
  locale,
  (v) => {
    document.title = t('meta.title')
    document.documentElement.lang = htmlLangFromLocale(v as string)
    persistLocale(v as string)
  },
  { immediate: true },
)
</script>

<style scoped>
.nav-link {
  @apply rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-100 hover:text-slate-900;
}
.nav-link-active {
  @apply bg-brand-50 text-brand-800;
}
</style>
