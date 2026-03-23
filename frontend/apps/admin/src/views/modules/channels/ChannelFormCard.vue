<template>
  <div
    class="w-full bg-white p-6"
    :class="embedded ? '' : 'col-span-12 rounded-2xl border border-slate-200 shadow-sm md:col-span-8'"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="text-xs text-slate-500">当前：{{ model.id ? `#${model.id}` : '新建' }}</div>
      <div v-if="saved" class="text-xs font-semibold text-emerald-700">已保存</div>
    </div>

    <div class="mt-4 grid grid-cols-12 gap-4">
      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">通道名称</span>
        <input v-model.trim="model.name" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">支付类型</span>
        <input v-model.trim="model.pay_type" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">上游 API 地址</span>
        <input v-model.trim="model.gateway_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">上游商户号</span>
        <input v-model.trim="model.upstream_merchant_no" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-6">
        <span class="text-xs font-medium text-slate-600">签名密钥（Sign Secret）</span>
        <input v-model.trim="model.sign_secret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <label class="col-span-12 grid gap-1">
        <span class="text-xs font-medium text-slate-600">RSA 私钥</span>
        <textarea v-model="model.rsa_private_key" rows="7" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-xs" />
      </label>

      <label class="col-span-12 grid gap-1 md:col-span-4">
        <span class="text-xs font-medium text-slate-600">权重（0-100）</span>
        <input v-model.number="model.weight" type="number" min="0" max="100" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-4">
        <span class="text-xs font-medium text-slate-600">单笔最小金额（分）</span>
        <input v-model.number="model.min_amount" type="number" min="0" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-4">
        <span class="text-xs font-medium text-slate-600">单笔最大金额（分）</span>
        <input v-model.number="model.max_amount" type="number" min="0" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
      </label>

      <div class="col-span-12 grid grid-cols-12 gap-4">
        <div class="col-span-12 rounded-xl border border-slate-200/90 bg-slate-50/60 p-3">
          <div class="mb-2 text-xs font-semibold text-slate-600">能力开关</div>
          <div class="grid grid-cols-12 gap-3">
            <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 bg-white px-3 py-2 md:col-span-6">
              <div class="text-sm text-slate-700">支持代收</div>
              <input v-model="model.supports_payin" type="checkbox" class="h-4 w-4" />
            </label>
            <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 bg-white px-3 py-2 md:col-span-6">
              <div class="text-sm text-slate-700">支持代付</div>
              <input v-model="model.supports_payout" type="checkbox" class="h-4 w-4" />
            </label>
          </div>
        </div>

        <div class="col-span-12 rounded-xl border border-slate-200/90 bg-slate-50/60 p-3">
          <div class="mb-2 text-xs font-semibold text-slate-600">代收费率配置</div>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">上游代收费率（万分比）</span>
            <input
              v-model.number="model.upstream_payin_rate_bps"
              type="number"
              min="0"
              class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
            />
          </label>
        </div>

        <div class="col-span-12 rounded-xl border border-slate-200/90 bg-slate-50/60 p-3">
          <div class="mb-2 text-xs font-semibold text-slate-600">代付费率配置</div>
          <div v-if="!model.supports_payout" class="rounded-md border border-dashed border-slate-200 bg-white px-3 py-3 text-xs text-slate-500">
            当前通道未开启代付能力，开启“支持代付”后可配置代付费率模式。
          </div>
          <div v-else class="grid grid-cols-12 gap-3">
            <label class="col-span-12 grid gap-1 md:col-span-4">
              <span class="text-xs font-medium text-slate-600">上游代付费率（万分比）</span>
              <input
                v-model.number="model.upstream_payout_rate_bps"
                type="number"
                min="0"
                class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
              />
            </label>
            <label class="col-span-12 grid gap-1 md:col-span-4">
              <span class="text-xs font-medium text-slate-600">上游代付费率模式</span>
              <select v-model.number="model.upstream_payout_fee_mode" class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm">
                <option :value="1">比例</option>
                <option :value="2">固定金额</option>
                <option :value="3">固定+比例</option>
              </select>
            </label>
            <label class="col-span-12 grid gap-1 md:col-span-4">
              <span class="text-xs font-medium text-slate-600">上游代付固定手续费（分）</span>
              <input
                v-model.number="model.upstream_payout_fixed_fee"
                type="number"
                min="0"
                class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
              />
            </label>
          </div>
        </div>

        <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-6">
          <div class="text-sm text-slate-700">启用通道</div>
          <input v-model="model.enabled" type="checkbox" class="h-4 w-4" />
        </label>
        <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-6">
          <div class="text-sm text-slate-700">熔断开关</div>
          <input v-model="model.fuse_enabled" type="checkbox" class="h-4 w-4" />
        </label>
      </div>
    </div>

    <div v-if="error" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
      {{ error }}
    </div>

    <div v-if="!hideFooterActions" class="mt-6 flex flex-wrap items-center gap-3">
      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
        :disabled="saving || !canSave"
        @click="$emit('save')"
      >
        {{ saving ? '保存中...' : '保存配置' }}
      </button>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
        @click="$emit('reset')"
      >
        重置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AdminChannel } from './types'

const model = defineModel<AdminChannel>({ required: true })

withDefaults(
  defineProps<{
    saving: boolean
    saved: boolean
    error: string
    canSave: boolean
    /** 抽屉内嵌套时不画外框 */
    embedded?: boolean
    hideFooterActions?: boolean
  }>(),
  { embedded: false, hideFooterActions: false },
)

defineEmits<{
  save: []
  reset: []
}>()
</script>
