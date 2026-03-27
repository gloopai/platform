<template>
  <div class="w-full" :class="embedded ? '' : 'rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm'">
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div class="min-w-0">
        <div class="text-xs font-semibold text-slate-900">产品信息</div>
        <p class="mt-0.5 max-w-xl text-[11px] leading-snug text-slate-500">
          {{ model.id ? `编辑 #${model.id}；编码保存后请避免随意改动以免影响对接。` : '新建产品：填写对外 code 与展示名称，保存后可配置通道绑定。' }}
        </p>
      </div>
      <div v-if="saved" class="text-[11px] font-semibold text-emerald-700">已保存</div>
    </div>

    <div class="mt-3 space-y-4">
      <div class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">基本字段</div>
        <p class="mt-0.5 text-[11px] text-slate-500">排序数字越小越靠前；停用后商户侧不可选用该产品。</p>
        <div class="mt-2.5 grid gap-2.5 sm:grid-cols-12">
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-4">
            编码 code
            <input
              :value="model.code"
              autocomplete="off"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
              @input="patch({ code: ($event.target as HTMLInputElement).value.trim() })"
            />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-5">
            展示名称
            <input
              :value="model.name"
              autocomplete="off"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm"
              @input="patch({ name: ($event.target as HTMLInputElement).value.trim() })"
            />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-3">
            排序
            <input
              :value="model.sort_order"
              type="number"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
              @input="patch({ sort_order: Number(($event.target as HTMLInputElement).value) })"
            />
          </label>
          <label
            class="flex items-center justify-between gap-3 rounded-md border border-slate-200/80 bg-white px-2.5 py-2 sm:col-span-12"
          >
            <div>
              <div class="text-[11px] font-medium text-slate-700">启用该产品</div>
              <p class="mt-0.5 text-[10px] text-slate-500">关闭后路由与历史订单不受影响，新单不可选。</p>
            </div>
            <input
              :checked="model.enabled"
              type="checkbox"
              class="h-4 w-4 shrink-0 rounded border-slate-300 text-slate-900"
              @change="patch({ enabled: ($event.target as HTMLInputElement).checked })"
            />
          </label>
        </div>
      </div>
    </div>

    <div v-if="error" class="mt-3 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-[11px] text-rose-800">
      {{ error }}
    </div>
    <div v-if="!hideFooterActions" class="mt-4 flex flex-wrap gap-2">
      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
        :disabled="saving || !canSave"
        @click="$emit('save')"
      >
        {{ saving ? '保存中...' : '保存产品' }}
      </button>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50"
        @click="$emit('reset')"
      >
        重置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PayinProduct } from './types'

const props = withDefaults(
  defineProps<{
    model: PayinProduct
    saving: boolean
    saved: boolean
    error: string
    canSave: boolean
    embedded?: boolean
    hideFooterActions?: boolean
  }>(),
  { embedded: false, hideFooterActions: false },
)

const emit = defineEmits<{
  save: []
  reset: []
  'update:model': [v: PayinProduct]
}>()

function patch(p: Partial<PayinProduct>) {
  emit('update:model', { ...props.model, ...p })
}
</script>
