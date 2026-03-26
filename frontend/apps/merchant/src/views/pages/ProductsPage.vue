<template>
  <div class="space-y-6">
    <PageHeader title="我的产品" description="仅展示管理台已为当前商户开通的产品" />

    <div v-if="loading" class="rounded-2xl border border-slate-200/90 bg-white px-4 py-8 text-center text-sm text-slate-500">
      加载中…
    </div>

    <div v-else-if="products.length === 0" class="rounded-2xl border border-slate-200/90 bg-white px-4 py-8 text-center text-sm text-slate-500">
      暂无已开通产品
    </div>

    <template v-else>
      <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
        <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">代收产品</div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[540px] text-left text-sm">
            <thead class="border-b border-slate-100 bg-white text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th class="whitespace-nowrap px-4 py-3">产品名称</th>
                <th class="whitespace-nowrap px-4 py-3">费率</th>
                <th class="whitespace-nowrap px-4 py-3">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="payinProducts.length === 0">
                <td class="px-4 py-6 text-center text-slate-500" colspan="3">暂无已开通代收产品</td>
              </tr>
              <tr v-for="item in payinProducts" v-else :key="`payin-${item.product_id}`" class="hover:bg-slate-50/80">
                <td class="px-4 py-3 font-medium text-slate-900">{{ item.product_name || item.product_code || `代收产品#${item.product_id}` }}</td>
                <td class="px-4 py-3 text-slate-700">{{ formatRate(item) }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold"
                    :class="item.enabled ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ item.enabled ? '启用' : '停用' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
        <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">代付产品</div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[540px] text-left text-sm">
            <thead class="border-b border-slate-100 bg-white text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th class="whitespace-nowrap px-4 py-3">产品名称</th>
                <th class="whitespace-nowrap px-4 py-3">费率</th>
                <th class="whitespace-nowrap px-4 py-3">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="payoutProducts.length === 0">
                <td class="px-4 py-6 text-center text-slate-500" colspan="3">暂无已开通代付产品</td>
              </tr>
              <tr v-for="item in payoutProducts" v-else :key="`payout-${item.product_id}`" class="hover:bg-slate-50/80">
                <td class="px-4 py-3 font-medium text-slate-900">{{ item.product_name || item.product_code || `代付产品#${item.product_id}` }}</td>
                <td class="px-4 py-3 text-slate-700">{{ formatRate(item) }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold"
                    :class="item.enabled ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ item.enabled ? '启用' : '停用' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <ErrorCallout v-if="error" :message="error" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchMerchantOpenedProducts } from '@/api/console'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { merchantDisplaySettings } from '@/lib/displaySettings'
import type { MerchantOpenedProductItem } from '@/types/merchant.api'

const loading = ref(true)
const error = ref('')
const products = ref<MerchantOpenedProductItem[]>([])
const currencySymbol = computed(() => merchantDisplaySettings.value.currency_symbol || '¥')
const payinProducts = computed(() => products.value.filter((item) => item.product_type === 'payin'))
const payoutProducts = computed(() => products.value.filter((item) => item.product_type === 'payout'))

onMounted(async () => {
  try {
    const resp = await fetchMerchantOpenedProducts()
    products.value = resp.products || []
  } catch {
    error.value = '产品列表加载失败，请稍后重试。'
  } finally {
    loading.value = false
  }
})

function formatRate(item: MerchantOpenedProductItem): string {
  const hasRate = item.fee_rate_bps !== undefined && item.fee_rate_bps !== null
  const hasFixed = (item.fee_fixed_amount || 0) > 0
  const rateText = hasRate ? formatBps(item.fee_rate_bps || 0) : '-'
  const fixedText = `${formatMoney(item.fee_fixed_amount || 0)}/笔`

  if (item.product_type !== 'payout') {
    return rateText
  }
  if (item.fee_mode === 2) {
    return hasFixed ? fixedText : '-'
  }
  if (item.fee_mode === 3) {
    if (hasFixed && hasRate) return `${fixedText} + ${rateText}`
    if (hasFixed) return fixedText
    if (hasRate) return rateText
    return '-'
  }
  return rateText
}

function formatBps(v: number): string {
  return `${(v / 100).toFixed(2)}%`
}

function formatCent(v: number): string {
  return (v / 100).toFixed(2)
}

function formatMoney(v: number): string {
  return `${currencySymbol.value}${formatCent(v)}`
}
</script>
