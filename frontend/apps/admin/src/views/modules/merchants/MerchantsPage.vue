<template>
  <div class="grid gap-4">
    <MerchantsHeader @new-merchant="openNew" @refresh="reload" />

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
        <input
          v-model.trim="searchQuery"
          type="search"
          autocomplete="off"
          placeholder="搜索商户 ID…"
          class="w-full max-w-md rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
        />
        <label class="flex items-center gap-2 text-sm text-slate-600">
          <span class="text-slate-500">匹配</span>
          <span class="font-mono text-slate-900">{{ filteredMerchants.length }}</span>
          <span class="text-slate-500">条</span>
        </label>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">商户 ID</th>
              <th class="whitespace-nowrap px-4 py-3">余额</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3">默认代收(bps)</th>
              <th class="whitespace-nowrap px-4 py-3">默认代付(bps)</th>
              <th class="whitespace-nowrap px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="px-4 py-8 text-center text-slate-500">加载中...</td>
            </tr>
            <tr v-else-if="!filteredMerchants.length">
              <td colspan="6" class="px-4 py-8 text-center text-slate-500">暂无数据</td>
            </tr>
            <tr
              v-for="m in pagedMerchants"
              v-else
              :key="m.merchant_id"
              class="border-b border-slate-100 transition hover:bg-slate-50/80"
            >
              <td class="px-4 py-3 font-mono font-semibold text-slate-900">{{ m.merchant_id }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ formatMoney(m.balance) }}</td>
              <td class="px-4 py-3">
                <span
                  v-if="m.status === 1"
                  class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
                >
                  启用
                </span>
                <span v-else class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700">锁定</span>
              </td>
              <td class="px-4 py-3 tabular-nums text-slate-600">{{ m.default_collect_rate_bps }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-600">{{ m.default_payout_rate_bps }}</td>
              <td class="px-4 py-3 text-right">
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300"
                  @click="openEdit(m.merchant_id)"
                >
                  编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <AdminPaginationBar
        v-if="!loading && filteredMerchants.length"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :page-count="pageCount"
        @update:page="page = $event"
        @update:page-size="pageSize = $event"
      />
    </div>

    <AdminDrawer
      v-model="drawerOpen"
      :title="drawerTitle"
      subtitle="保存后生效；代收/代付产品授权在下方分栏配置。"
      max-width-class="max-w-3xl"
    >
      <div v-if="drawerOpen" class="space-y-4">
        <div class="flex flex-wrap border-b border-slate-200" role="tablist">
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
            :is-new="isNew"
            embedded
            hide-footer-actions
            :saving="saving"
            :saved="saved"
            :error="formError"
            :can-save="canSaveForm"
            :status-for-lock="form.status"
            @save="saveForm"
            @reset="resetForm"
            @toggle-lock="toggleLock"
            @reset-secret="resetSecret"
          />
        </div>

        <div v-show="rightTab === 'bindings_collect'" role="tabpanel">
          <MerchantPayProductsCard
            v-if="!isNew && selectedMerchant"
            embedded
            :product-ids="selectedMerchant.pay_product_ids || []"
            :catalog="payProducts"
            :loading="loadingProducts"
            :saving="bindingSaving"
            :bind-error="bindError"
            @remove="removePayProduct"
            @add="addPayProduct"
          />
          <p v-else class="rounded-lg border border-dashed border-slate-200 px-4 py-6 text-center text-sm text-slate-500">
            请先保存商户基本信息后再配置代收产品。
          </p>
        </div>

        <div v-show="rightTab === 'bindings_payout'" role="tabpanel">
          <MerchantPayoutProductsCard
            v-if="!isNew && selectedMerchant"
            embedded
            :product-ids="selectedMerchant.payout_product_ids || []"
            :catalog="payoutProducts"
            :loading="loadingPayoutProducts"
            :saving="bindingSaving"
            :bind-error="bindError"
            @remove="removePayoutProduct"
            @add="addPayoutProduct"
          />
          <p v-else class="rounded-lg border border-dashed border-slate-200 px-4 py-6 text-center text-sm text-slate-500">
            请先保存商户基本信息后再配置代付产品。
          </p>
        </div>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-start gap-3">
          <template v-if="rightTab === 'basic'">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
              @click="resetForm"
            >
              重置
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
              :disabled="saving || !canSaveForm"
              @click="saveForm"
            >
              {{ saving ? '保存中...' : '保存配置' }}
            </button>
          </template>
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
            @click="closeDrawer"
          >
            关闭
          </button>
        </div>
      </template>
    </AdminDrawer>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'

import AdminDrawer from '../../../components/AdminDrawer.vue'
import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { useClientPagination } from '../../../composables/useClientPagination'
import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import MerchantFormCard from './MerchantFormCard.vue'
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
const drawerOpen = ref(false)
const searchQuery = ref('')

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

const drawerTitle = computed(() =>
  isNew.value ? '新建商户' : `编辑商户 · ${form.value.merchant_id || ''}`,
)

const canSaveForm = computed(() => {
  if (isNew.value) return !!form.value.merchant_id?.trim()
  return true
})

const filteredMerchants = computed(() => {
  const s = searchQuery.value.trim().toLowerCase()
  const list = merchants.value
  if (!s) return list
  return list.filter((m) => (m.merchant_id || '').toLowerCase().includes(s))
})

const { page, pageSize, total, pageCount, slice: pagedMerchants } = useClientPagination(filteredMerchants, 10)

watch(searchQuery, () => {
  page.value = 1
})

function formatMoney(v: number) {
  return `¥ ${(v / 100).toFixed(2)}`
}

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

function openEdit(merchantId: string) {
  selectedMerchantId.value = merchantId
  rightTab.value = 'basic'
  applySelectedToForm()
  saved.value = false
  formError.value = ''
  drawerOpen.value = true
}

function openNew() {
  selectedMerchantId.value = null
  rightTab.value = 'basic'
  form.value = emptyMerchantForm()
  saved.value = false
  formError.value = ''
  drawerOpen.value = true
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

function closeDrawer() {
  drawerOpen.value = false
}

watch(drawerOpen, (open, wasOpen) => {
  if (wasOpen === true && open === false) void reload()
})

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
