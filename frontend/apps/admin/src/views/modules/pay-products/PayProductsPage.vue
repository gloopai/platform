<template>
  <div class="grid gap-4">
    <PayProductsHeader :title="headerTitle" :subtitle="headerSubtitle" @new-product="openNew" @refresh="reloadAll" />

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
        <input
          v-model.trim="searchQuery"
          type="search"
          autocomplete="off"
          placeholder="搜索编码、名称、ID…"
          class="w-full max-w-md rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
        />
        <label class="flex items-center gap-2 text-sm text-slate-600">
          <span class="text-slate-500">匹配</span>
          <span class="font-mono text-slate-900">{{ filteredProducts.length }}</span>
          <span class="text-slate-500">条</span>
        </label>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">ID</th>
              <th class="whitespace-nowrap px-4 py-3">编码</th>
              <th class="whitespace-nowrap px-4 py-3">名称</th>
              <th class="whitespace-nowrap px-4 py-3">排序</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="whitespace-nowrap px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingProducts">
              <td colspan="6" class="px-4 py-8 text-center text-slate-500">加载中...</td>
            </tr>
            <tr v-else-if="!filteredProducts.length">
              <td colspan="6" class="px-4 py-8 text-center text-slate-500">暂无数据</td>
            </tr>
            <tr
              v-for="p in pagedProducts"
              v-else
              :key="p.id"
              class="border-b border-slate-100 transition hover:bg-slate-50/80"
            >
              <td class="px-4 py-3 font-mono text-slate-800">#{{ p.id }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ p.code }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ p.name }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-600">{{ p.sort_order }}</td>
              <td class="px-4 py-3">
                <span
                  v-if="p.enabled"
                  class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
                >
                  启用
                </span>
                <span v-else class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">停用</span>
              </td>
              <td class="px-4 py-3 text-right">
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300"
                  @click="openEdit(p.id)"
                >
                  编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <AdminPaginationBar
        v-if="!loadingProducts && filteredProducts.length"
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
      :subtitle="headerSubtitle"
      max-width-class="max-w-3xl"
    >
      <div v-if="drawerOpen" class="space-y-4">
        <PayProductFormCard
          :model="form"
          embedded
          hide-footer-actions
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
          embedded
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

      <template #footer>
        <div class="flex flex-wrap items-center justify-start gap-3">
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
            @click="resetProductForm"
          >
            重置
          </button>
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="savingProduct || !adminTokenValue || !form.code || !form.name"
            @click="saveProduct"
          >
            {{ savingProduct ? '保存中...' : '保存产品' }}
          </button>
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
import { adminDelete, adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import PayProductBindingsCard from './PayProductBindingsCard.vue'
import PayProductFormCard from './PayProductFormCard.vue'
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
const drawerOpen = ref(false)
const searchQuery = ref('')

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
  props.payoutMode ? '代付产品与上游通道' : '代收产品与上游通道',
)
const headerSubtitle = computed(() =>
  props.payoutMode
    ? '维护代付对外产品编码与路由绑定；仅 supports_payout 的通道可参与。费率在通道与商户侧配置。'
    : '维护对外展示编码（code）、排序与启用；为每个产品绑定上游通道及权重。费率在通道与商户侧配置，不在产品绑定行设置。',
)

const drawerTitle = computed(() =>
  selectedProductId.value == null
    ? props.payoutMode
      ? '新建代付产品'
      : '新建代收产品'
    : props.payoutMode
      ? `编辑代付产品 · ${form.value.code || form.value.id}`
      : `编辑代收产品 · ${form.value.code || form.value.id}`,
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

const newBind = ref<{ channel_id: number; weight: number; enabled: boolean }>({
  channel_id: 0,
  weight: 100,
  enabled: true,
})

const boundChannelIds = computed(() => bindings.value.map((b) => b.channel_id))

const filteredProducts = computed(() => {
  const list = products.value
  const s = searchQuery.value.trim().toLowerCase()
  if (!s) return list
  return list.filter((p) => {
    const idStr = String(p.id)
    return (
      idStr.includes(s) ||
      (p.code || '').toLowerCase().includes(s) ||
      (p.name || '').toLowerCase().includes(s)
    )
  })
})

const { page, pageSize, total, pageCount, slice: pagedProducts } = useClientPagination(filteredProducts, 10)

watch(searchQuery, () => {
  page.value = 1
})

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
    const data = await adminGet<{ products: PayProduct[] }>(apiProducts.value)
    products.value = data.products || []
    if (selectedProductId.value && products.value.some((p) => p.id === selectedProductId.value)) {
      applySelectedProduct()
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

function openEdit(id: number) {
  selectProduct(id)
  drawerOpen.value = true
}

function newProduct() {
  selectedProductId.value = null
  form.value = emptyForm()
  bindings.value = []
  savedProduct.value = false
  productError.value = ''
  newBind.value = { channel_id: 0, weight: 100, enabled: true }
}

function openNew() {
  newProduct()
  drawerOpen.value = true
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

async function addBinding() {
  if (!form.value.id || newBind.value.channel_id <= 0) return
  addingBinding.value = true
  bindingError.value = ''
  try {
    await adminPost(apiBindings(form.value.id), {
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
    await adminPut(apiBindingRow(bindingId), { weight, enabled })
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

function closeDrawer() {
  drawerOpen.value = false
}

watch(drawerOpen, (open, wasOpen) => {
  if (wasOpen === true && open === false) void reloadAll()
})

let unregister: (() => void) | null = null
onMounted(() => {
  void reloadAll()
  if (registerRefresh) unregister = registerRefresh(() => void reloadAll())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
