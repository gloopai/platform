<template>
  <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6 lg:px-8 lg:py-20">
    <header class="max-w-2xl">
      <h1 class="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">{{ t('docs.title') }}</h1>
      <p class="mt-4 text-base leading-relaxed text-slate-600">{{ t('docs.lead') }}</p>
    </header>

    <section class="mt-12 rounded-3xl border border-slate-200 bg-white p-8 shadow-sm sm:p-10">
      <h2 class="text-lg font-semibold text-slate-900">{{ t('docs.quick.title') }}</h2>
      <div class="mt-8 grid gap-4 md:grid-cols-3">
        <div class="rounded-2xl border border-slate-100 bg-slate-50/80 p-5">
          <div class="text-xs font-bold uppercase tracking-wide text-brand-700">{{ t('docs.quick.step1.title') }}</div>
          <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ t('docs.quick.step1.desc') }}</p>
        </div>
        <div class="rounded-2xl border border-slate-100 bg-slate-50/80 p-5">
          <div class="text-xs font-bold uppercase tracking-wide text-brand-700">{{ t('docs.quick.step2.title') }}</div>
          <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ t('docs.quick.step2.desc') }}</p>
        </div>
        <div class="rounded-2xl border border-slate-100 bg-slate-50/80 p-5">
          <div class="text-xs font-bold uppercase tracking-wide text-brand-700">{{ t('docs.quick.step3.title') }}</div>
          <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ t('docs.quick.step3.desc') }}</p>
        </div>
      </div>
    </section>

    <div class="mt-8 grid gap-6 md:grid-cols-2">
      <section class="rounded-3xl border border-slate-200 bg-white p-8 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-900">{{ t('docs.sdk.title') }}</h2>
        <ul class="mt-4 space-y-2 text-sm">
          <li>
            <a class="font-medium text-brand-700 hover:text-brand-800" href="#" @click.prevent>{{ t('docs.sdk.go') }}</a>
          </li>
          <li>
            <a class="font-medium text-brand-700 hover:text-brand-800" href="#" @click.prevent>{{ t('docs.sdk.php') }}</a>
          </li>
          <li>
            <a class="font-medium text-brand-700 hover:text-brand-800" href="#" @click.prevent>{{ t('docs.sdk.java') }}</a>
          </li>
          <li>
            <a class="font-medium text-brand-700 hover:text-brand-800" href="#" @click.prevent>{{ t('docs.sdk.python') }}</a>
          </li>
        </ul>
      </section>
      <section class="rounded-3xl border border-slate-200 bg-white p-8 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-900">{{ t('docs.online.title') }}</h2>
        <p class="mt-2 text-sm leading-relaxed text-slate-600">{{ t('docs.online.lead') }}</p>
        <div class="mt-6 flex flex-wrap gap-3">
          <a
            class="inline-flex rounded-xl border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-800 transition hover:border-brand-200 hover:bg-brand-50"
            href="#"
            @click.prevent
            >{{ t('docs.online.swagger') }}</a
          >
          <a
            class="inline-flex rounded-xl border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-800 transition hover:border-brand-200 hover:bg-brand-50"
            href="#"
            @click.prevent
            >{{ t('docs.online.gitbook') }}</a
          >
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
