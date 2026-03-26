<template>
  <div
    class="w-full bg-white p-6"
    :class="embedded ? '' : 'rounded-2xl border border-slate-200 shadow-sm'"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="text-xs text-slate-500">产品配置：{{ model.id ? `#${model.id}` : '新建' }}</div>
      <div v-if="saved" class="text-xs font-semibold text-emerald-700">已保存</div>
    </div>
    <div class="mt-4 grid grid-cols-12 gap-4">
      <label class="col-span-12 grid gap-1 md:col-span-4">
        <span class="text-xs font-medium text-slate-600">编码 code</span>
        <input
          :value="model.code"
          class="rounded-md border border-slate-200 px-3 py-2 font-mono text-sm"
          @input="patch({ code: ($event.target as HTMLInputElement).value.trim() })"
        />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-5">
        <span class="text-xs font-medium text-slate-600">展示名称</span>
        <input
          :value="model.name"
          class="rounded-md border border-slate-200 px-3 py-2 text-sm"
          @input="patch({ name: ($event.target as HTMLInputElement).value.trim() })"
        />
      </label>
      <label class="col-span-12 grid gap-1 md:col-span-3">
        <span class="text-xs font-medium text-slate-600">排序</span>
        <input
          :value="model.sort_order"
          type="number"
          class="rounded-md border border-slate-200 px-3 py-2 text-sm"
          @input="patch({ sort_order: Number(($event.target as HTMLInputElement).value) })"
        />
      </label>
      <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-12">
        <div class="text-sm text-slate-700">启用该产品</div>
        <input
          :checked="model.enabled"
          type="checkbox"
          class="h-4 w-4"
          @change="patch({ enabled: ($event.target as HTMLInputElement).checked })"
        />
      </label>
    </div>
    <div v-if="error" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
      {{ error }}
    </div>
    <div v-if="!hideFooterActions" class="mt-6 flex flex-wrap gap-3">
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
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
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
