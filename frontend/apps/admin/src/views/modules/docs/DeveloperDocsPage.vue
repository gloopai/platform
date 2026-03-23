<template>
  <div class="space-y-4">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">开发文档</h1>
      <p class="mt-1 text-sm text-slate-600">商户接入相关 API 文档统一入口（直接读取仓库 Markdown）。</p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="flex flex-wrap gap-2 border-b border-slate-200 p-3">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="rounded-lg border px-3 py-1.5 text-xs font-semibold transition"
          :class="activeTab === tab.key
            ? 'border-slate-900 bg-slate-900 text-white'
            : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="p-4">
        <div class="mb-3 text-xs text-slate-500">
          当前文档：<span class="font-mono text-slate-700">{{ activeDoc?.path }}</span>
        </div>
        <pre class="max-h-[70vh] overflow-auto rounded-xl border border-slate-100 bg-slate-50 p-4 text-xs leading-6 text-slate-800">{{ activeDoc?.content || '' }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import merchantPlatformDoc from '../../../../../../../docs/merchant-platform.md?raw'
import checkoutPlatformDoc from '../../../../../../../docs/checkout-platform.md?raw'
import openApiErrorCodesDoc from '../../../../../../../docs/开放API错误码.md?raw'
import e2eDoc from '../../../../../../../docs/端到端联调一遍.md?raw'

const tabs = [
  { key: 'merchant', label: '商户平台文档', path: 'docs/merchant-platform.md', content: merchantPlatformDoc },
  { key: 'checkout', label: '收银台文档', path: 'docs/checkout-platform.md', content: checkoutPlatformDoc },
  { key: 'error-codes', label: '开放 API 错误码', path: 'docs/开放API错误码.md', content: openApiErrorCodesDoc },
  { key: 'e2e', label: '端到端联调', path: 'docs/端到端联调一遍.md', content: e2eDoc },
] as const

const activeTab = ref<(typeof tabs)[number]['key']>('merchant')
const activeDoc = computed(() => tabs.find((x) => x.key === activeTab.value) || tabs[0])
</script>

