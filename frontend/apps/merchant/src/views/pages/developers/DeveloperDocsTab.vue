<template>
  <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
    <div class="flex flex-wrap gap-2">
      <button
        v-for="doc in docs"
        :key="doc.key"
        type="button"
        class="rounded-lg border px-3 py-1.5 text-xs font-semibold transition"
        :class="activeKey === doc.key ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300'"
        @click="activeKey = doc.key"
      >
        {{ doc.label }}
      </button>
    </div>
    <div class="mt-4 text-xs text-slate-500">
      文档路径：<span class="font-mono text-slate-700">{{ activeDoc.path }}</span>
    </div>
    <article class="doc-markdown mt-3 rounded-2xl border border-slate-200 bg-slate-50 p-5" v-html="activeHtml" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { marked } from 'marked'
import createCollectOrderDoc from '@/devdocs/create-collect-order.md?raw'
import createPayoutOrderDoc from '@/devdocs/create-payout-order.md?raw'
import queryCollectOrderDoc from '@/devdocs/query-collect-order.md?raw'
import queryPayoutOrderDoc from '@/devdocs/query-payout-order.md?raw'
import errorCodesDoc from '@/devdocs/error-codes.md?raw'
import signingDoc from '@/devdocs/signing.md?raw'

const docs = [
  { key: 'create-collect', label: '创建代收', path: 'src/devdocs/create-collect-order.md', content: createCollectOrderDoc },
  { key: 'create-payout', label: '创建代付', path: 'src/devdocs/create-payout-order.md', content: createPayoutOrderDoc },
  { key: 'query-collect', label: '查询代收状态', path: 'src/devdocs/query-collect-order.md', content: queryCollectOrderDoc },
  { key: 'query-payout', label: '查询代付状态', path: 'src/devdocs/query-payout-order.md', content: queryPayoutOrderDoc },
  { key: 'error-codes', label: '错误码', path: 'src/devdocs/error-codes.md', content: errorCodesDoc },
  { key: 'signing', label: '接口签名', path: 'src/devdocs/signing.md', content: signingDoc },
] as const

const activeKey = ref<(typeof docs)[number]['key']>('create-collect')
const activeDoc = computed(() => docs.find((x) => x.key === activeKey.value) || docs[0])
const activeHtml = computed(() => marked.parse(activeDoc.value.content, { breaks: true }) as string)
</script>

<style scoped>
.doc-markdown :deep(h1) {
  @apply mt-2 text-lg font-semibold text-slate-900;
}
.doc-markdown :deep(h2) {
  @apply mt-6 text-base font-semibold text-slate-900;
}
.doc-markdown :deep(h3) {
  @apply mt-4 text-sm font-semibold text-slate-900;
}
.doc-markdown :deep(p) {
  @apply mt-2 text-sm leading-6 text-slate-700;
}
.doc-markdown :deep(ul) {
  @apply mt-2 list-disc space-y-1 pl-5 text-sm text-slate-700;
}
.doc-markdown :deep(code) {
  @apply rounded bg-slate-200/70 px-1 py-0.5 font-mono text-xs text-slate-800;
}
.doc-markdown :deep(pre) {
  @apply mt-3 overflow-auto rounded-xl border border-slate-200 bg-slate-900 p-3 text-xs text-slate-100;
}
.doc-markdown :deep(pre code) {
  @apply bg-transparent p-0 text-slate-100;
}
.doc-markdown :deep(table) {
  @apply mt-3 w-full border-collapse overflow-hidden rounded-lg border border-slate-200 text-xs;
}
.doc-markdown :deep(th) {
  @apply border border-slate-200 bg-slate-100 px-2 py-1.5 text-left font-semibold text-slate-700;
}
.doc-markdown :deep(td) {
  @apply border border-slate-200 bg-white px-2 py-1.5 text-slate-700;
}
</style>

