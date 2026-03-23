import { computed, ref } from 'vue'
import { merchantConsoleGet } from './http'

export type MerchantDisplaySettings = {
  country_code: string
  currency_code: string
  currency_symbol: string
}

const settings = ref<MerchantDisplaySettings>({
  country_code: 'CN',
  currency_code: 'CNY',
  currency_symbol: '¥',
})

export const merchantDisplaySettings = computed(() => settings.value)

export async function loadMerchantDisplaySettings() {
  const r = await merchantConsoleGet<MerchantDisplaySettings>('/v1/merchant/display_settings')
  settings.value = {
    country_code: r.country_code || 'CN',
    currency_code: r.currency_code || 'CNY',
    currency_symbol: r.currency_symbol || '¥',
  }
}

export function formatMerchantMoney(cents: number): string {
  const symbol = settings.value.currency_symbol || '¥'
  return `${symbol} ${(cents / 100).toFixed(2)}`
}
