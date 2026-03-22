<template>
  <div class="grid gap-4">
    <MerchantsHeader @new-merchant="newMerchant" @refresh="reload" />

    <div class="grid grid-cols-12 gap-4">
      <MerchantList
        :merchants="merchants"
        :loading="loading"
        :selected-id="selectedMerchantId"
        @select="selectMerchant"
      />

      <div class="col-span-12 md:col-span-8">
        <MerchantFormCard
          v-if="isNew"
          v-model="form"
          :is-new="true"
          :saving="saving"
          :saved="saved"
          :error="formError"
          :can-save="canSaveForm"
          :status-for-lock="form.status"
          @save="saveForm"
          @reset="resetForm"
        />

        <div
          v-else-if="selectedMerchant"
          class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm"
        >
          <div class="flex flex-wrap border-b border-slate-200 px-2 pt-3" role="tablist" aria-label="商户详情">
            <button
              v-for="tab in detailTabs"
              :key="tab.key"
              type="button"
              role="tab"
              :aria-selected="rightTab === tab.key"
              class="relative -mb-px border-b-2 px-3 pb-3 text-sm font-semibold transition md:px-4"
              :class="
                rightTab === tab.key
                  ? 'border-slate-900 text-slate-900'
                  : 'border-transparent text-slate-500 hover:text-slate-800'
              "
              @click="rightTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </div>

          <div v-show="rightTab === 'basic'" role="tabpanel">
            <MerchantFormCard
              v-model="form"
              :is-new="false"
              embedded
              :saving="saving"
              :saved="saved"
              :error="formError"
              :can-save="canSaveForm"
              :status-for-lock="selectedMerchant.status"
              @save="saveForm"
              @reset="resetForm"
              @toggle-lock="toggleLock"
              @reset-secret="resetSecret"
            />
          </div>

          <div v-show="rightTab === 'bindings_collect'" role="tabpanel">
            <MerchantPayProductsCard
              embedded
              :product-ids="selectedMerchant.pay_product_ids || []"
              :catalog="payProducts"
              :loading="loadingProducts"
              :saving="bindingSaving"
              :bind-error="bindError"
              @remove="removePayProduct"
              @add="addPayProduct"
            />
          </div>

          <div v-show="rightTab === 'bindings_payout'" role="tabpanel">
            <MerchantPayoutProductsCard
              embedded
              :product-ids="selectedMerchant.payout_product_ids || []"
              :catalog="payoutProducts"
              :loading="loadingPayoutProducts"
              :saving="bindingSaving"
              :bind-error="bindError"
              @remove="removePayoutProduct"
              @add="addPayoutProduct"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import MerchantFormCard from './MerchantFormCard.vue'
import MerchantList from './MerchantList.vue'
import MerchantPayProductsCard from './MerchantPayProductsCard.vue'
import MerchantPayoutProductsCard from './MerchantPayoutProductsCard.vue'
import MerchantsHeader from './MerchantsHeader.vue'
import type { AdminMerchantInfo, MerchantForm, PayProductRow } from './types'
import { emptyMerchantForm, merchantToForm } from './types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined

const loading = ref(false)
const loadingProducts = ref(false)
const loadingPayoutProducts = ref(false)
const saving = ref(false)
const bindingSaving = ref(false)
const saved = ref(false)
const formError = ref('')
const bindError = ref('')

const merchants = ref<AdminMerchantInfo[]>([])
const payProducts = ref<PayProductRow[]>([])
const payoutProducts = ref<PayProductRow[]>([])
const selectedMerchantId = ref<string | null>(null)
const rightTab = ref<'basic' | 'bindings_collect' | 'bindings_payout'>('basic')

const detailTabs = [
  { key: 'basic' as const, label: '基本信息' },
  { key: 'bindings_collect' as const, label: '代收产品' },
  { key: 'bindings_payout' as const, label: '代付产品' },
]

const form = ref<MerchantForm>(emptyMerchantForm())

const isNew = computed(() => selectedMerchantId.value === null)

const selectedMerchant = computed(() => {
  const id = selectedMerchantId.value
  if (!id) return null
  return merchants.value.find((m) => m.merchant_id === id) ?? null
})

const canSaveForm = computed(() => {
  if (isNew.value) return !!form.value.merchant_id?.trim()
  return true
})

function applySelectedToForm() {
  const m = selectedMerchant.value
  if (!m) return
  form.value = merchantToForm(m)
}

function resetForm() {
  saved.value = false
  formError.value = ''
  if (selectedMerchant.value) applySelectedToForm()
  else form.value = emptyMerchantForm()
}

function selectMerchant(merchantId: string) {
  selectedMerchantId.value = merchantId
  rightTab.value = 'basic'
  applySelectedToForm()
  saved.value = false
  formError.value = ''
}

function newMerchant() {
  selectedMerchantId.value = null
  rightTab.value = 'basic'
  form.value = emptyMerchantForm()
  saved.value = false
  formError.value = ''
}

async function loadPayProducts() {
  loadingProducts.value = true
  try {
    const res = await adminGet<{ products: PayProductRow[] }>('/v1/admin/pay_products')
    payProducts.value = res.products || []
  } catch {
    payProducts.value = []
  } finally {
    loadingProducts.value = false
  }
}

async function loadPayoutProducts() {
  loadingPayoutProducts.value = true
  try {
    const res = await adminGet<{ products: PayProductRow[] }>('/v1/admin/payout_products')
    payoutProducts.value = res.products || []
  } catch {
    payoutProducts.value = []
  } finally {
    loadingPayoutProducts.value = false
  }
}

async function reload() {
  formError.value = ''
  saved.value = false
  loading.value = true
  try {
    const data = await adminGet<{ merchants: AdminMerchantInfo[] }>('/v1/admin/merchants')
    merchants.value = data.merchants || []
    if (selectedMerchantId.value && merchants.value.some((m) => m.merchant_id === selectedMerchantId.value)) {
      applySelectedToForm()
    } else if (merchants.value.length > 0) {
      selectMerchant(merchants.value[0].merchant_id)
    } else {
      newMerchant()
    }
  } catch {
    formError.value = '网络错误'
  } finally {
    loading.value = false
  }
}

async function saveForm() {
  saving.value = true
  formError.value = ''
  saved.value = false
  try {
    if (isNew.value) {
      const resp = await adminPost<{ merchant: AdminMerchantInfo }>('/v1/admin/merchants', {
        merchant_id: form.value.merchant_id.trim(),
        api_secret: form.value.api_secret,
        default_collect_rate_bps: form.value.default_collect_rate_bps,
        default_payout_rate_bps: form.value.default_payout_rate_bps,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        pay_product_ids: [],
        payout_product_ids: [],
      })
      const row = resp.merchant
      merchants.value.push(row)
      merchants.value.sort((a, b) => a.merchant_id.localeCompare(b.merchant_id))
      selectedMerchantId.value = row.merchant_id
      rightTab.value = 'basic'
      form.value = merchantToForm(row)
    } else {
      const mid = selectedMerchant.value!.merchant_id
      const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(mid)}`, {
        status: form.value.status,
        default_collect_rate_bps: form.value.default_collect_rate_bps,
        default_payout_rate_bps: form.value.default_payout_rate_bps,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        pay_product_ids: selectedMerchant.value!.pay_product_ids || [],
        payout_product_ids: selectedMerchant.value!.payout_product_ids || [],
      })
      const row = resp.merchant
      const idx = merchants.value.findIndex((m) => m.merchant_id === row.merchant_id)
      if (idx >= 0) merchants.value[idx] = row
      form.value = merchantToForm(row)
    }
    saved.value = true
  } catch {
    formError.value = '网络错误'
  } finally {
    saving.value = false
  }
}

async function toggleLock() {
  const m = selectedMerchant.value
  if (!m) return
  formError.value = ''
  try {
    const target = m.status === 1 ? 0 : 1
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: target,
      default_collect_rate_bps: m.default_collect_rate_bps,
      default_payout_rate_bps: m.default_payout_rate_bps,
      notify_url: m.notify_url,
      return_url: m.return_url,
      ip_whitelist: m.ip_whitelist,
      pay_product_ids: m.pay_product_ids || [],
      payout_product_ids: m.payout_product_ids || [],
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
  } catch {
    formError.value = '更新状态失败'
  }
}

async function resetSecret() {
  const m = selectedMerchant.value
  if (!m) return
  formError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(
      `/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`,
      {
        reset_secret: true,
        status: m.status,
        default_collect_rate_bps: m.default_collect_rate_bps,
        default_payout_rate_bps: m.default_payout_rate_bps,
        notify_url: m.notify_url,
        return_url: m.return_url,
        ip_whitelist: m.ip_whitelist,
        pay_product_ids: m.pay_product_ids || [],
        payout_product_ids: m.payout_product_ids || [],
      },
    )
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
  } catch {
    formError.value = '重置密钥失败'
  }
}

async function persistPayProducts(newIds: number[]) {
  const m = selectedMerchant.value
  if (!m) return
  bindingSaving.value = true
  bindError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: m.status,
      default_collect_rate_bps: m.default_collect_rate_bps,
      default_payout_rate_bps: m.default_payout_rate_bps,
      notify_url: m.notify_url,
      return_url: m.return_url,
      ip_whitelist: m.ip_whitelist,
      pay_product_ids: newIds,
      payout_product_ids: m.payout_product_ids || [],
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
  } catch {
    bindError.value = '保存失败'
  } finally {
    bindingSaving.value = false
  }
}

async function persistPayoutProducts(newIds: number[]) {
  const m = selectedMerchant.value
  if (!m) return
  bindingSaving.value = true
  bindError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: m.status,
      default_collect_rate_bps: m.default_collect_rate_bps,
      default_payout_rate_bps: m.default_payout_rate_bps,
      notify_url: m.notify_url,
      return_url: m.return_url,
      ip_whitelist: m.ip_whitelist,
      pay_product_ids: m.pay_product_ids || [],
      payout_product_ids: newIds,
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
  } catch {
    bindError.value = '保存失败'
  } finally {
    bindingSaving.value = false
  }
}

function removePayProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m) return
  const cur = [...(m.pay_product_ids || [])]
  const next = cur.filter((id) => id !== productId)
  void persistPayProducts(next)
}

function addPayProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m || productId <= 0) return
  const cur = [...(m.pay_product_ids || [])]
  if (cur.includes(productId)) return
  cur.push(productId)
  void persistPayProducts(cur)
}

function removePayoutProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m) return
  const cur = [...(m.payout_product_ids || [])]
  const next = cur.filter((id) => id !== productId)
  void persistPayoutProducts(next)
}

function addPayoutProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m || productId <= 0) return
  const cur = [...(m.payout_product_ids || [])]
  if (cur.includes(productId)) return
  cur.push(productId)
  void persistPayoutProducts(cur)
}

let unregister: (() => void) | null = null
onMounted(() => {
  void loadPayProducts()
  void loadPayoutProducts()
  void reload()
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
