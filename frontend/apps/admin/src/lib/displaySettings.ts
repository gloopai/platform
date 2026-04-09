import { computed, ref } from 'vue'
import { adminGet } from './adminApi'

export type AdminDisplaySettings = {
  country_code: string
  currency_code: string
  currency_symbol: string
  fiat_to_usdt_rate: number
  /** 管理台标题与 MFA issuer；空则用浏览器默认短标题 */
  system_name: string
}

const DEFAULT_DOC_TITLE = '管理台'

const settings = ref<AdminDisplaySettings>({
  country_code: 'CN',
  currency_code: 'CNY',
  currency_symbol: '¥',
  fiat_to_usdt_rate: 7.2,
  system_name: '',
})

export const adminDisplaySettings = computed(() => settings.value)

export async function loadAdminDisplaySettings() {
  const r = await adminGet<AdminDisplaySettings>('/v1/admin/display_settings')
  settings.value = {
    country_code: r.country_code || 'CN',
    currency_code: r.currency_code || 'CNY',
    currency_symbol: r.currency_symbol || '¥',
    fiat_to_usdt_rate: r.fiat_to_usdt_rate > 0 ? r.fiat_to_usdt_rate : 7.2,
    system_name: (r.system_name || '').trim(),
  }
  applyAdminDocumentTitle(settings.value.system_name)
}

export function applyAdminDisplaySettings(next: AdminDisplaySettings) {
  settings.value = {
    country_code: next.country_code || 'CN',
    currency_code: next.currency_code || 'CNY',
    currency_symbol: next.currency_symbol || '¥',
    fiat_to_usdt_rate: next.fiat_to_usdt_rate > 0 ? next.fiat_to_usdt_rate : 7.2,
    system_name: (next.system_name || '').trim(),
  }
  applyAdminDocumentTitle(settings.value.system_name)
}

/** 根据系统设置更新 `document.title`（空则使用短默认标题） */
export function applyAdminDocumentTitle(systemName: string) {
  const t = systemName.trim()
  document.title = t || DEFAULT_DOC_TITLE
}

export function formatAdminMoney(cents: number): string {
  if (!Number.isFinite(cents)) return '—'
  const symbol = settings.value.currency_symbol || '¥'
  return `${symbol} ${(cents / 100).toFixed(2)}`
}
