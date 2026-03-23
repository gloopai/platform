<template>
  <div
    class="w-full bg-white p-6"
    :class="embedded ? '' : 'rounded-2xl border border-slate-200 shadow-sm'"
  >
    <div class="text-sm font-semibold text-slate-900">上游通道绑定</div>
    <p class="mt-1 text-xs text-slate-500">
      同产品下多条通道按权重加权随机；「上游成本」留空表示使用通道默认值。
    </p>

    <div v-if="loading" class="mt-4 text-sm text-slate-500">加载绑定...</div>
    <div v-else class="mt-4 max-h-72 overflow-auto rounded-lg border border-slate-100">
      <table class="min-w-full text-left text-sm">
        <thead class="sticky top-0 z-10 border-b border-slate-200 bg-white text-xs text-slate-500 shadow-sm">
          <tr>
            <th class="py-2 pr-3">通道</th>
            <th class="py-2 pr-3">权重</th>
            <th class="py-2 pr-3">上游成本(bps)</th>
            <th class="py-2 pr-3">启用</th>
            <th class="py-2">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in bindings" :key="b.id" class="border-b border-slate-100">
            <td class="py-2 pr-3">
              <div class="font-medium text-slate-900">#{{ b.channel_id }} {{ b.channel_name || '-' }}</div>
            </td>
            <td class="py-2 pr-3">
              <input
                :value="rowWeight[b.id]"
                type="number"
                min="1"
                class="w-24 rounded border border-slate-200 px-2 py-1 text-sm"
                @input="setWeight(b.id, Number(($event.target as HTMLInputElement).value))"
              />
            </td>
            <td class="py-2 pr-3">
              <input
                :value="rowCostDisplay[b.id]"
                type="text"
                inputmode="numeric"
                placeholder="默认"
                class="w-24 rounded border border-slate-200 px-2 py-1 text-sm"
                @input="setCostRaw(b.id, ($event.target as HTMLInputElement).value)"
              />
            </td>
            <td class="py-2 pr-3">
              <input
                :checked="rowEnabled[b.id]"
                type="checkbox"
                class="h-4 w-4"
                @change="setEnabled(b.id, ($event.target as HTMLInputElement).checked)"
              />
            </td>
            <td class="py-2">
              <button
                type="button"
                class="mr-2 text-xs font-semibold text-slate-700 underline"
                @click="emitSaveRow(b)"
              >
                保存
              </button>
              <button
                type="button"
                class="text-xs font-semibold text-rose-700 underline"
                @click="$emit('delete-row', b.id)"
              >
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-6 rounded-xl border border-dashed border-slate-200 p-4">
      <div class="text-xs font-semibold text-slate-600">新增绑定</div>
      <p v-if="channels.length > 80" class="mt-1 text-[11px] text-slate-500">
        通道较多时在下方用搜索选择；已在本产品中绑定的通道不会重复出现。
      </p>
      <div class="mt-3 flex flex-wrap items-end gap-3">
        <ChannelPicker
          :channels="channels"
          :model-value="draft.channel_id"
          :exclude-channel-ids="excludeChannelIds"
          @update:model-value="emitDraft({ channel_id: $event })"
        />
        <label class="grid gap-1">
          <span class="text-xs text-slate-500">权重</span>
          <input
            :value="draft.weight"
            type="number"
            min="1"
            class="w-24 rounded-md border border-slate-200 px-3 py-2 text-sm"
            @input="emitDraft({ weight: Number(($event.target as HTMLInputElement).value) })"
          />
        </label>
        <label class="grid gap-1">
          <span class="text-xs text-slate-500">上游成本(bps)</span>
          <input
            :value="draftCostStr"
            type="text"
            inputmode="numeric"
            placeholder="默认"
            class="w-24 rounded-md border border-slate-200 px-3 py-2 text-sm"
            @input="onDraftCostInput(($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="flex items-center gap-2 pb-2">
          <input
            :checked="draft.enabled"
            type="checkbox"
            class="h-4 w-4"
            @change="emitDraft({ enabled: ($event.target as HTMLInputElement).checked })"
          />
          <span class="text-sm text-slate-700">启用</span>
        </label>
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="adding || draft.channel_id <= 0"
          @click="$emit('add')"
        >
          {{ adding ? '提交...' : '添加' }}
        </button>
      </div>
      <div v-if="error" class="mt-3 text-sm text-rose-700">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import ChannelPicker from '../../../components/ChannelPicker.vue'

import type { PayProductBinding, PayProductChannelOption } from './types'

const props = withDefaults(
  defineProps<{
    bindings: PayProductBinding[]
    channels: PayProductChannelOption[]
    excludeChannelIds: number[]
    loading: boolean
    error: string
    adding: boolean
    draft: { channel_id: number; weight: number; enabled: boolean; cost_rate_bps?: number | null }
    embedded?: boolean
  }>(),
  { embedded: false },
)

const emit = defineEmits<{
  'update:draft': [v: { channel_id: number; weight: number; enabled: boolean; cost_rate_bps?: number | null }]
  'save-row': [payload: { id: number; weight: number; enabled: boolean; cost_rate_bps?: number | null }]
  'delete-row': [bindingId: number]
  add: []
}>()

const rowWeight = reactive<Record<number, number>>({})
const rowEnabled = reactive<Record<number, boolean>>({})
/** 本地编辑中的成本：undefined=未改，null=清空用默认，number=覆盖 */
const rowCost = reactive<Record<number, number | null | undefined>>({})
const rowCostDisplay = reactive<Record<number, string>>({})

function syncRows(rows: PayProductBinding[]) {
  Object.keys(rowWeight).forEach((k) => delete rowWeight[Number(k)])
  Object.keys(rowEnabled).forEach((k) => delete rowEnabled[Number(k)])
  Object.keys(rowCost).forEach((k) => delete rowCost[Number(k)])
  Object.keys(rowCostDisplay).forEach((k) => delete rowCostDisplay[Number(k)])
  for (const b of rows) {
    rowWeight[b.id] = b.weight
    rowEnabled[b.id] = b.enabled
    rowCost[b.id] = b.cost_rate_bps === undefined || b.cost_rate_bps === null ? undefined : b.cost_rate_bps
    rowCostDisplay[b.id] =
      b.cost_rate_bps === undefined || b.cost_rate_bps === null ? '' : String(b.cost_rate_bps)
  }
}

watch(
  () => props.bindings,
  (b) => syncRows(b),
  { immediate: true, deep: true },
)

function setWeight(id: number, v: number) {
  rowWeight[id] = v
}

function setEnabled(id: number, v: boolean) {
  rowEnabled[id] = v
}

function setCostRaw(id: number, raw: string) {
  rowCostDisplay[id] = raw
  const t = raw.trim()
  if (t === '') {
    rowCost[id] = null
    return
  }
  const n = Number(t)
  rowCost[id] = Number.isFinite(n) ? n : undefined
}

const draftCostStr = computed(() => {
  const v = props.draft.cost_rate_bps
  if (v === undefined || v === null) return ''
  return String(v)
})

function onDraftCostInput(raw: string) {
  const t = raw.trim()
  if (t === '') {
    emit('update:draft', { ...props.draft, cost_rate_bps: null })
    return
  }
  const n = Number(t)
  emit('update:draft', { ...props.draft, cost_rate_bps: Number.isFinite(n) ? n : undefined })
}

function emitDraft(p: Partial<{ channel_id: number; weight: number; enabled: boolean; cost_rate_bps?: number | null }>) {
  emit('update:draft', { ...props.draft, ...p })
}

function emitSaveRow(b: PayProductBinding) {
  const w = rowWeight[b.id]
  const en = rowEnabled[b.id]
  const c = rowCost[b.id]
  emit('save-row', {
    id: b.id,
    weight: w !== undefined ? w : b.weight,
    enabled: en !== undefined ? en : b.enabled,
    cost_rate_bps: c === undefined ? undefined : c,
  })
}
</script>
