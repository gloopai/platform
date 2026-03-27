<template>
  <div
    class="w-full bg-white p-6"
    :class="embedded ? '' : 'rounded-2xl border border-slate-200 shadow-sm'"
  >
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="text-xs text-slate-500">
        当前：{{ isNew ? '新建商户' : model.merchant_id ? `「${model.merchant_id}」` : '—' }}
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <div v-if="saved" class="text-xs font-semibold text-emerald-700">已保存</div>
        <template v-if="!isNew && model.merchant_id">
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700"
            @click="$emit('toggle-lock')"
          >
            {{ lockLabel }}
          </button>
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700"
            @click="$emit('reset-secret')"
          >
            重置密钥
          </button>
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700"
            @click="$emit('reset-password')"
          >
            重置密码
          </button>
        </template>
      </div>
    </div>

    <div class="mt-4 grid grid-cols-12 gap-4">
      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">merchant_id</span>
        <input
          v-model.trim="model.merchant_id"
          class="rounded-md border border-slate-200 px-3 py-2 text-sm"
          :disabled="!isNew"
        />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">邮箱 email</span>
        <input v-model.trim="model.email" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">状态</span>
        <select v-model.number="model.status" class="rounded-md border border-slate-200 px-3 py-2 text-sm">
          <option :value="1">启用</option>
          <option :value="0">锁定</option>
        </select>
      </label>

      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">Notify URL</span>
        <input v-model.trim="model.notify_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>
      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">Return URL</span>
        <input v-model.trim="model.return_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">提现 USDT 地址</span>
        <input v-model.trim="model.withdraw_usdt_address" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-sm" />
      </label>

      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">IP 白名单</span>
        <textarea v-model="model.ip_whitelist" rows="5" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-xs" />
      </label>
    </div>

    <div v-if="error" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
      {{ error }}
    </div>

    <div v-if="!hideFooterActions" class="mt-6 flex flex-wrap items-center gap-3">
      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
        :disabled="saving || !canSave"
        @click="$emit('save')"
      >
        {{ saving ? '保存中...' : '保存配置' }}
      </button>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
        @click="$emit('reset')"
      >
        重置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { MerchantForm } from './types'

const model = defineModel<MerchantForm>({ required: true })

const props = withDefaults(
  defineProps<{
    isNew: boolean
    saving: boolean
    saved: boolean
    error: string
    canSave: boolean
    statusForLock: number
    /** 嵌入 Tab 面板时不画外框 */
    embedded?: boolean
    /** 抽屉底部统一放保存/关闭时隐藏表单底部按钮 */
    hideFooterActions?: boolean
  }>(),
  { embedded: false, hideFooterActions: false },
)

defineEmits<{
  save: []
  reset: []
  'toggle-lock': []
  'reset-secret': []
  'reset-password': []
}>()

const lockLabel = computed(() => (props.statusForLock === 1 ? '锁定' : '解锁'))
</script>
