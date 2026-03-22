<template>
  <div class="grid gap-4">
    <PayProductsHeader @new-product="newProduct" @refresh="reloadAll" />

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
          :channels="channels"
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

const emptyForm = (): PayProduct => ({
  id: 0,
  code: '',
  name: '',
  sort_order: 0,
  enabled: true,
})

const form = ref<PayProduct>(emptyForm())
const bindings = ref<PayProductBinding[]>([])

const newBind = ref({ channel_id: 0, weight: 100, enabled: true })

const boundChannelIds = computed(() => bindings.value.map((b) => b.channel_id))

function setForm(v: PayProduct) {
  form.value = v
}

function setNewBindDraft(v: { channel_id: number; weight: number; enabled: boolean }) {
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
    const data = await adminGet<{ products: PayProduct[] }>('/v1/admin/pay_products')
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
    const data = await adminGet<{ bindings: PayProductBinding[] }>(`/v1/admin/pay_products/${productId}/bindings`)
    bindings.value = data.bindings || []
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
  newBind.value = { channel_id: 0, weight: 100, enabled: true }
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
    const url = isUpdate ? `/v1/admin/pay_products/${form.value.id}` : '/v1/admin/pay_products'
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

async function addBinding() {
  if (!form.value.id || newBind.value.channel_id <= 0) return
  addingBinding.value = true
  bindingError.value = ''
  try {
    await adminPost(`/v1/admin/pay_products/${form.value.id}/bindings`, {
      channel_id: newBind.value.channel_id,
      weight: newBind.value.weight,
      enabled: newBind.value.enabled,
    })
    newBind.value = { channel_id: 0, weight: 100, enabled: true }
    await loadBindings(form.value.id)
  } catch {
    bindingError.value = '添加失败（通道不存在或已绑定）'
  } finally {
    addingBinding.value = false
  }
}

function onSaveBindingRow(payload: { id: number; weight: number; enabled: boolean }) {
  void updateBinding(payload.id, payload.weight, payload.enabled)
}

async function updateBinding(bindingId: number, weight: number, enabled: boolean) {
  bindingError.value = ''
  if (weight == null || weight < 1) {
    bindingError.value = '权重须为正整数'
    return
  }
  try {
    await adminPut(`/v1/admin/pay_product_bindings/${bindingId}`, { weight, enabled })
    await loadBindings(form.value.id)
  } catch {
    bindingError.value = '更新绑定失败'
  }
}

async function removeBinding(bindingId: number) {
  if (!confirm('确定删除该通道绑定？')) return
  bindingError.value = ''
  try {
    await adminDelete(`/v1/admin/pay_product_bindings/${bindingId}`)
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
