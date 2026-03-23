<template>
  <div
    class="w-full bg-white p-6"
    :class="embedded ? '' : 'rounded-2xl border border-slate-200 shadow-sm'"
  >
    <template v-if="!embedded">
      <div class="text-sm font-semibold text-slate-900">代付产品绑定</div>
      <p class="mt-1 text-xs text-slate-500">与代收独立配置；仅开通代付产品时该商户才具备对应代付能力。</p>
    </template>
    <p v-else class="text-xs text-slate-500">以下为该商户可用的代付产品编码（后续代付 API 将校验此白名单）。</p>

    <div v-if="loading" class="mt-4 text-sm text-slate-500">加载...</div>
    <div v-else class="mt-4 max-h-72 overflow-auto rounded-lg border border-slate-100">
      <table class="min-w-full text-left text-sm">
        <thead class="sticky top-0 z-10 border-b border-slate-200 bg-white text-xs text-slate-500 shadow-sm">
          <tr>
            <th class="py-2 pr-3">产品</th>
            <th class="py-2 pr-3">编码</th>
            <th class="py-2 pr-3">费率模式</th>
            <th class="py-2 pr-3">比例(bps)</th>
            <th class="py-2 pr-3">固定金额(分)</th>
            <th class="py-2">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in boundRows" :key="row.id" class="border-b border-slate-100">
            <td class="py-2 pr-3 font-medium text-slate-900">{{ row.name }}</td>
            <td class="py-2 pr-3 font-mono text-xs text-slate-600">{{ row.code }}</td>
            <td class="py-2 pr-3">
              <select
                v-model.number="row.fee_mode"
                class="rounded-md border border-slate-200 px-2 py-1 text-xs"
                :disabled="saving"
                @change="emitUpdate(row)"
              >
                <option :value="1">比例</option>
                <option :value="2">固定金额</option>
                <option :value="3">固定+比例</option>
              </select>
            </td>
            <td class="py-2 pr-3">
              <input
                v-model.number="row.merchant_rate_bps"
                type="number"
                min="0"
                class="w-24 rounded-md border border-slate-200 px-2 py-1 text-xs"
                :disabled="saving"
                @change="emitUpdate(row)"
              />
            </td>
            <td class="py-2 pr-3">
              <input
                v-model.number="row.fee_fixed_amount"
                type="number"
                min="0"
                class="w-24 rounded-md border border-slate-200 px-2 py-1 text-xs"
                :disabled="saving"
                @change="emitUpdate(row)"
              />
            </td>
            <td class="py-2">
              <button
                type="button"
                class="text-xs font-semibold text-rose-700 underline disabled:opacity-40"
                :disabled="saving"
                @click="$emit('remove', row.id)"
              >
                移除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!boundRows.length" class="px-3 py-6 text-center text-sm text-slate-500">尚未绑定代付产品</div>
    </div>

    <div class="mt-6 rounded-xl border border-dashed border-slate-200 p-4">
      <div class="text-xs font-semibold text-slate-600">新增绑定</div>
      <p class="mt-1 text-[11px] text-slate-500">仅显示未绑定的代付产品。</p>
      <div class="mt-3 flex flex-wrap items-end gap-3">
        <label class="grid min-w-[200px] flex-1 gap-1">
          <span class="text-xs text-slate-500">代付产品</span>
          <select v-model.number="localPick" class="rounded-md border border-slate-200 px-3 py-2 text-sm">
            <option :value="0">请选择</option>
            <option v-for="p in availableToAdd" :key="p.id" :value="p.id">{{ p.code }} — {{ p.name }}</option>
          </select>
        </label>
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="saving || localPick <= 0"
          @click="emitAdd"
        >
          {{ saving ? '提交...' : '添加' }}
        </button>
      </div>
      <div v-if="bindError" class="mt-3 text-sm text-rose-700">{{ bindError }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { MerchantPayoutGrant, PayProductRow } from './types'

const props = withDefaults(
  defineProps<{
    grants: MerchantPayoutGrant[]
    catalog: PayProductRow[]
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
  update: [grant: MerchantPayoutGrant]
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
      const id = g.payout_product_id
      const p = cat.find((x) => x.id === id)
      return p
        ? {
            id: p.id,
            code: p.code,
            name: p.name,
            fee_mode: g.fee_mode || 1,
            merchant_rate_bps: g.merchant_rate_bps ?? 0,
            fee_fixed_amount: g.fee_fixed_amount ?? 0,
          }
        : {
            id,
            code: `#${id}`,
            name: '（未知产品）',
            fee_mode: g.fee_mode || 1,
            merchant_rate_bps: g.merchant_rate_bps ?? 0,
            fee_fixed_amount: g.fee_fixed_amount ?? 0,
          }
    })
    .sort((a, b) => a.code.localeCompare(b.code))
})

const boundSet = computed(() => new Set((props.grants || []).map((g) => g.payout_product_id)))

const availableToAdd = computed(() => props.catalog.filter((p) => !boundSet.value.has(p.id)))

function emitAdd() {
  if (localPick.value <= 0) return
  emit('add', localPick.value)
  localPick.value = 0
}

function emitUpdate(row: { id: number; fee_mode: number; merchant_rate_bps: number; fee_fixed_amount: number }) {
  emit('update', {
    payout_product_id: row.id,
    fee_mode: row.fee_mode || 1,
    merchant_rate_bps: row.merchant_rate_bps ?? 0,
    fee_fixed_amount: row.fee_fixed_amount ?? 0,
  })
}
</script>
