import { computed, ref } from 'vue'
import { adminGet } from './adminApi'

export type AdminDisplaySettings = {
  country_code: string
  currency_code: string
  currency_symbol: string
  fiat_to_usdt_rate: number
}

const settings = ref<AdminDisplaySettings>({
  country_code: 'CN',
  currency_code: 'CNY',
  currency_symbol: '¥',
  fiat_to_usdt_rate: 7.2,
})

export const adminDisplaySettings = computed(() => settings.value)

export async function loadAdminDisplaySettings() {
  const r = await adminGet<AdminDisplaySettings>('/v1/admin/display_settings')
  settings.value = {
    country_code: r.country_code || 'CN',
    currency_code: r.currency_code || 'CNY',
    currency_symbol: r.currency_symbol || '¥',
    fiat_to_usdt_rate: r.fiat_to_usdt_rate > 0 ? r.fiat_to_usdt_rate : 7.2,
  }
}

export function applyAdminDisplaySettings(next: AdminDisplaySettings) {
  settings.value = {
    country_code: next.country_code || 'CN',
    currency_code: next.currency_code || 'CNY',
    currency_symbol: next.currency_symbol || '¥',
    fiat_to_usdt_rate: next.fiat_to_usdt_rate > 0 ? next.fiat_to_usdt_rate : 7.2,
  }
}

export function formatAdminMoney(cents: number): string {
  if (!Number.isFinite(cents)) return '—'
  const symbol = settings.value.currency_symbol || '¥'
  return `${symbol} ${(cents / 100).toFixed(2)}`
}
