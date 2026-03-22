<template>
  <div class="grid gap-4">
    <PayProductsHeader :title="headerTitle" :subtitle="headerSubtitle" @new-product="newProduct" @refresh="reloadAll" />

    <div class="grid grid-cols-12 gap-4">
      <PayProductList
        :products="products"
        :loading="loadingProducts"
        :selected-id="selectedProductId"
        @select="selectProduct"
      />

      <div class="col-span-12 space-y-4 md:col-span-8">
        <PayProductFormCard
          :model="form"
          :saving="savingProduct"
          :saved="savedProduct"
          :error="productError"
          :can-save="!!adminTokenValue && !!form.code && !!form.name"
          @update:model="setForm"
          @save="saveProduct"
          @reset="resetProductForm"
        />

        <PayProductBindingsCard
          v-if="form.id"
          :bindings="bindings"
          :channels="filteredChannels"
          :exclude-channel-ids="boundChannelIds"
          :loading="loadingBindings"
          :error="bindingError"
          :adding="addingBinding"
          :draft="newBind"
          @update:draft="setNewBindDraft"
          @save-row="onSaveBindingRow"
          @delete-row="removeBinding"
          @add="addBinding"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'

import { adminDelete, adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import PayProductBindingsCard from './PayProductBindingsCard.vue'
import PayProductFormCard from './PayProductFormCard.vue'
import PayProductList from './PayProductList.vue'
import PayProductsHeader from './PayProductsHeader.vue'
import type { PayProduct, PayProductBinding, PayProductChannelOption } from './types'

const props = withDefaults(
  defineProps<{
    /** true = 代付产品与路由配置（独立 API） */
    payoutMode?: boolean
  }>(),
  { payoutMode: false },
)

const adminToken = inject('adminToken') as { value: string } | undefined
const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const adminTokenValue = computed(() => adminToken?.value || '')

const loadingProducts = ref(false)
const loadingBindings = ref(false)
const savingProduct = ref(false)
const addingBinding = ref(false)
const savedProduct = ref(false)
const productError = ref('')
const bindingError = ref('')

const products = ref<PayProduct[]>([])
const channels = ref<PayProductChannelOption[]>([])
const selectedProductId = ref<number | null>(null)

const apiProducts = computed(() => (props.payoutMode ? '/v1/admin/payout_products' : '/v1/admin/pay_products'))
const apiBindings = (productId: number) =>
  props.payoutMode
    ? `/v1/admin/payout_products/${productId}/bindings`
    : `/v1/admin/pay_products/${productId}/bindings`
const apiBindingRow = (bindingId: number) =>
  props.payoutMode
    ? `/v1/admin/payout_product_bindings/${bindingId}`
    : `/v1/admin/pay_product_bindings/${bindingId}`

const headerTitle = computed(() =>
  props.payoutMode ? '代付产品与上游通道' : '支付产品与上游通道',
)
const headerSubtitle = computed(() =>
  props.payoutMode
    ? '维护代付对外产品编码与绑定；仅 supports_payout 的通道可参与代付绑定。'
    : '维护对外展示编码（code）、排序与启用；为每个产品绑定多条上游通道及权重。绑定行可覆盖上游成本（万分比）。',
)

const filteredChannels = computed(() => {
  if (props.payoutMode) {
    return channels.value.filter((c) => c.supports_payout !== false)
  }
  return channels.value.filter((c) => c.supports_collect !== false)
})

const emptyForm = (): PayProduct => ({
  id: 0,
  code: '',
  name: '',
  sort_order: 0,
  enabled: true,
})

const form = ref<PayProduct>(emptyForm())
const bindings = ref<PayProductBinding[]>([])

const newBind = ref<{ channel_id: number; weight: number; enabled: boolean; cost_rate_bps?: number | null }>({
  channel_id: 0,
  weight: 100,
  enabled: true,
  cost_rate_bps: null,
})

const boundChannelIds = computed(() => bindings.value.map((b) => b.channel_id))

function setForm(v: PayProduct) {
  form.value = v
}

function setNewBindDraft(v: {
  channel_id: number
  weight: number
  enabled: boolean
  cost_rate_bps?: number | null
}) {
  newBind.value = v
}

async function loadChannels() {
  try {
    const data = await adminGet<{ channels: PayProductChannelOption[] }>('/v1/admin/channels')
    channels.value = data.channels || []
  } catch {
    channels.value = []
  }
}

async function loadProducts() {
  loadingProducts.value = true
  productError.value = ''
  try {
    const data = await adminGet<{ products: PayProduct[] }>(apiProducts.value)
    products.value = data.products || []
    if (selectedProductId.value && products.value.some((p) => p.id === selectedProductId.value)) {
      applySelectedProduct()
    } else if (products.value.length > 0) {
      selectProduct(products.value[0].id)
    } else {
      newProduct()
    }
  } catch {
    productError.value = '加载产品列表失败'
    products.value = []
  } finally {
    loadingProducts.value = false
  }
}

async function loadBindings(productId: number) {
  loadingBindings.value = true
  bindingError.value = ''
  try {
    const data = await adminGet<{ bindings: PayProductBinding[] }>(apiBindings(productId))
    const raw = data.bindings || []
    bindings.value = props.payoutMode
      ? raw.map((b) => ({ ...b, pay_product_id: b.payout_product_id }))
      : raw
  } catch {
    bindings.value = []
    bindingError.value = '加载绑定失败'
  } finally {
    loadingBindings.value = false
  }
}

function applySelectedProduct() {
  const p = products.value.find((x) => x.id === selectedProductId.value)
  if (!p) return
  form.value = { ...p }
}

function selectProduct(id: number) {
  selectedProductId.value = id
  applySelectedProduct()
  savedProduct.value = false
  productError.value = ''
  void loadBindings(id)
}

function newProduct() {
  selectedProductId.value = null
  form.value = emptyForm()
  bindings.value = []
  savedProduct.value = false
  productError.value = ''
  newBind.value = { channel_id: 0, weight: 100, enabled: true, cost_rate_bps: null }
}

function resetProductForm() {
  if (selectedProductId.value) applySelectedProduct()
  else form.value = emptyForm()
  savedProduct.value = false
  productError.value = ''
}

async function saveProduct() {
  savingProduct.value = true
  productError.value = ''
  savedProduct.value = false
  try {
    const body = {
      code: form.value.code,
      name: form.value.name,
      sort_order: form.value.sort_order,
      enabled: form.value.enabled,
    }
    const isUpdate = form.value.id > 0
    const url = isUpdate ? `${apiProducts.value}/${form.value.id}` : apiProducts.value
    const data = isUpdate
      ? await adminPut<{ product: PayProduct }>(url, body)
      : await adminPost<{ product: PayProduct }>(url, body)
    const p = data.product
    const idx = products.value.findIndex((x) => x.id === p.id)
    if (idx >= 0) products.value[idx] = p
    else products.value.unshift(p)
    selectedProductId.value = p.id
    form.value = { ...p }
    savedProduct.value = true
    await loadBindings(p.id)
  } catch {
    productError.value = '保存失败（编码重复或网络错误）'
  } finally {
    savingProduct.value = false
  }
}

function bodyWithCost(cost: number | null | undefined) {
  const o: Record<string, unknown> = {}
  if (cost === undefined) return o
  o.cost_rate_bps = cost
  return o
}

async function addBinding() {
  if (!form.value.id || newBind.value.channel_id <= 0) return
  addingBinding.value = true
  bindingError.value = ''
  try {
    const payload: Record<string, unknown> = {
      channel_id: newBind.value.channel_id,
      weight: newBind.value.weight,
      enabled: newBind.value.enabled,
      ...bodyWithCost(newBind.value.cost_rate_bps === null ? null : newBind.value.cost_rate_bps),
    }
    await adminPost(apiBindings(form.value.id), payload)
    newBind.value = { channel_id: 0, weight: 100, enabled: true, cost_rate_bps: null }
    await loadBindings(form.value.id)
  } catch {
    bindingError.value = '添加失败（通道不存在或已绑定）'
  } finally {
    addingBinding.value = false
  }
}

function onSaveBindingRow(payload: {
  id: number
  weight: number
  enabled: boolean
  cost_rate_bps?: number | null
}) {
  void updateBinding(payload.id, payload.weight, payload.enabled, payload.cost_rate_bps)
}

async function updateBinding(
  bindingId: number,
  weight: number,
  enabled: boolean,
  cost_rate_bps?: number | null,
) {
  bindingError.value = ''
  if (weight == null || weight < 1) {
    bindingError.value = '权重须为正整数'
    return
  }
  try {
    const body: Record<string, unknown> = { weight, enabled, ...bodyWithCost(cost_rate_bps) }
    await adminPut(apiBindingRow(bindingId), body)
    await loadBindings(form.value.id)
  } catch {
    bindingError.value = '更新绑定失败'
  }
}

async function removeBinding(bindingId: number) {
  if (!confirm('确定删除该通道绑定？')) return
  bindingError.value = ''
  try {
    await adminDelete(apiBindingRow(bindingId))
    await loadBindings(form.value.id)
  } catch {
    bindingError.value = '删除失败'
  }
}

async function reloadAll() {
  savedProduct.value = false
  await loadChannels()
  await loadProducts()
}

let unregister: (() => void) | null = null
onMounted(() => {
  void reloadAll()
  if (registerRefresh) unregister = registerRefresh(() => void reloadAll())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
