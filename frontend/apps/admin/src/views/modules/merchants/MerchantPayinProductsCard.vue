<template>
  <div class="w-full" :class="embedded ? '' : 'rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm'">
    <template v-if="!embedded">
      <div class="text-xs font-semibold text-slate-900">支付产品绑定</div>
      <p class="mt-0.5 text-[11px] text-slate-500">开放给该商户收银台的代收产品；未绑定则无可用支付方式。</p>
    </template>
    <p v-else class="text-[11px] leading-snug text-slate-500">
      收银台可见的代收产品；费率单位为 bps（万分比）。
    </p>

    <div v-if="loading" class="mt-3 rounded-lg border border-dashed border-slate-200 bg-slate-50/50 py-6 text-center text-[11px] text-slate-500">
      加载中…
    </div>
    <div v-else class="mt-3">
      <div
        v-if="boundRows.length"
        class="custom-scrollbar max-h-72 space-y-2 overflow-y-auto pr-0.5"
      >
        <div
          v-for="row in boundRows"
          :key="row.id"
          class="rounded-lg border border-slate-200/90 bg-slate-50/40 px-3 py-2.5 transition hover:border-slate-300/90"
        >
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0">
              <div class="text-sm font-semibold text-slate-900">{{ row.name }}</div>
              <div class="mt-0.5 inline-flex items-center gap-1.5">
                <span class="rounded bg-slate-200/80 px-1.5 py-0.5 font-mono text-[10px] font-medium text-slate-700">
                  {{ row.code }}
                </span>
              </div>
            </div>
            <div class="flex shrink-0 flex-wrap items-center gap-2 sm:justify-end">
              <label class="flex items-center gap-1.5">
                <span class="whitespace-nowrap text-[11px] text-slate-500">费率</span>
                <input
                  v-model.number="row.merchant_rate_bps"
                  type="number"
                  min="0"
                  class="w-[4.5rem] rounded-md border border-slate-200 bg-white px-2 py-1 text-center font-mono text-xs tabular-nums text-slate-900"
                  :disabled="saving"
                  @change="emitUpdate(row)"
                />
                <span class="text-[11px] text-slate-400">bps</span>
              </label>
              <button
                type="button"
                class="rounded-md border border-rose-200/90 bg-rose-50 px-2 py-1 text-[11px] font-semibold text-rose-800 transition hover:bg-rose-100 disabled:opacity-40"
                :disabled="saving"
                @click="$emit('remove', row.id)"
              >
                移除
              </button>
            </div>
          </div>
        </div>
      </div>
      <div
        v-else
        class="rounded-lg border border-dashed border-slate-200 bg-slate-50/30 py-8 text-center text-[11px] text-slate-500"
      >
        尚未绑定任何代收产品
      </div>
    </div>

    <div class="mt-3 rounded-xl border border-slate-200/90 bg-slate-50/40 p-3">
      <div class="text-[11px] font-semibold text-slate-800">新增绑定</div>
      <p class="mt-0.5 text-[10px] text-slate-500">仅列出未绑定的产品。</p>
      <div class="mt-2.5 flex flex-col gap-2 sm:flex-row sm:items-end">
        <label class="min-w-0 flex-1 grid gap-0.5">
          <span class="text-[11px] font-medium text-slate-600">支付产品</span>
          <select
            v-model.number="localPick"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm"
          >
            <option :value="0">请选择</option>
            <option v-for="p in availableToAdd" :key="p.id" :value="p.id">{{ p.code }} — {{ p.name }}</option>
          </select>
        </label>
        <button
          type="button"
          class="shrink-0 rounded-md bg-slate-900 px-3 py-1.5 text-[11px] font-semibold text-white disabled:opacity-40"
          :disabled="saving || localPick <= 0"
          @click="emitAdd"
        >
          {{ saving ? '提交…' : '添加' }}
        </button>
      </div>
      <div v-if="bindError" class="mt-2 rounded-md border border-rose-200 bg-rose-50 px-2 py-1.5 text-[11px] text-rose-800">
        {{ bindError }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { MerchantPayinGrant, ProductRow } from './types'

const props = withDefaults(
  defineProps<{
    grants: MerchantPayinGrant[]
    catalog: ProductRow[]
    loading: boolean
    saving: boolean
    bindError: string
    embedded?: boolean
  }>(),
  { embedded: false },
)

const emit = defineEmits<{
  remove: [productId: number]
  add: [productId: number]
  update: [grant: MerchantPayinGrant]
}>()

const localPick = ref(0)

watch(
  () => props.grants,
  () => {
    localPick.value = 0
  },
)

const boundRows = computed(() => {
  const grants = props.grants || []
  const cat = props.catalog
  return grants
    .map((g) => {
      const id = g.payin_product_id
      const p = cat.find((x) => x.id === id)
      return p
        ? { id: p.id, code: p.code, name: p.name, merchant_rate_bps: g.merchant_rate_bps ?? 0 }
        : { id, code: `#${id}`, name: '（未知产品）', merchant_rate_bps: g.merchant_rate_bps ?? 0 }
    })
    .sort((a, b) => a.code.localeCompare(b.code))
})

const boundSet = computed(() => new Set((props.grants || []).map((g) => g.payin_product_id)))

const availableToAdd = computed(() =>
  props.catalog.filter((p) => !boundSet.value.has(p.id)),
)

function emitAdd() {
  if (localPick.value <= 0) return
  emit('add', localPick.value)
  localPick.value = 0
}

function emitUpdate(row: { id: number; merchant_rate_bps: number }) {
  emit('update', {
    payin_product_id: row.id,
    merchant_rate_bps: row.merchant_rate_bps ?? 0,
  })
}
</script>

<style scoped>
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: rgb(203 213 225) transparent;
}
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(203 213 225);
}
</style>
