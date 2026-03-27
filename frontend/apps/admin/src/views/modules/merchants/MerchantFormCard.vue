<template>
  <div class="w-full" :class="embedded ? '' : 'rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm'">
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div class="min-w-0">
        <div class="text-xs font-semibold text-slate-900">{{ panelTitle }}</div>
        <p class="mt-0.5 max-w-xl text-[11px] leading-snug text-slate-500">{{ panelSubtitle }}</p>
      </div>
      <div class="flex shrink-0 flex-wrap items-center gap-1.5">
        <div v-if="saved" class="text-[11px] font-semibold text-emerald-700">已保存</div>
        <template v-if="section === 'basic' && !isNew && model.merchant_id">
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-2 py-1 text-[11px] font-semibold text-slate-700 hover:bg-slate-50"
            @click="$emit('toggle-lock')"
          >
            {{ lockLabel }}
          </button>
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-2 py-1 text-[11px] font-semibold text-slate-700 hover:bg-slate-50"
            @click="$emit('reset-password')"
          >
            重置密码
          </button>
        </template>
      </div>
    </div>

    <div class="mt-3 space-y-4">
      <!-- 账户 -->
      <div v-if="section === 'basic'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
          <div class="text-xs font-semibold text-slate-800">账户信息</div>
          <p class="mt-0.5 text-[11px] text-slate-500">邮箱创建后不可在此修改。</p>
          <div class="mt-2.5 grid gap-2.5 sm:grid-cols-2">
            <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
              商户 ID
              <input
                v-model.trim="model.merchant_id"
                type="text"
                autocomplete="off"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
                :disabled="!isNew"
              />
            </label>
            <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
              登录邮箱
              <input
                v-model.trim="model.email"
                type="email"
                autocomplete="off"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm"
                :disabled="!isNew"
              />
            </label>
            <p v-if="!isNew" class="text-[11px] leading-snug text-slate-500 sm:col-span-2">
              已入驻商户邮箱变更请走内部流程。
            </p>
            <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
              账户状态
              <select v-model.number="model.status" class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm">
                <option :value="1">启用</option>
                <option :value="0">锁定</option>
              </select>
            </label>
          </div>
        </div>

      <!-- 对接 + 通知 + IP + 重置密钥 -->
      <div v-if="section === 'api'" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
        <div class="text-xs font-semibold text-slate-800">开放 API：对接、通知与访问控制</div>
        <p class="mt-0.5 text-[11px] text-slate-500">
          凭证与 IP 白名单控制谁能调开放接口；Notify / Return 用于支付结果与页面回跳。
        </p>

        <template v-if="!isNew && merchantInfo">
          <div class="mt-2.5 grid gap-2 sm:grid-cols-2">
            <div class="rounded-md border border-slate-200/80 bg-white px-2.5 py-1.5">
              <div class="text-[11px] font-medium text-slate-600">AppID</div>
              <div class="mt-0.5 flex items-center gap-1">
                <span class="min-w-0 flex-1 font-mono text-xs text-slate-900">{{ merchantInfo.app_id || '—' }}</span>
                <button
                  type="button"
                  class="inline-flex shrink-0 rounded p-0.5 text-slate-400 hover:text-indigo-600"
                  title="复制 AppID"
                  aria-label="复制 AppID"
                  @click="copyText(merchantInfo.app_id, 'AppID')"
                >
                  <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                    />
                  </svg>
                </button>
              </div>
            </div>
            <div class="rounded-md border border-slate-200/80 bg-white px-2.5 py-1.5">
              <div class="text-[11px] font-medium text-slate-600">API 密钥（脱敏）</div>
              <div class="mt-0.5 flex items-center gap-1">
                <span class="min-w-0 flex-1 font-mono text-xs text-slate-900">{{ maskedSecret(merchantInfo.app_secret) }}</span>
                <button
                  type="button"
                  class="inline-flex shrink-0 rounded p-0.5 text-slate-400 hover:text-indigo-600"
                  title="复制密钥"
                  aria-label="复制密钥"
                  @click="copyText(merchantInfo.app_secret, '密钥')"
                >
                  <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
          <div class="mt-2.5 flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="rounded-md border border-amber-200/90 bg-amber-50 px-2.5 py-1 text-[11px] font-semibold text-amber-900 hover:bg-amber-100"
              @click="$emit('reset-secret')"
            >
              重置 API 密钥
            </button>
            <span class="text-[10px] text-slate-500">旧密钥立即失效，请通知商户更新。</span>
          </div>
        </template>
        <p v-else class="mt-2.5 text-[11px] text-slate-500">保存商户后将生成 AppID 与密钥，并可在下方维护通知地址与 IP 白名单。</p>

        <div class="mt-3 border-t border-slate-200/80 pt-3">
          <div class="text-[11px] font-semibold text-slate-700">异步通知与同步跳转</div>
          <p class="mt-0.5 text-[11px] text-slate-500">服务端回调（Notify）与支付完成回跳（Return）。</p>
          <div class="mt-2 grid gap-2">
            <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
              Notify URL
              <input
                v-model.trim="model.notify_url"
                type="url"
                placeholder="https://"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
              />
            </label>
            <label class="grid gap-0.5 text-[11px] font-medium text-slate-600">
              Return URL
              <input
                v-model.trim="model.return_url"
                type="url"
                placeholder="https://"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
              />
            </label>
          </div>
        </div>

        <div class="mt-3 border-t border-slate-200/80 pt-3">
          <div class="text-[11px] font-semibold text-slate-700">API IP 白名单</div>
          <p class="mt-0.5 text-[11px] text-slate-500">限制开放 API 请求来源；变更前请与商户同步联调环境。</p>
          <label class="mt-2 grid gap-0.5 text-[11px] font-medium text-slate-600">
            白名单列表
            <textarea
              v-model="model.ip_whitelist"
              rows="5"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-[11px] leading-relaxed"
              placeholder="每行一个 IP 或网段；留空以网关策略为准"
            />
          </label>
        </div>
      </div>

      <!-- 财务：余额 + USDT 地址 -->
      <template v-if="section === 'finance'">
        <div v-if="!isNew && merchantInfo" class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
          <div class="text-xs font-semibold text-slate-800">账户余额</div>
          <p class="mt-0.5 text-[11px] text-slate-500">列表与开放接口共用数据；充值、提现、划转请用工具栏入口。</p>
          <div class="mt-2.5 grid gap-2 sm:grid-cols-2">
            <div class="rounded-md border border-slate-200/80 bg-white px-2.5 py-1.5">
              <div class="text-[11px] font-medium text-slate-600">代收余额</div>
              <div class="mt-0.5 font-mono text-xs font-semibold tabular-nums text-slate-900">
                {{ formatMoney(merchantInfo.payin_balance) }}
              </div>
            </div>
            <div class="rounded-md border border-slate-200/80 bg-white px-2.5 py-1.5">
              <div class="text-[11px] font-medium text-slate-600">可用余额（代付）</div>
              <div class="mt-0.5 font-mono text-xs font-semibold tabular-nums text-slate-900">
                {{ formatMoney(merchantInfo.available_balance ?? 0) }}
              </div>
            </div>
          </div>
        </div>
        <p v-else class="rounded-lg border border-dashed border-slate-200 bg-slate-50/40 px-3 py-3 text-[11px] text-slate-500">
          保存商户后将显示代收与代付可用余额。
        </p>

        <div class="space-y-2">
          <div class="rounded-xl border border-slate-200/90 bg-slate-50/40 p-3.5">
            <div class="text-xs font-semibold text-slate-800">提现 USDT 收款</div>
            <p class="mt-0.5 text-[11px] text-slate-500">法币提现链上打款地址（与结算模块一致）。</p>
            <label class="mt-2.5 grid gap-0.5 text-[11px] font-medium text-slate-600">
              链上地址
              <input
                v-model.trim="model.withdraw_usdt_address"
                type="text"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 font-mono text-sm"
                placeholder="如 TRC20 地址"
              />
            </label>
          </div>
          <p class="text-[11px] leading-snug text-amber-800/90">请与商户确认主网（如 TRC20），错误地址可能导致资金损失。</p>
        </div>
      </template>
    </div>

    <div v-if="error" class="mt-3 rounded-md border border-rose-200 bg-rose-50 px-2.5 py-2 text-xs text-rose-800">
      {{ error }}
    </div>

    <div v-if="!hideFooterActions" class="mt-4 flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white shadow-sm disabled:opacity-40"
        :disabled="saving || !canSave"
        @click="$emit('save')"
      >
        {{ saving ? '保存中...' : '保存配置' }}
      </button>
      <button
        type="button"
        class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 shadow-sm"
        @click="$emit('reset')"
      >
        重置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { formatAdminMoney } from '../../../lib/displaySettings'
import { useUiToast } from '../../../composables/useUiToast'

import type { AdminMerchantInfo, MerchantForm } from './types'

const model = defineModel<MerchantForm>({ required: true })

const props = withDefaults(
  defineProps<{
    /** 抽屉内分 Tab：基本资料 / API / 财务 */
    section: 'basic' | 'api' | 'finance'
    isNew: boolean
    saving: boolean
    saved: boolean
    error: string
    canSave: boolean
    statusForLock: number
    /** 仅编辑态：用于展示 AppID、密钥脱敏与余额 */
    merchantInfo?: AdminMerchantInfo | null
    /** 嵌入 Tab 面板时不画外框 */
    embedded?: boolean
    /** 抽屉底部统一放保存/关闭时隐藏表单底部按钮 */
    hideFooterActions?: boolean
  }>(),
  { embedded: false, hideFooterActions: false, merchantInfo: null },
)

defineEmits<{
  save: []
  reset: []
  'toggle-lock': []
  'reset-secret': []
  'reset-password': []
}>()

const toast = useUiToast()

const lockLabel = computed(() => (props.statusForLock === 1 ? '锁定账户' : '解除锁定'))

const panelTitle = computed(() => {
  if (props.section === 'basic') return props.isNew ? '新建商户' : '基本资料'
  if (props.section === 'api') return 'API 对接'
  return '财务'
})

const panelSubtitle = computed(() => {
  if (props.section === 'basic') {
    return props.isNew
      ? '填写账户并保存后，可在「API 对接」「财务」与产品 Tab 继续配置。'
      : '商户 ID 与邮箱；状态与登录安全（锁定、重置密码）在本页。'
  }
  if (props.section === 'api') {
    return props.isNew
      ? '通知地址与 IP 可在创建时填写，保存后系统分配 AppID 与密钥。'
      : '开放接口凭证、回调、IP 白名单与密钥轮换。'
  }
  return props.isNew
    ? '可预填链上收款地址；余额在商户创建成功后显示。'
    : '代收/代付余额与提现 USDT 收款地址。'
})

function formatMoney(v: number) {
  return formatAdminMoney(v)
}

function maskedSecret(secret: string) {
  const s = String(secret || '')
  if (s.length <= 8) return s || '—'
  return `${s.slice(0, 4)}****${s.slice(-4)}`
}

async function copyText(value: string, label: string) {
  const text = String(value || '').trim()
  if (!text) {
    toast.error(`${label}为空，无法复制`)
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    toast.success(`${label}已复制`)
  } catch {
    toast.error(`复制${label}失败`)
  }
}
</script>
