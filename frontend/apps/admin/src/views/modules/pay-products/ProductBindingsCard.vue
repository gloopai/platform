<template>
  <div class="w-full" :class="embedded ? '' : 'rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm'">
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div class="min-w-0">
        <div class="text-xs font-semibold text-slate-900">上游通道绑定</div>
        <p class="mt-0.5 max-w-xl text-[11px] leading-snug text-slate-500">
          同产品下多条通道按权重加权随机。手续费规则（上游/对客）请在「通道管理」与「商户管理」配置；此处仅维护路由关系。
        </p>
      </div>
    </div>

    <div v-if="loading" class="mt-3 rounded-lg border border-dashed border-slate-200 bg-slate-50/30 px-3 py-6 text-center text-[11px] text-slate-500">
      加载绑定中…
    </div>
    <div v-else class="mt-3 overflow-hidden rounded-xl border border-slate-200/90">
      <div class="max-h-72 overflow-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="sticky top-0 z-10 border-b border-slate-200 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-3 py-2.5">通道</th>
              <th class="whitespace-nowrap px-3 py-2.5">权重</th>
              <th class="whitespace-nowrap px-3 py-2.5">启用</th>
              <th class="sticky right-0 z-20 whitespace-nowrap bg-slate-50 px-3 py-2.5 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 bg-white">
            <tr v-for="b in bindings" :key="b.id" class="group transition hover:bg-slate-50/80">
              <td class="px-3 py-2.5">
                <div class="font-mono text-xs text-slate-800">#{{ b.channel_id }}</div>
                <div class="text-[11px] font-medium text-slate-900">{{ b.channel_name || '—' }}</div>
              </td>
              <td class="px-3 py-2.5">
                <input
                  :value="rowWeight[b.id]"
                  type="number"
                  min="1"
                  class="w-24 rounded-md border border-slate-200 bg-white px-2 py-1.5 text-sm tabular-nums"
                  @input="setWeight(b.id, Number(($event.target as HTMLInputElement).value))"
                />
              </td>
              <td class="px-3 py-2.5">
                <input
                  :checked="rowEnabled[b.id]"
                  type="checkbox"
                  class="h-4 w-4 rounded border-slate-300 text-slate-900"
                  @change="setEnabled(b.id, ($event.target as HTMLInputElement).checked)"
                />
              </td>
              <td class="sticky right-0 z-10 bg-white px-3 py-2.5 text-right group-hover:bg-slate-50/80">
                <div class="flex flex-wrap items-center justify-end gap-1.5">
                  <button
                    type="button"
                    class="rounded-md border border-slate-200 bg-white px-2 py-1 text-[11px] font-semibold text-slate-700 hover:bg-slate-50"
                    @click="emitSaveRow(b)"
                  >
                    保存
                  </button>
                  <button
                    type="button"
                    class="rounded-md border border-rose-200 bg-rose-50 px-2 py-1 text-[11px] font-semibold text-rose-700 hover:bg-rose-100"
                    @click="$emit('delete-row', b.id)"
                  >
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="mt-4 rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
      <div class="text-xs font-semibold text-slate-800">新增绑定</div>
      <p v-if="channels.length > 80" class="mt-0.5 text-[11px] text-slate-500">
        通道较多时用搜索选择；已绑定的通道不会出现在待选列表。
      </p>
      <div class="mt-2.5 flex flex-wrap items-end gap-2.5">
        <ChannelPicker
          :channels="channels"
          :model-value="draft.channel_id"
          :exclude-channel-ids="excludeChannelIds"
          @update:model-value="emitDraft({ channel_id: $event })"
        />
        <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
          权重
          <input
            :value="draft.weight"
            type="number"
            min="1"
            class="w-24 rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            @input="emitDraft({ weight: Number(($event.target as HTMLInputElement).value) })"
          />
        </label>
        <label class="flex items-center gap-2 pb-0.5">
          <input
            :checked="draft.enabled"
            type="checkbox"
            class="h-4 w-4 rounded border-slate-300 text-slate-900"
            @change="emitDraft({ enabled: ($event.target as HTMLInputElement).checked })"
          />
          <span class="text-[11px] font-medium text-slate-700">启用</span>
        </label>
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
          :disabled="adding || draft.channel_id <= 0"
          @click="$emit('add')"
        >
          {{ adding ? '提交中...' : '添加绑定' }}
        </button>
      </div>
      <div v-if="error" class="mt-2.5 rounded-md border border-rose-200 bg-rose-50 px-2.5 py-2 text-[11px] text-rose-800">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'

import ChannelPicker from '../../../components/ChannelPicker.vue'

import type { PayinProductBinding, PayinProductChannelOption } from './types'

const props = withDefaults(
  defineProps<{
    bindings: PayinProductBinding[]
    channels: PayinProductChannelOption[]
    excludeChannelIds: number[]
    loading: boolean
    error: string
    adding: boolean
    draft: { channel_id: number; weight: number; enabled: boolean }
    embedded?: boolean
  }>(),
  { embedded: false },
)

const emit = defineEmits<{
  'update:draft': [v: { channel_id: number; weight: number; enabled: boolean }]
  'save-row': [payload: { id: number; weight: number; enabled: boolean }]
  'delete-row': [bindingId: number]
  add: []
}>()

const rowWeight = reactive<Record<number, number>>({})
const rowEnabled = reactive<Record<number, boolean>>({})

function syncRows(rows: PayinProductBinding[]) {
  Object.keys(rowWeight).forEach((k) => delete rowWeight[Number(k)])
  Object.keys(rowEnabled).forEach((k) => delete rowEnabled[Number(k)])
  for (const b of rows) {
    rowWeight[b.id] = b.weight
    rowEnabled[b.id] = b.enabled
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

function emitDraft(p: Partial<{ channel_id: number; weight: number; enabled: boolean }>) {
  emit('update:draft', { ...props.draft, ...p })
}

function emitSaveRow(b: PayinProductBinding) {
  const w = rowWeight[b.id]
  const en = rowEnabled[b.id]
  emit('save-row', {
    id: b.id,
    weight: w !== undefined ? w : b.weight,
    enabled: en !== undefined ? en : b.enabled,
  })
}
</script>
