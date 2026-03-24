<template>
  <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6 lg:px-8 lg:py-20">
    <header class="max-w-2xl">
      <h1 class="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">{{ t('about.title') }}</h1>
      <p class="mt-4 text-base leading-relaxed text-slate-600">{{ t('about.lead') }}</p>
    </header>

    <div class="mt-12 grid gap-6 lg:grid-cols-2">
      <article class="rounded-3xl border border-slate-200 bg-white p-8 shadow-sm lg:p-10">
        <h2 class="text-lg font-semibold text-slate-900">{{ t('about.company.title') }}</h2>
        <p class="mt-3 text-sm leading-relaxed text-slate-600">{{ t('about.company.desc') }}</p>
      </article>
      <article class="rounded-3xl border border-slate-200 bg-gradient-to-br from-slate-900 to-slate-800 p-8 text-white shadow-lg lg:p-10">
        <h2 class="text-lg font-semibold">{{ t('about.security.title') }}</h2>
        <p class="mt-3 text-sm leading-relaxed text-slate-300">{{ t('about.security.desc') }}</p>
        <div class="mt-6 flex flex-wrap gap-2">
          <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-semibold backdrop-blur-sm">{{ t('about.security.pci') }}</span>
          <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-semibold backdrop-blur-sm">{{ t('about.security.privacy') }}</span>
          <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-semibold backdrop-blur-sm">{{ t('about.security.terms') }}</span>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
