<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统管理</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        管理全局展示参数（国家、货币、汇率）与后台登录安全开关。改动保存后会立即影响管理台的金额显示与登录策略。
      </p>
    </div>

    <div
      v-if="error"
      class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
    >
      {{ error }}
    </div>

    <div v-if="loading" class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
      <div class="animate-pulse space-y-3">
        <div class="h-4 w-36 rounded bg-slate-200" />
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div class="h-16 rounded-xl bg-slate-100" />
          <div class="h-16 rounded-xl bg-slate-100" />
          <div class="h-16 rounded-xl bg-slate-100" />
          <div class="h-16 rounded-xl bg-slate-100" />
        </div>
      </div>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-12">
      <section class="lg:col-span-8 rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
        <div class="mb-4 flex items-center justify-between gap-3 border-b border-slate-100 pb-3">
          <div>
            <div class="text-sm font-semibold text-slate-900">全局展示配置</div>
            <div class="mt-1 text-xs text-slate-500">用于金额文案和汇率计算展示。</div>
          </div>
          <span class="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-semibold text-slate-600">MVP</span>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <label class="grid gap-1 text-sm sm:col-span-2">
            <span class="text-xs font-medium text-slate-500">新建商户数字 ID 起始值</span>
            <input
              v-model.number="merchantNumericIdStart"
              type="number"
              min="1"
              max="9999999999"
              step="1"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="5000000000"
            />
            <span class="text-[11px] font-normal text-slate-500">
              自动分配 10 位商户号时的下限（含）。若序列表已超过此值，下一个号仍连续递增；仅在新号需「抬高」时改大并保存。
            </span>
          </label>
          <label class="grid gap-1 text-sm">
            <span class="text-xs font-medium text-slate-500">国家代码</span>
            <input
              v-model.trim="countryCode"
              type="text"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="CN"
            />
          </label>
          <label class="grid gap-1 text-sm">
            <span class="text-xs font-medium text-slate-500">货币代码</span>
            <input
              v-model.trim="currencyCode"
              type="text"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="CNY"
            />
          </label>
          <label class="grid gap-1 text-sm">
            <span class="text-xs font-medium text-slate-500">货币符号</span>
            <input
              v-model.trim="currencySymbol"
              type="text"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="¥"
            />
          </label>
          <label class="grid gap-1 text-sm">
            <span class="text-xs font-medium text-slate-500">法币 -> USDT 汇率</span>
            <input
              v-model.number="fiatToUsdtRate"
              type="number"
              min="0.000001"
              step="0.000001"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="7.200000"
            />
          </label>
        </div>

        <div class="mt-4 flex items-center gap-3 border-t border-slate-100 pt-4">
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="saving || !canSave"
            @click="saveDisplaySettings"
          >
            {{ saving ? '保存中…' : '保存配置' }}
          </button>
          <span v-if="saved" class="text-xs font-medium text-emerald-700">已保存</span>
          <span v-if="!canSave" class="text-xs text-amber-700">请填写完整且有效的配置</span>
        </div>
      </section>

      <aside class="space-y-4 lg:col-span-4">
        <section class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="text-sm font-semibold text-slate-900">安全策略</div>
          <p class="mt-1 text-xs text-slate-500">已绑定谷歌验证器的管理员登录时必须输入动态码；未绑定前仅可进入绑定流程。</p>
          <label class="mt-3 flex cursor-pointer items-start gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-700">
            <input v-model="adminMfaEnabled" type="checkbox" class="mt-0.5 h-4 w-4 rounded border-slate-300" />
            <span>保留项：后台 MFA 开关（当前登录校验以账号是否已绑定验证器为准，不依赖此项）</span>
          </label>
        </section>

        <section class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
          <div class="text-sm font-semibold text-slate-900">实时预览</div>
          <div class="mt-3 space-y-2 text-xs text-slate-600">
            <div class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2">
              <span>国家 / 货币</span>
              <span class="font-mono font-semibold text-slate-800">{{ countryCode.toUpperCase() }} / {{ currencyCode.toUpperCase() }}</span>
            </div>
            <div class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2">
              <span>金额样式</span>
              <span class="font-semibold text-slate-800">{{ currencySymbol || '¥' }} 12,345.67</span>
            </div>
            <div class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2">
              <span>汇率</span>
              <span class="font-mono font-semibold text-slate-800">{{ ratePreview }}</span>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border border-slate-100 bg-slate-50/80 px-4 py-3 text-sm text-slate-600">
          管理员账号与角色请在
          <RouterLink to="/rbac/admin-users" class="font-semibold text-indigo-600 underline">后台用户</RouterLink>
          与
          <RouterLink to="/rbac/roles" class="font-semibold text-indigo-600 underline">角色与授权</RouterLink>
          中维护。
        </section>
      </aside>
    </div>

    <div class="rounded-2xl border border-slate-100 bg-slate-50/80 px-4 py-3 text-sm text-slate-600">
      后续可继续扩展操作审计、参数中心、账号风险策略等系统治理能力。
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { adminGet } from '../../../lib/adminApi'
import { adminPut } from '../../../lib/adminApi'
import { useUiToast } from '../../../composables/useUiToast'
import { applyAdminDisplaySettings } from '../../../lib/displaySettings'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const toast = useUiToast()

const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref('')
const countryCode = ref('CN')
const currencyCode = ref('CNY')
const currencySymbol = ref('¥')
const fiatToUsdtRate = ref(7.2)
const adminMfaEnabled = ref(false)
const merchantNumericIdStart = ref(5_000_000_000)

const canSave = computed(() => {
  const floor = Number(merchantNumericIdStart.value)
  const floorOk =
    Number.isFinite(floor) && Number.isInteger(floor) && floor >= 1 && floor <= 9999999999
  return (
    countryCode.value.trim().length > 0 &&
    currencyCode.value.trim().length > 0 &&
    currencySymbol.value.trim().length > 0 &&
    Number.isFinite(fiatToUsdtRate.value) &&
    fiatToUsdtRate.value > 0 &&
    floorOk
  )
})

const ratePreview = computed(() => {
  const v = Number(fiatToUsdtRate.value)
  if (!Number.isFinite(v) || v <= 0) return '—'
  return `${v.toFixed(6)}`
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const ds = await adminGet<{
      country_code: string
      currency_code: string
      currency_symbol: string
      fiat_to_usdt_rate: number
      admin_mfa_enabled: number
      merchant_numeric_id_start?: number
    }>('/v1/admin/display_settings')
    countryCode.value = ds.country_code || 'CN'
    currencyCode.value = ds.currency_code || 'CNY'
    currencySymbol.value = ds.currency_symbol || '¥'
    fiatToUsdtRate.value = ds.fiat_to_usdt_rate > 0 ? ds.fiat_to_usdt_rate : 7.2
    adminMfaEnabled.value = Number(ds.admin_mfa_enabled || 0) === 1
    const st = Number(ds.merchant_numeric_id_start ?? 5_000_000_000)
    merchantNumericIdStart.value =
      Number.isFinite(st) && st >= 1 && st <= 9999999999 ? Math.floor(st) : 5_000_000_000
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载展示配置失败：${msg}`)
  } finally {
    loading.value = false
  }
}

async function saveDisplaySettings() {
  if (!canSave.value) return
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const r = await adminPut<{
      country_code: string
      currency_code: string
      currency_symbol: string
      fiat_to_usdt_rate: number
      admin_mfa_enabled: number
      merchant_numeric_id_start?: number
    }>('/v1/admin/display_settings', {
      country_code: countryCode.value.trim().toUpperCase(),
      currency_code: currencyCode.value.trim().toUpperCase(),
      currency_symbol: currencySymbol.value.trim(),
      fiat_to_usdt_rate: fiatToUsdtRate.value,
      admin_mfa_enabled: adminMfaEnabled.value ? 1 : 0,
      merchant_numeric_id_start: Math.floor(Number(merchantNumericIdStart.value)),
    })
    applyAdminDisplaySettings(r)
    countryCode.value = r.country_code
    currencyCode.value = r.currency_code
    currencySymbol.value = r.currency_symbol
    fiatToUsdtRate.value = r.fiat_to_usdt_rate
    adminMfaEnabled.value = Number(r.admin_mfa_enabled || 0) === 1
    const rst = Number(r.merchant_numeric_id_start ?? 5_000_000_000)
    merchantNumericIdStart.value =
      Number.isFinite(rst) && rst >= 1 && rst <= 9999999999 ? Math.floor(rst) : 5_000_000_000
    saved.value = true
    toast.success('展示配置已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存展示配置失败：${msg}`)
  } finally {
    saving.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void load()
  if (registerRefresh) unregister = registerRefresh(() => void load())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
