<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统管理</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：提供平台<strong>全局展示配置</strong>（国家/货币）。管理员列表与角色分配请在
        <RouterLink to="/rbac/admin-users" class="font-semibold text-indigo-600 underline">权限与安全 → 后台用户</RouterLink>
        中维护。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-4 shadow-sm">
      <div class="mb-3 text-sm font-semibold text-slate-800">全局展示配置</div>
      <div class="grid gap-3 sm:grid-cols-4">
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">国家代码</span>
          <input v-model.trim="countryCode" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">货币代码</span>
          <input v-model.trim="currencyCode" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">货币符号</span>
          <input v-model.trim="currencySymbol" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <label class="grid gap-1 text-sm">
          <span class="text-xs font-medium text-slate-500">法币 -> USDT 汇率</span>
          <input v-model.number="fiatToUsdtRate" type="number" min="0.000001" step="0.000001" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
      </div>
      <div class="mt-3 flex items-center gap-2">
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
          :disabled="saving"
          @click="saveDisplaySettings"
        >
          {{ saving ? '保存中…' : '保存配置' }}
        </button>
        <span v-if="saved" class="text-xs text-emerald-700">已保存</span>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-100 bg-slate-50/80 px-4 py-3 text-sm text-slate-600">
      操作审计、参数中心、账号新建/改密等后续迭代；管理员与角色请在
      <RouterLink to="/rbac/admin-users" class="font-semibold text-indigo-600 underline">后台用户</RouterLink>
      与
      <RouterLink to="/rbac/roles" class="font-semibold text-indigo-600 underline">角色与授权</RouterLink>
      配置。
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { adminGet } from '../../../lib/adminApi'
import { adminPut } from '../../../lib/adminApi'
import { useUiToast } from '../../../composables/ui'
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

async function load() {
  loading.value = true
  error.value = ''
  try {
    const ds = await adminGet<{ country_code: string; currency_code: string; currency_symbol: string; fiat_to_usdt_rate: number }>('/v1/admin/display_settings')
    countryCode.value = ds.country_code || 'CN'
    currencyCode.value = ds.currency_code || 'CNY'
    currencySymbol.value = ds.currency_symbol || '¥'
    fiatToUsdtRate.value = ds.fiat_to_usdt_rate > 0 ? ds.fiat_to_usdt_rate : 7.2
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载展示配置失败：${msg}`)
  } finally {
    loading.value = false
  }
}

async function saveDisplaySettings() {
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const r = await adminPut<{ country_code: string; currency_code: string; currency_symbol: string; fiat_to_usdt_rate: number }>('/v1/admin/display_settings', {
      country_code: countryCode.value.trim().toUpperCase(),
      currency_code: currencyCode.value.trim().toUpperCase(),
      currency_symbol: currencySymbol.value.trim(),
      fiat_to_usdt_rate: fiatToUsdtRate.value,
    })
    applyAdminDisplaySettings(r)
    countryCode.value = r.country_code
    currencyCode.value = r.currency_code
    currencySymbol.value = r.currency_symbol
    fiatToUsdtRate.value = r.fiat_to_usdt_rate
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
