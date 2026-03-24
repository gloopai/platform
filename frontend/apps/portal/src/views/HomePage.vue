<template>
  <div>
    <!-- Hero -->
    <section class="relative overflow-hidden border-b border-slate-200/80 bg-slate-950">
      <div
        class="pointer-events-none absolute inset-0 bg-[size:24px_24px] bg-grid-slate opacity-[0.35]"
        style="mask-image: radial-gradient(ellipse 80% 60% at 50% 0%, black, transparent)"
      />
      <div class="pointer-events-none absolute -left-40 top-20 h-72 w-72 rounded-full bg-brand-500/20 blur-3xl" />
      <div class="pointer-events-none absolute -right-20 top-40 h-64 w-64 rounded-full bg-indigo-500/20 blur-3xl" />

      <div class="relative mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:px-8 lg:py-28">
        <p
          class="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-semibold uppercase tracking-widest text-brand-200"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 shadow-[0_0_8px_#34d399]" />
          {{ t('home.badge') }}
        </p>
        <h1
          class="mt-6 max-w-3xl text-4xl font-bold tracking-tight text-white sm:text-5xl lg:text-[3.25rem] lg:leading-[1.1]"
        >
          {{ t('home.heroTitle') }}
        </h1>
        <p class="mt-6 max-w-2xl text-lg leading-relaxed text-slate-300">
          {{ t('home.heroLead') }}
        </p>
        <div class="mt-10 flex flex-wrap gap-4">
          <RouterLink
            to="/docs"
            class="inline-flex items-center justify-center rounded-xl bg-white px-6 py-3 text-sm font-semibold text-slate-900 shadow-glow transition hover:bg-slate-100"
          >
            {{ t('home.ctaPrimary') }}
          </RouterLink>
          <RouterLink
            to="/products"
            class="inline-flex items-center justify-center rounded-xl border border-white/20 bg-white/5 px-6 py-3 text-sm font-semibold text-white backdrop-blur-sm transition hover:bg-white/10"
          >
            {{ t('home.ctaSecondary') }}
          </RouterLink>
        </div>

        <dl
          class="mt-16 grid gap-6 border-t border-white/10 pt-10 sm:grid-cols-3"
        >
          <div v-for="(stat, i) in stats" :key="i" class="text-left">
            <dt class="text-xs font-semibold uppercase tracking-wider text-slate-500">{{ stat.label }}</dt>
            <dd class="mt-1 font-mono text-2xl font-semibold tabular-nums text-white">{{ stat.value }}</dd>
          </div>
        </dl>
      </div>
    </section>

    <!-- Features -->
    <section class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-24">
      <div class="grid gap-6 md:grid-cols-3">
        <article
          v-for="(card, i) in featureCards"
          :key="i"
          class="group rounded-2xl border border-slate-200/90 bg-white p-8 shadow-sm transition hover:border-brand-200 hover:shadow-md"
        >
          <div
            class="flex h-9 w-9 items-center justify-center rounded-xl bg-brand-50 text-brand-700 transition group-hover:bg-brand-100"
            v-html="card.icon"
          />
          <h2 class="mt-5 text-lg font-semibold text-slate-900">{{ card.title }}</h2>
          <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ card.desc }}</p>
        </article>
      </div>
    </section>

    <!-- Regional coverage -->
    <section class="border-y border-slate-200 bg-gradient-to-b from-slate-50 to-white">
      <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
        <div class="mx-auto max-w-2xl text-center">
          <h2 class="text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl">{{ t('home.regions.title') }}</h2>
          <p class="mt-3 text-sm leading-relaxed text-slate-600 sm:text-base">{{ t('home.regions.lead') }}</p>
        </div>
        <div class="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <article
            v-for="key in regionKeys"
            :key="key"
            class="flex flex-col rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm transition hover:border-brand-200 hover:shadow-md"
          >
            <div class="flex items-center gap-3">
              <span class="text-2xl leading-none" aria-hidden="true">{{ regionFlags[key] }}</span>
              <h3 class="text-lg font-semibold text-slate-900">{{ t(`home.regions.countries.${key}.name`) }}</h3>
            </div>
            <div class="mt-4">
              <div class="text-[11px] font-bold uppercase tracking-wider text-brand-700">
                {{ t(`home.regions.countries.${key}.paymentsLabel`) }}
              </div>
              <p class="mt-1.5 text-sm leading-relaxed text-slate-600">{{ t(`home.regions.countries.${key}.payments`) }}</p>
            </div>
            <div class="mt-4 border-t border-slate-100 pt-4">
              <div class="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                {{ t(`home.regions.countries.${key}.banksLabel`) }}
              </div>
              <p class="mt-1.5 text-sm leading-relaxed text-slate-600">{{ t(`home.regions.countries.${key}.banks`) }}</p>
            </div>
          </article>
        </div>
      </div>
    </section>

    <!-- Cryptocurrency -->
    <section class="relative border-y border-violet-200/60 bg-gradient-to-b from-violet-950/[0.03] via-slate-50 to-white">
      <div
        class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-violet-400/40 to-transparent"
      />
      <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="max-w-2xl">
            <p
              class="inline-flex items-center gap-2 rounded-full border border-violet-200/80 bg-violet-50/90 px-3 py-1 text-[11px] font-bold uppercase tracking-widest text-violet-800"
            >
              <span class="font-mono text-sm leading-none text-violet-600" aria-hidden="true">₿</span>
              Web3
            </p>
            <h2 class="mt-4 text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl">{{ t('home.crypto.title') }}</h2>
            <p class="mt-3 text-sm leading-relaxed text-slate-600 sm:text-base">{{ t('home.crypto.lead') }}</p>
          </div>
        </div>
        <div class="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <article
            v-for="key in cryptoKeys"
            :key="key"
            class="rounded-2xl border border-violet-100/90 bg-white p-6 shadow-sm transition hover:border-violet-200 hover:shadow-md"
          >
            <h3 class="text-base font-semibold text-slate-900">{{ t(`home.crypto.items.${key}.title`) }}</h3>
            <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ t(`home.crypto.items.${key}.desc`) }}</p>
          </article>
        </div>
        <p class="mt-8 text-xs leading-relaxed text-slate-500">{{ t('home.crypto.disclaimer') }}</p>
      </div>
    </section>

    <!-- Partners strip -->
    <section class="border-y border-slate-200 bg-white">
      <div class="mx-auto max-w-6xl px-4 py-12 sm:px-6 lg:px-8">
        <p class="text-center text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
          {{ t('home.partners') }}
        </p>
        <div class="mt-8 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div
            v-for="name in partners"
            :key="name"
            class="flex items-center justify-center rounded-2xl border border-slate-100 bg-slate-50/80 px-4 py-4 text-sm font-semibold text-slate-600 transition hover:border-slate-200 hover:bg-white"
          >
            {{ name }}
          </div>
        </div>
      </div>
    </section>

    <!-- Trust CTA -->
    <section class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
      <div
        class="relative overflow-hidden rounded-3xl border border-slate-200 bg-gradient-to-br from-slate-900 via-slate-800 to-brand-900 p-10 text-white shadow-xl"
      >
        <div class="pointer-events-none absolute right-0 top-0 h-64 w-64 -translate-y-1/2 translate-x-1/3 rounded-full bg-brand-400/30 blur-3xl" />
        <h2 class="relative text-2xl font-bold tracking-tight">{{ t('home.trust.title') }}</h2>
        <p class="relative mt-3 max-w-xl text-sm leading-relaxed text-slate-300">
          {{ t('home.trust.desc') }}
        </p>
        <RouterLink
          to="/docs"
          class="relative mt-8 inline-flex rounded-xl bg-white px-5 py-2.5 text-sm font-semibold text-slate-900 transition hover:bg-slate-100"
        >
          {{ t('home.ctaPrimary') }}
        </RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const icons = {
  bolt: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>`,
  flow: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h10M4 18h16"/></svg>`,
  route: `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l-4.553 2.276A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L15 7m0 0V4m0 3l4.553 2.276A1 1 0 0121 8.382v10.236a1 1 0 01-1.447.894L15 7"/></svg>`,
}

const featureCards = computed(() => [
  { title: t('home.features.collect.title'), desc: t('home.features.collect.desc'), icon: icons.bolt },
  { title: t('home.features.payout.title'), desc: t('home.features.payout.desc'), icon: icons.flow },
  { title: t('home.features.routing.title'), desc: t('home.features.routing.desc'), icon: icons.route },
])

const stats = computed(() => [
  { label: t('home.stats.uptime'), value: t('home.stats.uptimeVal') },
  { label: t('home.stats.latency'), value: t('home.stats.latencyVal') },
  { label: t('home.stats.regions'), value: t('home.stats.regionsVal') },
])

const partners = ['Alipay', 'WeChat Pay', 'Stripe', 'UnionPay']

const regionKeys = [ 'us', 'jp', 'in', 'br'] as const
const regionFlags: Record<(typeof regionKeys)[number], string> = {
  us: '🇺🇸',
  jp: '🇯🇵',
  in: '🇮🇳',
  br: '🇧🇷',
}

const cryptoKeys = ['onchain', 'stablecoin', 'compliance', 'experience'] as const
</script>
