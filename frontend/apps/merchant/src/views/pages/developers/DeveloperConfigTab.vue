<template>
  <section class="rounded-2xl border border-slate-200/90 bg-white p-6 shadow-sm">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h2 class="text-sm font-semibold text-slate-900">开发配置</h2>
      </div>
      <button type="button" class="btn-primary" :disabled="saving" @click="$emit('save')">
        {{ saving ? '保存中…' : '保存配置' }}
      </button>
    </div>

    <div class="mt-6 grid grid-cols-12 gap-4">
      <label class="col-span-12 grid gap-1.5 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">AppID</span>
        <input :value="merchantId" class="input-merchant bg-slate-50" disabled />
      </label>
      <label class="col-span-12 grid gap-1.5 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">AppSecret（管理台添加商户时生成）</span>
        <input :value="apiSecret" class="input-merchant bg-slate-50 font-mono text-xs" :type="secretVisible ? 'text' : 'password'" disabled />
      </label>
      <label class="col-span-12 grid gap-1.5 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">Notify URL 是否配置</span>
        <input :value="notifyUrl" class="input-merchant" placeholder="https://merchant.example.com/notify" @input="$emit('update:notifyUrl', ($event.target as HTMLInputElement).value.trim())" />
      </label>
      <label class="col-span-12 grid gap-1.5 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">IP 白名单是否配置</span>
        <input :value="ipWhitelist" class="input-merchant" placeholder="例如：127.0.0.1,10.0.0.0/24" @input="$emit('update:ipWhitelist', ($event.target as HTMLInputElement).value.trim())" />
      </label>
    </div>

    <div class="mt-4 flex items-center gap-3 text-xs">
      <button type="button" class="btn-lite" @click="secretVisible = !secretVisible">{{ secretVisible ? '隐藏 AppSecret' : '显示 AppSecret' }}</button>
      <span v-if="saveSuccess" class="text-emerald-700">{{ saveSuccess }}</span>
      <span v-if="saveError" class="text-rose-700">{{ saveError }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  merchantId: string
  apiSecret: string
  ipWhitelist: string
  notifyUrl: string
  saving: boolean
  saveError: string
  saveSuccess: string
}>()

defineEmits<{
  'update:merchantId': [value: string]
  'update:apiSecret': [value: string]
  'update:ipWhitelist': [value: string]
  'update:notifyUrl': [value: string]
  save: []
}>()

const secretVisible = ref(false)
</script>

<style scoped>
.input-merchant {
  @apply w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/20;
}
.btn-primary {
  @apply rounded-xl bg-slate-800 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-slate-900/15 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-40;
}
.btn-lite {
  @apply rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-50;
}
</style>

