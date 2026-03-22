<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex items-start justify-between gap-3">
        <div>
          <div class="text-sm font-semibold text-slate-900">支付产品与上游通道</div>
          <div class="mt-1 text-sm text-slate-600">
            维护对外展示编码（code）、排序与启用；为每个产品绑定多条上游通道及权重，供交易路由加权选用。
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white" @click="newProduct">新建产品</button>
          <button
            class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700"
            @click="reloadAll"
          >
            刷新
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-12 gap-4">
      <div class="col-span-12 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-4">
        <div class="text-xs font-semibold text-slate-500">支付产品</div>
        <div v-if="loadingProducts" class="mt-3 text-sm text-slate-500">加载中...</div>
        <div v-else class="mt-3 space-y-2">
          <button
            v-for="p in products"
            :key="p.id"
            class="w-full rounded-xl border px-3 py-3 text-left hover:bg-slate-50"
            :class="selectedProductId === p.id ? 'border-slate-900' : 'border-slate-200'"
            @click="selectProduct(p.id)"
          >
            <div class="flex items-start justify-between gap-2">
              <div>
                <div class="text-sm font-semibold text-slate-900">{{ p.name }}</div>
                <div class="mt-1 font-mono text-xs text-slate-500">{{ p.code }}</div>
              </div>
              <span
                v-if="p.enabled"
                class="shrink-0 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
              >
                启用
              </span>
              <span v-else class="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">
                停用
              </span>
            </div>
          </button>
        </div>
      </div>

      <div class="col-span-12 space-y-4 md:col-span-8">
        <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div class="text-xs text-slate-500">产品配置：{{ form.id ? `#${form.id}` : '新建' }}</div>
            <div v-if="savedProduct" class="text-xs font-semibold text-emerald-700">已保存</div>
          </div>
          <div class="mt-4 grid grid-cols-12 gap-4">
            <label class="col-span-12 grid gap-1 md:col-span-4">
              <span class="text-xs font-medium text-slate-600">编码 code</span>
              <input v-model.trim="form.code" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-sm" />
            </label>
            <label class="col-span-12 grid gap-1 md:col-span-5">
              <span class="text-xs font-medium text-slate-600">展示名称</span>
              <input v-model.trim="form.name" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="col-span-12 grid gap-1 md:col-span-3">
              <span class="text-xs font-medium text-slate-600">排序</span>
              <input v-model.number="form.sort_order" type="number" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-12">
              <div class="text-sm text-slate-700">启用该产品</div>
              <input v-model="form.enabled" type="checkbox" class="h-4 w-4" />
            </label>
          </div>
          <div v-if="productError" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
            {{ productError }}
          </div>
          <div class="mt-6 flex flex-wrap gap-3">
            <button
              class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
              :disabled="savingProduct || !adminTokenValue || !form.code || !form.name"
              @click="saveProduct"
            >
              {{ savingProduct ? '保存中...' : '保存产品' }}
            </button>
            <button
              class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
              @click="resetProductForm"
            >
              重置
            </button>
          </div>
        </div>

        <div v-if="form.id" class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div class="text-sm font-semibold text-slate-900">上游通道绑定</div>
          <p class="mt-1 text-xs text-slate-500">同产品下多条通道按权重加权随机；权重为相对值，建议与通道限额配合。</p>

          <div v-if="loadingBindings" class="mt-4 text-sm text-slate-500">加载绑定...</div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="min-w-full text-left text-sm">
              <thead>
                <tr class="border-b border-slate-200 text-xs text-slate-500">
                  <th class="py-2 pr-3">通道</th>
                  <th class="py-2 pr-3">权重</th>
                  <th class="py-2 pr-3">启用</th>
                  <th class="py-2">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="b in bindings" :key="b.id" class="border-b border-slate-100">
                  <td class="py-2 pr-3">
                    <div class="font-medium text-slate-900">#{{ b.channel_id }} {{ b.channel_name || '-' }}</div>
                  </td>
                  <td class="py-2 pr-3">
                    <input
                      v-model.number="editWeight[b.id]"
                      type="number"
                      min="1"
                      class="w-24 rounded border border-slate-200 px-2 py-1 text-sm"
                    />
                  </td>
                  <td class="py-2 pr-3">
                    <input v-model="editEnabled[b.id]" type="checkbox" class="h-4 w-4" />
                  </td>
                  <td class="py-2">
                    <button
                      class="mr-2 text-xs font-semibold text-slate-700 underline"
                      @click="updateBinding(b.id)"
                    >
                      保存
                    </button>
                    <button class="text-xs font-semibold text-rose-700 underline" @click="removeBinding(b.id)">删除</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="mt-6 rounded-xl border border-dashed border-slate-200 p-4">
            <div class="text-xs font-semibold text-slate-600">新增绑定</div>
            <div class="mt-3 flex flex-wrap items-end gap-3">
              <label class="grid gap-1">
                <span class="text-xs text-slate-500">通道</span>
                <select v-model.number="newBind.channel_id" class="min-w-[200px] rounded-md border border-slate-200 px-3 py-2 text-sm">
                  <option :value="0">选择通道</option>
                  <option v-for="c in channels" :key="c.id" :value="c.id">
                    #{{ c.id }} {{ c.name }} ({{ c.pay_type || '-' }})
                  </option>
                </select>
              </label>
              <label class="grid gap-1">
                <span class="text-xs text-slate-500">权重</span>
                <input v-model.number="newBind.weight" type="number" min="1" class="w-24 rounded-md border border-slate-200 px-3 py-2 text-sm" />
              </label>
              <label class="flex items-center gap-2 pb-2">
                <input v-model="newBind.enabled" type="checkbox" class="h-4 w-4" />
                <span class="text-sm text-slate-700">启用</span>
              </label>
              <button
                class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
                :disabled="addingBinding || newBind.channel_id <= 0"
                @click="addBinding"
              >
                {{ addingBinding ? '提交...' : '添加' }}
              </button>
            </div>
            <div v-if="bindingError" class="mt-3 text-sm text-rose-700">{{ bindingError }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, reactive, ref } from 'vue'

import { adminDelete, adminGet, adminPost, adminPut } from '../../lib/adminApi'

type PayProduct = {
  id: number
  code: string
  name: string
  sort_order: number
  enabled: boolean
}

type Binding = {
  id: number
  pay_product_id: number
  channel_id: number
  channel_name: string
  weight: number
  enabled: boolean
}

type Channel = {
  id: number
  name: string
  pay_type: string
}

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
const channels = ref<Channel[]>([])
const selectedProductId = ref<number | null>(null)

const emptyForm = (): PayProduct => ({
  id: 0,
  code: '',
  name: '',
  sort_order: 0,
  enabled: true,
})

const form = ref<PayProduct>(emptyForm())
const bindings = ref<Binding[]>([])
const editWeight = reactive<Record<number, number>>({})
const editEnabled = reactive<Record<number, boolean>>({})

const newBind = ref({ channel_id: 0, weight: 100, enabled: true })

function syncBindingEditors(b: Binding[]) {
  Object.keys(editWeight).forEach((k) => delete editWeight[Number(k)])
  Object.keys(editEnabled).forEach((k) => delete editEnabled[Number(k)])
  for (const row of b) {
    editWeight[row.id] = row.weight
    editEnabled[row.id] = row.enabled
  }
}

async function loadChannels() {
  try {
    const data = await adminGet<{ channels: Channel[] }>('/v1/admin/channels')
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
    const data = await adminGet<{ bindings: Binding[] }>(`/v1/admin/pay_products/${productId}/bindings`)
    bindings.value = data.bindings || []
    syncBindingEditors(bindings.value)
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
  syncBindingEditors([])
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

async function updateBinding(bindingId: number) {
  bindingError.value = ''
  const w = editWeight[bindingId]
  const en = editEnabled[bindingId]
  if (w == null || w < 1) {
    bindingError.value = '权重须为正整数'
    return
  }
  try {
    await adminPut(`/v1/admin/pay_product_bindings/${bindingId}`, { weight: w, enabled: en })
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
