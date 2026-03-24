<template>
  <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6 lg:px-8 lg:py-20">
    <header class="max-w-2xl">
      <h1 class="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">{{ t('products.title') }}</h1>
      <p class="mt-4 text-base leading-relaxed text-slate-600">{{ t('products.lead') }}</p>
    </header>

    <div class="mt-12 grid gap-6 md:grid-cols-3">
      <article
        v-for="(s, i) in scenarios"
        :key="i"
        class="rounded-2xl border border-slate-200 bg-white p-8 shadow-sm transition hover:border-brand-100 hover:shadow-md"
      >
        <div class="text-xs font-bold uppercase tracking-wider text-brand-600">{{ s.kicker }}</div>
        <h2 class="mt-3 text-lg font-semibold text-slate-900">{{ s.title }}</h2>
        <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ s.desc }}</p>
      </article>
    </div>

    <section class="mt-16 rounded-3xl border border-slate-200 bg-gradient-to-b from-white to-slate-50 p-8 sm:p-10">
      <h2 class="text-xl font-bold text-slate-900">{{ t('products.settlement.title') }}</h2>
      <p class="mt-2 max-w-3xl text-sm leading-relaxed text-slate-600">{{ t('products.settlement.lead') }}</p>
      <div class="mt-8 grid gap-4 md:grid-cols-3">
        <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
          <div class="font-mono text-sm font-bold text-brand-700">{{ t('products.settlement.d0.label') }}</div>
          <p class="mt-2 text-sm text-slate-600">{{ t('products.settlement.d0.desc') }}</p>
        </div>
        <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
          <div class="font-mono text-sm font-bold text-brand-700">{{ t('products.settlement.t1.label') }}</div>
          <p class="mt-2 text-sm text-slate-600">{{ t('products.settlement.t1.desc') }}</p>
        </div>
        <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
          <div class="font-mono text-sm font-bold text-brand-700">{{ t('products.settlement.recon.label') }}</div>
          <p class="mt-2 text-sm text-slate-600">{{ t('products.settlement.recon.desc') }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const scenarios = computed(() => [
  { kicker: '01', title: t('products.scenarios.ecommerce.title'), desc: t('products.scenarios.ecommerce.desc') },
  { kicker: '02', title: t('products.scenarios.gaming.title'), desc: t('products.scenarios.gaming.desc') },
  { kicker: '03', title: t('products.scenarios.retail.title'), desc: t('products.scenarios.retail.desc') },
])
</script>
