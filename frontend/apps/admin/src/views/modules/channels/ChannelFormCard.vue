<template>
  <div class="w-full" :class="embedded ? '' : 'rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm md:col-span-8'">
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div class="min-w-0">
        <div class="text-xs font-semibold text-slate-900">{{ panelTitle }}</div>
        <p class="mt-0.5 max-w-xl text-[11px] leading-snug text-slate-500">{{ panelSubtitle }}</p>
      </div>
      <div v-if="saved" class="text-[11px] font-semibold text-emerald-700">已保存</div>
    </div>

    <div class="mt-3 space-y-4">
      <!-- 基本设置 -->
      <div v-if="section === 'basic'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">标识与限额</div>
        <p class="mt-0.5 text-[11px] text-slate-500">名称与 DriverKey 用于列表展示；DriverKey 决定加载哪个 PSP 驱动；限额单位为分。</p>
        <div class="mt-2.5 grid gap-2.5 sm:grid-cols-2">
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
            通道名称
            <input v-model.trim="model.name" autocomplete="off" class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm" />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
            DriverKey（PSP 驱动）
            <input
              v-model.trim="model.driver_key"
              autocomplete="off"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
            />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-2">
            路由权重（0–100）
            <input
              v-model.number="model.weight"
              type="number"
              min="0"
              max="100"
              class="max-w-[12rem] rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
            单笔最小金额（分）
            <input
              v-model.number="model.min_amount"
              type="number"
              min="0"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
            单笔最大金额（分）
            <input
              v-model.number="model.max_amount"
              type="number"
              min="0"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            />
          </label>
        </div>
      </div>

      <div v-if="section === 'basic'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">运行状态</div>
        <p class="mt-0.5 text-[11px] text-slate-500">停用后不参与新路由；熔断用于临时屏蔽异常通道。</p>
        <div class="mt-2.5 grid gap-2.5 sm:grid-cols-2">
          <label
            class="flex items-center justify-between gap-3 rounded-md border border-slate-200/80 bg-white px-2.5 py-2"
          >
            <div>
              <div class="text-[11px] font-medium text-slate-700">启用通道</div>
              <p class="mt-0.5 text-[10px] text-slate-500">关闭后产品与路由不会选中该通道。</p>
            </div>
            <input v-model="model.enabled" type="checkbox" class="h-4 w-4 shrink-0 rounded border-slate-300 text-slate-900" />
          </label>
          <label
            class="flex items-center justify-between gap-3 rounded-md border border-slate-200/80 bg-white px-2.5 py-2"
          >
            <div>
              <div class="text-[11px] font-medium text-slate-700">熔断</div>
              <p class="mt-0.5 text-[10px] text-slate-500">开启后快速失败，用于通道异常时保护。</p>
            </div>
            <input v-model="model.fuse_enabled" type="checkbox" class="h-4 w-4 shrink-0 rounded border-slate-300 text-slate-900" />
          </label>
        </div>
      </div>

      <!-- 通道对接 -->
      <div v-if="section === 'channelConfig'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">通道对接配置（JSON）</div>
        <p class="mt-0.5 text-[11px] text-slate-500">
          整段文本存入数据库；保存时校验是否为合法 JSON。结构由你方与通道(PSP)约定，读配置时再解析即可。
        </p>
        <label class="mt-2.5 grid gap-1 text-[11px] font-medium text-slate-600">
          <span>配置内容</span>
          <textarea
            v-model="channelConfigJsonText"
            spellcheck="false"
            rows="18"
            class="min-h-[14rem] w-full resize-y rounded-md border border-slate-200 bg-white px-2.5 py-2 font-mono text-[11px] leading-relaxed"
            placeholder='{}'
          />
        </label>
        <p class="mt-1.5 text-[10px] text-slate-500">
          可为对象、数组等任意 JSON；留空表示不存配置。服务端与接口同样会做 JSON 校验。
        </p>
      </div>

      <!-- 费率与能力 -->
      <div v-if="section === 'rates'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">通道能力</div>
        <p class="mt-0.5 text-[11px] text-slate-500">与产品绑定、商户授权配合使用；关闭后代付相关费率可保留但不生效。</p>
        <div class="mt-2.5 grid gap-2.5 sm:grid-cols-2">
          <label
            class="flex items-center justify-between gap-3 rounded-md border border-slate-200/80 bg-white px-2.5 py-2"
          >
            <div>
              <div class="text-[11px] font-medium text-slate-700">支持代收</div>
              <p class="mt-0.5 text-[10px] text-slate-500">可参与代收产品路由。</p>
            </div>
            <input v-model="model.supports_payin" type="checkbox" class="h-4 w-4 shrink-0 rounded border-slate-300 text-slate-900" />
          </label>
          <label
            class="flex items-center justify-between gap-3 rounded-md border border-slate-200/80 bg-white px-2.5 py-2"
          >
            <div>
              <div class="text-[11px] font-medium text-slate-700">支持代付</div>
              <p class="mt-0.5 text-[10px] text-slate-500">可参与代付出款路由。</p>
            </div>
            <input v-model="model.supports_payout" type="checkbox" class="h-4 w-4 shrink-0 rounded border-slate-300 text-slate-900" />
          </label>
        </div>
      </div>

      <div v-if="section === 'rates'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">通道代收成本</div>
        <p class="mt-0.5 text-[11px] text-slate-500">平台相对通道的代收成本；比例按百分数填写（保存为万分比整数）。</p>
        <label class="mt-2.5 grid max-w-md gap-0.5 text-[11px] font-medium text-slate-600">
          {{ LABEL_CHANNEL_PAYIN_RATE }}
          <input
            :value="bpsToPercentInputValue(model.channel_payin_rate_bps)"
            type="number"
            min="0"
            step="0.01"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            @input="onChannelPayinPercentInput($event)"
          />
        </label>
      </div>

      <div v-if="section === 'rates'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">通道代付成本</div>
        <p class="mt-0.5 text-[11px] text-slate-500">需开启「支持代付」后用于通道成本核算。</p>
        <div v-if="!model.supports_payout" class="mt-2.5 rounded-lg border border-dashed border-slate-200 bg-white/60 px-3 py-4 text-center text-[11px] text-slate-500">
          当前未开启代付，可先打开「支持代付」再配置下方字段。
        </div>
        <div v-else class="mt-2.5 grid gap-2.5 sm:grid-cols-12">
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-4">
            {{ LABEL_CHANNEL_PAYOUT_RATE }}
            <input
            :value="bpsToPercentInputValue(model.channel_payout_rate_bps)"
            type="number"
            min="0"
            step="0.01"
            class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
            @input="onChannelPayoutPercentInput($event)"
          />
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-4">
            {{ LABEL_CHANNEL_PAYOUT_FEE_MODE }}
            <select v-model.number="model.channel_payout_fee_mode" class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm">
              <option v-for="opt in FEE_MODE_SELECT_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </label>
          <label class="grid gap-0.5 text-[11px] font-medium text-slate-600 sm:col-span-4">
            {{ LABEL_CHANNEL_PAYOUT_FIXED }}
            <input
              v-model.number="model.channel_payout_fixed_fee"
              type="number"
              min="0"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm tabular-nums"
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
        {{ saving ? '保存中...' : '保存配置' }}
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
import { computed, ref, watch } from 'vue'

import {
  FEE_MODE_SELECT_OPTIONS,
  LABEL_CHANNEL_PAYIN_RATE,
  LABEL_CHANNEL_PAYOUT_FEE_MODE,
  LABEL_CHANNEL_PAYOUT_FIXED,
  LABEL_CHANNEL_PAYOUT_RATE,
} from '../../../lib/feeSemantics'
import { bpsToPercentInputValue, percentToBps } from '../../../lib/ratePercent'
import type { AdminChannel } from './types'

export type ChannelFormSection = 'basic' | 'channelConfig' | 'rates'

const model = defineModel<AdminChannel>({ required: true })

const props = withDefaults(
  defineProps<{
    section: ChannelFormSection
    saving: boolean
    saved: boolean
    error: string
    canSave: boolean
    embedded?: boolean
    hideFooterActions?: boolean
  }>(),
  { embedded: false, hideFooterActions: false },
)

const channelConfigJsonText = ref('')

function syncChannelConfigJsonFromModel() {
  channelConfigJsonText.value = model.value.channel_config ?? ''
}

watch(
  () => model.value.channel_config,
  () => {
    if (props.section !== 'channelConfig') return
    syncChannelConfigJsonFromModel()
  },
  { immediate: true },
)

/** 将通道对接 JSON 写回 model；成功返回 null，失败返回错误文案（供保存前校验）。 */
function applyChannelConfigJsonToModel(): string | null {
  if (props.section !== 'channelConfig') return null
  try {
    const raw = channelConfigJsonText.value.trim()
    if (raw === '') {
      model.value.channel_config = ''
      return null
    }
    const parsed: unknown = JSON.parse(raw)
    model.value.channel_config = JSON.stringify(parsed, null, 2)
    syncChannelConfigJsonFromModel()
    return null
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    return `通道对接 JSON 无法解析：${msg}`
  }
}

defineExpose({ applyChannelConfigJsonToModel })

function onChannelPayinPercentInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value
  const n = parseFloat(raw)
  model.value.channel_payin_rate_bps = Number.isFinite(n) ? percentToBps(n) : 0
}

function onChannelPayoutPercentInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value
  const n = parseFloat(raw)
  model.value.channel_payout_rate_bps = Number.isFinite(n) ? percentToBps(n) : 0
}

defineEmits<{
  save: []
  reset: []
}>()

const panelTitle = computed(() => {
  switch (props.section) {
    case 'basic':
      return '基本设置'
    case 'channelConfig':
      return '通道对接'
    case 'rates':
      return '费率与能力'
    default:
      return '通道'
  }
})

const panelSubtitle = computed(() => {
  const id = model.value.id
  const hint = id ? `通道 #${id}` : '新建通道'
  switch (props.section) {
    case 'basic':
      return `${hint} · 名称、类型、限额与运行开关。`
    case 'channelConfig':
      return `${hint} · 通道对接 JSON（API、商户号、密钥与私钥）。`
    case 'rates':
      return `${hint} · 代收/代付能力与通道费率参数。`
    default:
      return hint
  }
})
</script>
